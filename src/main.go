//
//  main.go
//  goray
//
//  Created by Ross Light on 2010-05-27.
//

package main

import (
	"flag"
	"fmt"
	"os"
	"http"
	"runtime"
	"syscall"

	"buildversion"

	"goray/job"
	"goray/logging"
	"goray/server"
	"goray/version"

	_ "goray/std/all"
	"goray/std/yamlscene"
	"goray/std/textures/image/fileloader"
)

var showHelp, showVersion bool
var httpAddress string
var outputPath, imagePath string
var debug int

func main() {
	flag.BoolVar(&showHelp, "help", false, "display this help")
	flag.BoolVar(&showVersion, "version", false, "display the version")
	flag.StringVar(&httpAddress, "http", "", "start HTTP server")
	flag.StringVar(&outputPath, "o", "", "path for the output")
	flag.IntVar(&debug, "d", 0, "set debug verbosity level")
	flag.StringVar(&imagePath, "t", ".", "texture directory (default: current directory)")
	maxProcs := flag.Int("procs", 1, "set the number of processors to use")

	flag.Usage = printInstructions
	flag.Parse()

	setupLogging()

	runtime.GOMAXPROCS(*maxProcs)
	logging.MainLog.Debug("Using %d processor(s)", runtime.GOMAXPROCS(0))

	var exitCode int
	switch {
	case showHelp:
		printInstructions()
	case showVersion:
		printVersion()
	case httpAddress != "":
		exitCode = httpServer()
	default:
		exitCode = singleFile()
	}
	os.Exit(exitCode)
}

func setupLogging() {
	level := logging.Level(logging.InfoLevel - 10*debug)
	writeHandler := logging.NewWriterHandler(os.Stdout)
	logging.MainLog.AddHandler(logging.NewMinLevelFilter(writeHandler, level))
}

func printInstructions() {
	fmt.Println("USAGE: goray [OPTIONS] FILE")
	fmt.Println("       goray -http=:PORT [OPTIONS]")
	fmt.Println("OPTIONS:")
	flag.PrintDefaults()
}

func printVersion() {
	fmt.Printf("goray v%s - The Concurrent Raytracer\n", version.GetString())
	// Copyright notice
	fmt.Println("Copyright © 2011 Ross Light")
	fmt.Println("Based on YafaRay: Copyright © 2005 Mathias Wein, Alejandro Conty, and Alfredo")
	fmt.Println("de Greef")
	fmt.Println("jQuery: Copyright © 2011 John Resig")
	fmt.Println()
	fmt.Println("goray comes with ABSOLUTELY NO WARRANTY.  goray is free software, and you are")
	fmt.Println("welcome to redistribute it under the conditions of the GNU Lesser General")
	fmt.Println("Public License v3, or (at your option) any later version.")
	fmt.Println()
	// Build information
	if buildversion.Source == "bzr" {
		fmt.Printf("Built from \"%s\" branch\n", buildversion.BranchNickname)
		fmt.Printf("  Revision: %s [%s]\n", buildversion.RevNo, buildversion.RevID)
		if buildversion.CleanWC == 0 {
			fmt.Printf("  With local modifications\n")
		}
	} else {
		fmt.Println("Built from a source archive")
	}
	fmt.Printf("Go runtime: %s\n", runtime.Version())
	fmt.Printf("Built for %s (%s)\n", syscall.OS, syscall.ARCH)
}

func singleFile() int {
	if flag.NArg() != 1 {
		printInstructions()
		return 1
	}

	// Open input file
	inFile, err := os.Open(flag.Arg(0))
	if err != nil {
		logging.MainLog.Critical("Error opening input file: %v", err)
		return 1
	}
	defer inFile.Close()

	// Open output file
	if outputPath == "" {
		outputPath = "goray.png"
	}
	outFile, err := os.Create(outputPath)
	if err != nil {
		logging.MainLog.Critical("Error opening output file: %v", err)
		return 1
	}
	defer outFile.Close()

	// Create job
	j := job.New("job", inFile, yamlscene.Params{
		"ImageLoader": fileloader.New(imagePath),
	})
	ch := j.StatusChan()
	j.SceneLog = logging.NewFormatFilter(
		logging.MainLog,
		func(rec logging.Record) string {
			return "  SCENE: " + rec.String()
		},
	)
	j.RenderLog = logging.NewFormatFilter(
		logging.MainLog,
		func(rec logging.Record) string {
			return "  RENDER: " + rec.String()
		},
	)
	go j.Render(outFile)

	// Log progress
	for stat := range ch {
		switch stat.Code {
		case job.StatusReading:
			logging.MainLog.Info("Reading scene file...")
		case job.StatusUpdating:
			logging.MainLog.Info("Preparing scene...")
		case job.StatusRendering:
			logging.MainLog.Info("Finalized in %v", stat.UpdateTime)
			logging.MainLog.Info("Rendering...")
		case job.StatusWriting:
			logging.MainLog.Info("Render finished in %v", stat.RenderTime)
			logging.MainLog.Info("Writing...")
		case job.StatusDone:
			logging.MainLog.Info("TOTAL TIME: %v", stat.TotalTime())
		case job.StatusError:
			logging.MainLog.Critical("Error: %v", stat.Error)
			return 1
		}
	}
	return 0
}

func httpServer() int {
	if flag.NArg() != 0 {
		printInstructions()
		return 1
	}
	if outputPath == "" {
		outputPath = "output"
	}
	storage, err := job.NewFileStorage(outputPath)
	if err != nil {
		logging.MainLog.Critical("FileStorage: %v", err)
		return 1
	}

	s := server.New(job.NewManager(storage, 5), "data")
	s.BaseParams = yamlscene.Params{
		"ImageLoader": fileloader.New(imagePath),
	}
	logging.MainLog.Info("Starting HTTP server")
	err = http.ListenAndServe(httpAddress, s)
	if err != nil {
		logging.MainLog.Critical("ListenAndServe: %v", err)
		return 1
	}
	return 0
}
