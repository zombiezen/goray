//
//  server.go
//  goray
//
//  Created by Ross Light on 2011-02-05.
//

package main

import (
	"flag"
	"http"
	"goray/server"
)

func main() {
	flag.Parse()
	if flag.NArg() != 2 {
		return
	}

	s := server.New(flag.Arg(1), "data")
	http.ListenAndServe(flag.Arg(0), s)
}
