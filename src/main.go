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
    "time"
)

import (
    "./buildversion"
	"./goray/camera"
	"./goray/object"
	"./goray/primitive"
	"./goray/scene"
	"./goray/vector"
    "./goray/version"
	trivialInt "./goray/std/integrators/trivial"
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
    fmt.Println("Copyright © 2010 Ross Light")
    fmt.Println()
    fmt.Println("Based on the excellent YafaRay Ray-Tracer by Mathias Wein, Alejandro Conty, and")
    fmt.Println("Alfredo de Greef.")
    fmt.Println("Go rewrite by Ross Light")
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
}

func main() {
	var err os.Error

	showHelp := flag.Bool("help", false, "display this help")
	format := flag.String("f", "png", "the output format")
	outputPath := flag.String("o", "goray.png", "path for the output file")
    width := flag.Int("w", 100, "the output width")
    height := flag.Int("h", 100, "the output height")
	debug := flag.Int("d", 0, "set debug verbosity level")
	showVersion := flag.Bool("v", false, "display the version")

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

    // Eventually, we will take an input file.
//	if flag.NArg() != 1 {
//		printInstructions()
//		return
//	}

	_ = debug
	f, err := os.Open(*outputPath, os.O_WRONLY|os.O_CREAT, 0644)
	defer f.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	sc := scene.New()

	fmt.Println("Setting up scene...")
	// We should be doing this:
	//ok := parseXMLFile(f, scene)
	// For now, we'll do this:
	sc.SetCamera(camera.NewOrtho(vector.New(0.0, 0.0, 10.0), vector.New(0.0, 0.0, 0.0), vector.New(0.0, 1.0, 10.0), *width, *height, 1.0, 2.0))
	sc.AddObject(object.NewPrimitive(primitive.NewSphere(vector.New(0.0, 0.0, 0.0), 1.0, nil)))
	sc.SetSurfaceIntegrator(trivialInt.New())

	fmt.Println("Rendering...")
    startTime := time.Nanoseconds()
	outputImage, err := sc.Render()
    endTime := time.Nanoseconds()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Rendering error: %v\n", err)
		return
	}
    fmt.Printf("Render finished in %.3fs\n", float(endTime - startTime) * 1e-9)

	fmt.Println("Writing and finishing...")
	switch *format {
	case "png":
		err = png.Encode(f, outputImage)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing: %v\n", err)
		return
	}
}
