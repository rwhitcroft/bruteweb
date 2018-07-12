package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"sync"
)

type Config struct {
	extension  string
	httpClient *http.Client
	numThreads int
	recursive  bool
	url        string
}

var config Config

func ParseCmdLine() {
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

func main() {
	ParseCmdLine()

	config.httpClient = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	var urls []*Url
	urls = append(urls, ParseURL(config.url))

	for len(urls) > 0 {
		base := urls[0]
		urls = urls[1:]

		fmt.Println("Scanning base URL:", base.Flatten())

		words := make(chan string, config.numThreads)
		results := make(chan *Url)

		var wg_producers sync.WaitGroup
		var wg_consumers sync.WaitGroup
		wg_producers.Add(config.numThreads)
		wg_consumers.Add(1)

		for i := 0; i < config.numThreads; i++ {
			go func() {
				defer wg_producers.Done()
				for {
					word := <-words
					if word == "" {
						break
					}

					if config.extension != "" {
						word += "." + config.extension
					}

					url := base.Clone(word)
					url.Fetch()
					results <- url
				}
			}()
		}

		go func() {
			defer wg_consumers.Done()
			for result := range results {
				print_status(result)
				if result.statusCode != http.StatusNotFound {
					result.Report()
				}

				if config.recursive && result.statusCode == http.StatusOK {
					urls = append(urls, result)
				}
			}
		}()

		for _, word := range *GetWords() {
			words <- word
		}

		close(words)
		wg_producers.Wait()
		close(results)
		wg_consumers.Wait()
	}

	fmt.Println("\r" + CLEAR_EOL)
}

func print_status(u *Url) {
	fmt.Print("\r" + u.Flatten() + CLEAR_EOL)
}
