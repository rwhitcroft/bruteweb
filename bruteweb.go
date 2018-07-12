package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"sync"
)

type Config struct {
	extension   string
	http_client *http.Client
	method      string
	num_threads int
	recursive   bool
	url         string
}

var config Config

func ParseCmdLine() {
	method_head := false

	flag.BoolVar(&method_head, "H", false, "Use HEAD instead of GET")
	flag.StringVar(&config.url, "u", "", "Base URL (e.g., https://example.com:8443/dir/)")
	flag.BoolVar(&config.recursive, "r", false, "Recurse into subdirectories")
	flag.IntVar(&config.num_threads, "t", 4, "Number of worker threads")
	flag.StringVar(&config.extension, "x", "", "Add this file extension to all guesses")
	flag.Parse()

	if config.url == "" {
		flag.Usage()
		panic("Base URL not specified (-u)")
	}

	config.method = "GET"
	if method_head {
		config.method = "HEAD"
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

	config.http_client = &http.Client{
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
		base_url := urls[0]
		urls = urls[1:]

		fmt.Println("Scanning base URL:", base_url.Flatten())

		words := make(chan string, config.num_threads)
		results := make(chan *Url)

		var wg_producers sync.WaitGroup
		var wg_consumers sync.WaitGroup
		wg_producers.Add(config.num_threads)
		wg_consumers.Add(1)

		for i := 0; i < config.num_threads; i++ {
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

					url := base_url.Clone(word)
					url.Fetch()
					results <- url
				}
			}()
		}

		go func() {
			defer wg_consumers.Done()
			for result := range results {
				print_status(result)
				if result.status_code != http.StatusNotFound {
					result.Report()
				}

				if config.recursive && result.status_code == http.StatusOK {
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
