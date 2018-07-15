package main

import (
	"flag"
	"fmt"
	"net/http"
)

type Config struct {
	extension   string
	httpClient  *http.Client
	ignoreCodes []int
	numThreads  int
	recursive   bool
	url         string
}

func parseCmdLine() {
	flag.StringVar(&config.url, "u", "", "Base URL (e.g., https://example.com:8443/dir/)")
	flag.BoolVar(&config.recursive, "r", false, "Recurse into subdirectories")
	flag.IntVar(&config.numThreads, "t", 4, "Number of worker threads")
	flag.StringVar(&config.extension, "x", "", "Add this file extension to all guesses")
	flag.Parse()

	if config.url == "" {
		flag.Usage()
		panic("Base URL not specified (-u)")
	}

	if config.recursive && config.extension != "" {
		config.recursive = false
		fmt.Println("Can't use recursion with a file extension. Recursion disabled.")
	}

	if config.extension != "" && config.extension[0] == '.' {
		config.extension = config.extension[1:]
	}
}
