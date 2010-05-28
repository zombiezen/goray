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
)

func printInstructions() {
    fmt.Println("USAGE: goray [OPTION]... FILE")
    fmt.Println("OPTIONS:")
    flag.PrintDefaults()
}

func main() {
    help := flag.Bool("h", false, "display this help")
    format := flag.String("f", "png", "the output format")
    outputPath := flag.String("o", "", "path for the output file")
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
    
    _, _, _ = format, outputPath, debug
}
