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
    "./goray/scene"
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
    f, err := os.Open(outputPath, os.O_WRONLY | os.O_CREAT, 0644)
    defer f.Close()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return
    }
    sc := scene.New()
    
    fmt.Println("Setting up scene...")
    //ok := parseXMLFile(f, scene)
    
    fmt.Println("Rendering...")
    outputImage := sc.Render()
    
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
