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
    "./goray/camera"
    "./goray/object"
    "./goray/primitive"
    "./goray/scene"
    "./goray/vector"
)

func printInstructions() {
    fmt.Println("USAGE: goray [OPTION]... FILE")
    fmt.Println("OPTIONS:")
    flag.PrintDefaults()
}

func main() {
    var err os.Error
    
    help := flag.Bool("h", false, "display this help")
    format := flag.String("f", "png", "the output format")
    outputPath := flag.String("o", "goray.png", "path for the output file")
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
    
    if flag.NArg() != 1 {
        printInstructions()
        return
    }
    
    _ = debug
    f, err := os.Open(*outputPath, os.O_WRONLY | os.O_CREAT, 0644)
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
    sc.SetCamera(camera.NewOrtho(vector.New(0.0, 0.0, 10.0), vector.New(0.0, 0.0, 0.0), vector.New(0.0, 1.0, 0.0), 640, 480, 1.0, 1.0))
    sc.AddObject(object.NewPrimitive(primitive.NewSphere(vector.New(0.0, 0.0, 0.0), 1.0, nil)))
    
    fmt.Println("Rendering...")
    outputImage, err := sc.Render()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Rendering error: %v\n", err)
        return
    }
    
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
