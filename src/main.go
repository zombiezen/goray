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
	"./logging"
	"./goray/camera"
	"./goray/object"
	"./goray/render"
	"./goray/scene"
	"./goray/vector"
	"./goray/version"
	trivialInt "./goray/std/integrators/trivial"
	"./goray/std/objects/mesh"
	"./goray/std/primitives/sphere"
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
	// Back
	cube.AddTriangle(mesh.NewTriangle(0, 3, 2, cube))
	cube.AddTriangle(mesh.NewTriangle(0, 2, 1, cube))
	// Top
	cube.AddTriangle(mesh.NewTriangle(3, 7, 2, cube))
	cube.AddTriangle(mesh.NewTriangle(6, 2, 7, cube))
	// Bottom
	cube.AddTriangle(mesh.NewTriangle(0, 1, 4, cube))
	cube.AddTriangle(mesh.NewTriangle(5, 4, 1, cube))
	// Left
	cube.AddTriangle(mesh.NewTriangle(7, 3, 4, cube))
	cube.AddTriangle(mesh.NewTriangle(0, 4, 3, cube))
	// Right
	cube.AddTriangle(mesh.NewTriangle(6, 5, 2, cube))
	cube.AddTriangle(mesh.NewTriangle(1, 2, 5, cube))
	// Front
	cube.AddTriangle(mesh.NewTriangle(4, 6, 7, cube))
	cube.AddTriangle(mesh.NewTriangle(5, 6, 4, cube))

	sc.SetCamera(camera.NewOrtho(vector.New(5.0, 5.0, 5.0), vector.New(0.0, 0.0, 0.0), vector.New(5.0, 6.0, 5.0), *width, *height, 1.0, 3.0))
	sc.AddObject(cube)
	sc.AddObject(object.PrimitiveObject{sphere.New(vector.New(1, 0, 1), 0.5, nil)})
	sc.SetSurfaceIntegrator(trivialInt.New())

	logging.MainLog.Info("Finalizing scene...")
	finalizeTime := stopwatch(func() {
		sc.Update()
	})
	logging.MainLog.Info("Finalized in %v", finalizeTime)

	logging.MainLog.Info("Rendering...")

	var outputImage *render.Image
	renderTime := stopwatch(func() {
		outputImage, err = sc.Render()
	})
	if err != nil {
		logging.MainLog.Error("Rendering error: %v", err)
		return
	}
	logging.MainLog.Info("Render finished in %v", renderTime)

	logging.MainLog.Info("TOTAL TIME: %v", addTime(finalizeTime, renderTime))

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

type Time float64

func stopwatch(f func()) Time {
	startTime := time.Nanoseconds()
	f()
	endTime := time.Nanoseconds()
	return Time(float64(endTime-startTime) * 1e-9)
}

func addTime(t1, t2 Time, tn ...Time) Time {
	accum := float64(t1) + float64(t2)
	for _, t := range tn {
		accum += float64(t)
	}
	return Time(accum)
}

func (t Time) Split() (hours, minutes int, seconds float64) {
	const secondsPerMinute = 60
	const secondsPerHour = secondsPerMinute * 60

	seconds = float64(t)
	hours = int(t / secondsPerHour)
	seconds -= float64(hours * secondsPerHour)
	minutes = int(seconds / secondsPerMinute)
	seconds -= float64(minutes * secondsPerMinute)
	return
}

func (t Time) String() string {
	h, m, s := t.Split()
	switch {
	case h == 0 && m == 0:
		return fmt.Sprintf("%.3fs", s)
	case h == 0:
		return fmt.Sprintf("%02d:%05.2f", m, s)
	}
	return fmt.Sprintf("%d:%02d:%05.2f", h, m, s)
}
