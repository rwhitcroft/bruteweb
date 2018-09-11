package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Config struct {
	extension   string
	httpClient  *http.Client
	ignoreCodes []int
	ignoreBody  string
	method      string
	numThreads  int
	recursive   bool
	url         string
	userAgent   string
}

func initConfig() {
	// sane defaults
	config.ignoreCodes = []int{404}
	config.method = "GET"
	config.numThreads = 4
	config.recursive = false
	config.userAgent = "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/6.0)"

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
	var ignoreCodes string

	flag.StringVar(&config.userAgent, "a", config.userAgent, "Set User-Agent string")
	flag.StringVar(&ignoreCodes, "i", "", "Ignore specified status codes, comma-sep (e.g., 403,404,500)")
	flag.StringVar(&config.ignoreBody, "b", "", "Ignore specified string in <body>")
	flag.StringVar(&config.method, "m", config.method, "HTTP method (e.g., GET, HEAD)")
	flag.BoolVar(&config.recursive, "r", config.recursive, "Recurse into subdirectories")
	flag.IntVar(&config.numThreads, "t", config.numThreads, "Number of worker threads")
	flag.StringVar(&config.url, "u", "", "Base URL (e.g., https://example.com:8443/dir/)")
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

	if len(ignoreCodes) > 0 {
		parseIgnoreCodes(ignoreCodes)
	}
}

func parseIgnoreCodes(input string) {
	codes := strings.Split(input, ",")
	for _, v := range codes {
		if num, err := strconv.Atoi(v); err == nil {
			config.ignoreCodes = appendIfUnique(config.ignoreCodes, num)
		} else {
			fmt.Println("Ignoring invalid status code:", v)
		}
	}
}

func appendIfUnique(slice []int, i int) []int {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}
