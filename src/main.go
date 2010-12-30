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
	"image/png"
	"runtime"
	"syscall"
)

import (
	"buildversion"
	"goray/logging"
	"goray/time"
	"goray/version"
	"goray/core/integrator"
	"goray/core/render"
	"goray/core/scene"
	"goray/std/yamlscene"
)

func printInstructions() {
	fmt.Println("USAGE: goray [OPTION]... FILE")
	fmt.Println("OPTIONS:")
	flag.PrintDefaults()
}

func printVersion() {
	fmt.Printf("goray v%s - The Concurrent Raytracer\n", version.GetString())
	// Copyright notice
	fmt.Println("Copyright © 2005 Mathias Wein, Alejandro Conty, and Alfredo de Greef")
	fmt.Println("Copyright © 2006 Kirill Simonov")
	fmt.Println("Copyright © 2010 Ross Light")
	fmt.Println()
	fmt.Println("Based on the excellent YafaRay Ray-Tracer by Mathias Wein, Alejandro Conty, and")
	fmt.Println("Alfredo de Greef.")
	fmt.Println("Parts of the YAML parser are ported from libyaml (written by Kirill Simonov).")
	fmt.Println("Go rewrite by Ross Light in 2010.")
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

func main() {
	var err os.Error

	showHelp := flag.Bool("help", false, "display this help")
	format := flag.String("f", "png", "the output format")
	outputPath := flag.String("o", "goray.png", "path for the output file")
	//width := flag.Int("w", 100, "the output width")
	//height := flag.Int("h", 100, "the output height")
	debug := flag.Int("d", 0, "set debug verbosity level")
	showVersion := flag.Bool("version", false, "display the version")
	maxProcs := flag.Int("procs", 1, "set the number of processors to use")

	flag.Usage = printInstructions
	flag.Parse()

	switch {
	case *showHelp:
		printInstructions()
		return
	case *showVersion:
		printVersion()
		return
	}

	// Open input file
	if flag.NArg() != 1 {
		printInstructions()
		return
	}
	inFile, err := os.Open(flag.Arg(0), os.O_RDONLY, 0)
	if err != nil {
		logging.MainLog.Critical("Error opening input file: %v", err)
		return
	}
	defer inFile.Close()

	// Set up logging
	{
		level := logging.Level(logging.InfoLevel - 10*(*debug))
		writeHandler := logging.NewWriterHandler(os.Stdout)
		logging.MainLog.AddHandler(logging.NewMinLevelFilter(writeHandler, level))
	}
	defer logging.MainLog.Close()

	// Change the number of processors to use
	runtime.GOMAXPROCS(*maxProcs)
	logging.MainLog.Debug("Using %d processor(s)", runtime.GOMAXPROCS(0))

	// Open output file
	// Normally, we want to truncate upon open, but we do that right before
	// writing data so that we preserve data for as long as possible.
	outFile, err := os.Open(*outputPath, os.O_WRONLY|os.O_CREAT, 0644)
	if err != nil {
		logging.MainLog.Critical("Error opening output file: %v", err)
		return
	}
	defer outFile.Close()

	// Set up scene
	sc := scene.New()

	logging.MainLog.Info("Setting up scene...")
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
		return
	}

	// Update scene (build tree structures and whatnot)
	logging.MainLog.Info("Finalizing scene...")
	finalizeTime := time.Stopwatch(func() {
		sc.Update()
	})
	logging.MainLog.Info("Finalized in %v", finalizeTime)

	logging.MainLog.Info("Rendering...")

	var outputImage *render.Image
	renderTime := time.Stopwatch(func() {
		renderLog := logging.NewFormatFilter(
			logging.MainLog,
			func(rec logging.Record) string {
				return "  RENDER: " + rec.String()
			},
		)
		outputImage = integrator.Render(sc, integ, renderLog)
	})
	if err != nil {
		logging.MainLog.Error("Rendering error: %v", err)
		return
	}
	logging.MainLog.Info("Render finished in %v", renderTime)

	logging.MainLog.Info("TOTAL TIME: %v", finalizeTime+renderTime)

	logging.MainLog.Info("Writing and finishing...")

	outFile.Truncate(0)
	switch *format {
	case "png":
		err = png.Encode(outFile, outputImage)
	}

	if err != nil {
		logging.MainLog.Critical("Error while writing: %v", err)
		return
	}
}
