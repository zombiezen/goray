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
    "time"
	"./goray/camera"
	"./goray/object"
	"./goray/primitive"
	"./goray/scene"
	"./goray/vector"
	trivialInt "./goray/std/integrators/trivial"
)

func printInstructions() {
	fmt.Println("USAGE: goray [OPTION]... FILE")
	fmt.Println("OPTIONS:")
	flag.PrintDefaults()
}

func main() {
	var err os.Error

	help := flag.Bool("help", false, "display this help")
	format := flag.String("f", "png", "the output format")
	outputPath := flag.String("o", "goray.png", "path for the output file")
    width := flag.Int("w", 100, "the output width")
    height := flag.Int("h", 100, "the output height")
	debug := flag.Int("d", 0, "set debug verbosity level")
	version := flag.Bool("v", false, "display the version")

	flag.Usage = printInstructions
	flag.Parse()

	switch {
	case *help:
		printInstructions()
		return
	case *version:
		fmt.Println("This is SPARTA!")
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
