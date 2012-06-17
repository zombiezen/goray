/*
	Copyright (c) 2011 Ross Light.
	Copyright (c) 2005 Mathias Wein, Alejandro Conty, and Alfredo de Greef.

	This file is part of goray.

	goray is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	goray is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with goray.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	"bitbucket.org/zombiezen/goray/job"
	"bitbucket.org/zombiezen/goray/logging"

	_ "bitbucket.org/zombiezen/goray/std"
	"bitbucket.org/zombiezen/goray/std/textures/image/fileloader"
	"bitbucket.org/zombiezen/goray/std/yamlscene"
)

var (
	showHelp     bool
	showVersion  bool
	httpAddress  string
	dataRoot     string
	outputPath   string
	outputFormat string
	imagePath    string
	cpuprofile   string
	debug        int
)

func main() {
	flag.BoolVar(&showHelp, "help", false, "display this help")
	flag.BoolVar(&showVersion, "version", false, "display the version")
	flag.StringVar(&httpAddress, "http", "", "start HTTP server")
	flag.StringVar(&dataRoot, "dataroot", "data", "web server resource files")
	flag.StringVar(&outputPath, "o", "", "path for the output")
	flag.StringVar(&outputFormat, "f", job.DefaultFormat, "output format (default: "+job.DefaultFormat+")")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write CPU profile to file")
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
	fmt.Println("goray v0.1.0 - The Concurrent Raytracer")

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
	fmt.Printf("Go runtime: %s\n", runtime.Version())
	fmt.Printf("Built for %s (%s)\n", runtime.GOOS, runtime.GOARCH)
}

func singleFile() int {
	if flag.NArg() != 1 {
		printInstructions()
		return 1
	}

	// Check output format
	formatStruct, found := job.FormatMap[outputFormat]
	if !found {
		logging.MainLog.Critical("Unrecognized output format: %s", outputFormat)
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
		outputPath = "goray" + formatStruct.Extension
	}
	outFile, err := os.Create(outputPath)
	if err != nil {
		logging.MainLog.Critical("Error opening output file: %v", err)
		return 1
	}
	defer outFile.Close()

	// Set up profile file
	var cpuprofileFile *os.File
	if cpuprofile != "" {
		cpuprofileFile, err = os.Create(cpuprofile)
		if err != nil {
			logging.MainLog.Critical("Error opening cpuprofile file: %v", err)
			return 1
		}
		defer cpuprofileFile.Close()
	}

	// Create job
	j := job.New("job", inFile, yamlscene.Params{
		"ImageLoader":  fileloader.New(imagePath),
		"OutputFormat": formatStruct,
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
	go func() {
		if cpuprofileFile != nil {
			pprof.StartCPUProfile(cpuprofileFile)
		}
		j.Render(outFile)
	}()

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
			pprof.StopCPUProfile()
		case job.StatusDone:
			logging.MainLog.Info("TOTAL TIME: %v", stat.TotalTime())
		case job.StatusError:
			logging.MainLog.Critical("Error: %v", stat.Error)
			return 1
		}
	}
	return 0
}