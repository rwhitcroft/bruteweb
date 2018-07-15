package main

import (
	"crypto/tls"
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
	userAgent   string
	verb        string
}

func initConfig() {
	// sane defaults
	config.ignoreCodes = []int{404}
	config.userAgent = "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/6.0)"
	config.verb = "GET"

	// properly handle redirs and SSL cert verification errors
	config.httpClient = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
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
