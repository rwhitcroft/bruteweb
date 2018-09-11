package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

var config Config

func main() {
	initConfig()
	parseCmdLine()

	var urls []*Url
	urls = append(urls, parseUrl(config.url))

	for len(urls) > 0 {
		base := urls[0]
		urls = urls[1:]

		fmt.Println("Scanning base URL:", base.ToString())

		words := make(chan string, config.numThreads)
		results := make(chan *Url)

		var wgProducers sync.WaitGroup
		var wgConsumers sync.WaitGroup
		wgProducers.Add(config.numThreads)
		wgConsumers.Add(1)

		for i := 0; i < config.numThreads; i++ {
			go func() {
				defer wgProducers.Done()
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

					if !strings.Contains(url.body, config.ignoreBody) {
						results <- url
					}
				}
			}()
		}

		go func() {
			defer wgConsumers.Done()
			for result := range results {
				printStatus(result)
				if !codeIsIgnored(result.statusCode) {
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
		wgProducers.Wait()
		close(results)
		wgConsumers.Wait()
	}

	fmt.Println("\r" + CLEAR_EOL)
}

func codeIsIgnored(code int) bool {
	for _, v := range config.ignoreCodes {
		if v == code {
			return true
		}
	}
	return false
}

func printStatus(u *Url) {
	fmt.Print("\r" + u.ToString() + CLEAR_EOL)
}
