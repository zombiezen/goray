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
	"image/png"
	"runtime"
	"syscall"

	"buildversion"

	"goray/job"
	"goray/logging"
	"goray/time"
	"goray/server"
	"goray/version"

	"goray/core/integrator"
	"goray/core/render"
	"goray/core/scene"

	"goray/std/yamlscene"
)

var showHelp, showVersion bool
var httpAddress string
var outputPath string
var debug int

func printInstructions() {
	fmt.Println("USAGE: goray [OPTIONS] FILE")
	fmt.Println("       goray -http=:PORT [OPTIONS]")
	fmt.Println("OPTIONS:")
	flag.PrintDefaults()
}

func printVersion() {
	fmt.Printf("goray v%s - The Concurrent Raytracer\n", version.GetString())
	// Copyright notice
	fmt.Println("Copyright © 2005 Mathias Wein, Alejandro Conty, and Alfredo de Greef")
	fmt.Println("Copyright © 2011 Ross Light")
	fmt.Println("Copyright © 2011 John Resig")
	fmt.Println()
	fmt.Println("Based on the excellent YafaRay Ray-Tracer by Mathias Wein, Alejandro Conty, and")
	fmt.Println("Alfredo de Greef.")
	fmt.Println("Go rewrite by Ross Light in 2010.")
	fmt.Println("Web frontend uses jQuery by John Resig")
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

func garbage() {
	before := runtime.MemStats.Alloc
	runtime.GC()
	after := runtime.MemStats.Alloc
	logging.MainLog.VerboseDebug("  [GC] %d KiB", int64(before-after)/1024)
}

func logMemInfo() {
	logging.MainLog.Debug("  [MEM] %d KiB/%d KiB", runtime.MemStats.Alloc/1024, runtime.MemStats.Sys/1024)
}

func setupLogging() {
	level := logging.Level(logging.InfoLevel - 10*debug)
	writeHandler := logging.NewWriterHandler(os.Stdout)
	logging.MainLog.AddHandler(logging.NewMinLevelFilter(writeHandler, level))
}

func main() {
	flag.BoolVar(&showHelp, "help", false, "display this help")
	flag.BoolVar(&showVersion, "version", false, "display the version")
	flag.StringVar(&httpAddress, "http", "", "start HTTP server")
	flag.StringVar(&outputPath, "o", "", "path for the output")
	flag.IntVar(&debug, "d", 0, "set debug verbosity level")
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
	logging.MainLog.Close()
	os.Exit(exitCode)
}

func singleFile() int {
	if flag.NArg() != 1 {
		printInstructions()
		return 1
	}

	// Open input file
	inFile, err := os.Open(flag.Arg(0), os.O_RDONLY, 0)
	if err != nil {
		logging.MainLog.Critical("Error opening input file: %v", err)
		return 1
	}
	defer inFile.Close()

	// Open output file
	// Normally, we want to truncate upon open, but we do that right before
	// writing data so that we preserve data for as long as possible.
	if outputPath == "" {
		outputPath = "goray.png"
	}
	outFile, err := os.Open(outputPath, os.O_WRONLY|os.O_CREAT, 0644)
	if err != nil {
		logging.MainLog.Critical("Error opening output file: %v", err)
		return 1
	}
	defer outFile.Close()

	// Set up scene
	logging.MainLog.Info("Setting up scene...")
	sc := scene.New()
	sc.GetLog().AddHandler(logging.NewFormatFilter(
		logging.MainLog,
		func(rec logging.Record) string {
			return "  SCENE: " + rec.String()
		},
	))

	// Parse input file
	integ, err := yamlscene.Load(inFile, sc)
	if err != nil {
		logging.MainLog.Critical("Error parsing input file: %v", err)
		return 1
	}

	logMemInfo()
	garbage()

	// Update scene (build tree structures and whatnot)
	logging.MainLog.Info("Finalizing scene...")
	finalizeTime := time.Stopwatch(func() {
		sc.Update()
	})
	logging.MainLog.Info("Finalized in %v", finalizeTime)

	logMemInfo()
	garbage()

	// Render scene
	var outputImage *render.Image
	logging.MainLog.Info("Rendering...")
	renderTime := time.Stopwatch(func() {
		renderLog := logging.NewFormatFilter(
			logging.MainLog,
			func(rec logging.Record) string {
				return "  RENDER: " + rec.String()
			},
		)
		outputImage = integrator.Render(sc, integ, renderLog)
	})
	logging.MainLog.Info("Render finished in %v", renderTime)
	logging.MainLog.Info("TOTAL TIME: %v", finalizeTime+renderTime)

	logMemInfo()
	garbage()

	// Write file
	logging.MainLog.Info("Writing and finishing...")
	outFile.Truncate(0)
	err = png.Encode(outFile, outputImage)
	if err != nil {
		logging.MainLog.Critical("Error while writing: %v", err)
		return 1
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
	err = http.ListenAndServe(httpAddress, s)
	if err != nil {
		logging.MainLog.Critical("ListenAndServe: %v", err)
		return 1
	}
	return 0
}
