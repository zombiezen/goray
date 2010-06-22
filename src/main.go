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
	"goray/core/camera"
	"goray/core/color"
	"goray/core/integrator"
	"goray/core/object"
	"goray/core/render"
	"goray/core/scene"
	"goray/core/vector"
	"goray/core/version"
	"goray/std/integrators/directlight"
	pointLight "goray/std/lights/point"
	debugMaterial "goray/std/materials/debug"
	"goray/std/objects/mesh"
	"goray/std/primitives/sphere"
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
	fmt.Printf("Built for %s (%s)\n", syscall.OS, syscall.ARCH)
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

	// Set up logging
	level := logging.InfoLevel - 10*(*debug)
	logging.MainLog.AddHandler(logging.NewMinLevelFilter(level, logging.NewWriterHandler(os.Stdout)))
	defer logging.MainLog.Close()

	// Open output file
	f, err := os.Open(*outputPath, os.O_WRONLY|os.O_CREAT, 0644)
	defer f.Close()
	if err != nil {
		logging.MainLog.Critical("Error opening output file: %v", err)
		return
	}
	sc := scene.New()

	logging.MainLog.Info("Setting up scene...")
	sceneFilter := func(rec logging.Record) logging.Record {
		return logging.BasicRecord{"  SCENE: " + rec.String(), rec.Level()}
	}
	sc.GetLog().AddHandler(logging.Filter{logging.MainLog, sceneFilter})
	// We should be doing this:
	//ok := parseXMLFile(f, scene)
	// For now, we'll do this:
	cube := mesh.New(12, false)
	cube.SetData([]vector.Vector3D{
		vector.New(-0.5, -0.5, -0.5),
		vector.New(0.5, -0.5, -0.5),
		vector.New(0.5, 0.5, -0.5),
		vector.New(-0.5, 0.5, -0.5),

		vector.New(-0.5, -0.5, 0.5),
		vector.New(0.5, -0.5, 0.5),
		vector.New(0.5, 0.5, 0.5),
		vector.New(-0.5, 0.5, 0.5),
	},
		nil, nil)
	faces := [][3]int{
		// Back
		[3]int{0, 3, 2},
		[3]int{0, 2, 1},
		// Top
		[3]int{3, 7, 2},
		[3]int{6, 2, 7},
		// Bottom
		[3]int{0, 1, 4},
		[3]int{5, 4, 1},
		// Left
		[3]int{7, 3, 4},
		[3]int{0, 4, 3},
		// Right
		[3]int{6, 5, 2},
		[3]int{1, 2, 5},
		// Front
		[3]int{4, 6, 7},
		[3]int{5, 6, 4},
	}
	
	mat := debugMaterial.New(color.NewRGB(1.0, 1.0, 1.0))
	for _, fdata := range faces {
		face := mesh.NewTriangle(fdata[0], fdata[1], fdata[2], cube)
		face.SetMaterial(mat)
		cube.AddTriangle(face)
	}

	sc.SetCamera(camera.NewOrtho(vector.New(5.0, 5.0, 5.0), vector.New(0.0, 0.0, 0.0), vector.New(5.0, 6.0, 5.0), *width, *height, 1.0, 3.0))
	sc.AddLight(pointLight.New(vector.New(10.0,  0.0,  0.0), color.NewRGB(1.0, 0.0, 0.0), 200.0))
	sc.AddLight(pointLight.New(vector.New( 0.0, 10.0,  0.0), color.NewRGB(0.0, 1.0, 0.0), 100.0))
	sc.AddLight(pointLight.New(vector.New( 0.0,  0.0, 10.0), color.NewRGB(0.0, 0.0, 1.0), 50.0))
	sc.AddObject(cube)
	sc.AddObject(object.PrimitiveObject{sphere.New(vector.New(1, 0, 0), 0.5, mat)})

	logging.MainLog.Info("Finalizing scene...")
	finalizeTime := time.Stopwatch(func() {
		sc.Update()
	})
	logging.MainLog.Info("Finalized in %v", finalizeTime)

	logging.MainLog.Info("Rendering...")

	var outputImage *render.Image
	renderFilter := func(rec logging.Record) logging.Record {
		return logging.BasicRecord{"  RENDER: " + rec.String(), rec.Level()}
	}
	renderTime := time.Stopwatch(func() {
		outputImage = integrator.Render(sc, directlight.New(false, 3, 10), logging.Filter{logging.MainLog, renderFilter})
	})
	if err != nil {
		logging.MainLog.Error("Rendering error: %v", err)
		return
	}
	logging.MainLog.Info("Render finished in %v", renderTime)

	logging.MainLog.Info("TOTAL TIME: %v", finalizeTime + renderTime)

	logging.MainLog.Info("Writing and finishing...")
	switch *format {
	case "png":
		err = png.Encode(f, outputImage)
	}

	if err != nil {
		logging.MainLog.Critical("Error while writing: %v", err)
		return
	}
}
