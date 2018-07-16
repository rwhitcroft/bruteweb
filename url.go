package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const (
	COLOR_RED   = "\033[31m"
	COLOR_GREEN = "\033[32m"
	COLOR_CYAN  = "\033[36m"
	COLOR_RESET = "\033[0m"
	CLEAR_EOL   = "\033[K"
)

type Url struct {
	fqdn       string
	location   string
	path       []string
	port       int
	proto      string
	statusCode int
}

func (u *Url) AddPathItem(dir string) {
	u.path = append(u.path, dir)
}

func (u *Url) Clone(dir string) *Url {
	url := NewUrl(u.proto, u.fqdn, u.port, u.path)
	url.AddPathItem(dir)
	return url
}

func (u *Url) Fetch() {
	req, err := http.NewRequest(config.method, u.ToString(), nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Set("User-Agent", config.userAgent)

	resp, err := config.httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	u.statusCode = resp.StatusCode
	if u.statusCode == http.StatusFound || u.statusCode == http.StatusMovedPermanently {
		u.location = resp.Header["Location"][0]
	}
}

func (u *Url) Report() {
	var color string
	var location string

	switch u.statusCode {
	case http.StatusOK:
		color = COLOR_GREEN
	case http.StatusFound, http.StatusMovedPermanently:
		color = COLOR_CYAN
	default:
		color = COLOR_RED
	}

	if u.location != "" {
		location = " -> " + u.location
	}

	fmt.Println("\r" + color + strconv.Itoa(u.statusCode) + " " + COLOR_RESET + u.ToString() + location + CLEAR_EOL)
}

func (u *Url) ToString() string {
	port := ":" + strconv.Itoa(u.port)
	if (u.proto == "http" && u.port == 80) || (u.proto == "https" && u.port == 443) {
		port = ""
	}

	ret := string(u.proto + "://" + u.fqdn + port)

	if len(u.path) > 0 {
		ret += "/" + strings.Join(u.path, "/")
	}

	if config.extension == "" {
		ret += "/"
	}

	return ret
}

func NewUrl(proto string, fqdn string, port int, path []string) *Url {
	return &Url{proto: proto, fqdn: fqdn, port: port, path: path}
}

func parseUrl(url string) *Url {
	re := regexp.MustCompile(`^(https?)://([0-9A-Za-z-.]+?)(:\d+)?(/.*)?$`)
	match := re.FindStringSubmatch(url)
	if match == nil {
		panic("Invalid URL specified")
	}

	proto, fqdn, path := match[1], match[2], match[4]
	port, _ := strconv.Atoi(match[3])
	if port == 0 {
		port = 80
		if proto == "https" {
			port = 443
		}
	}

	var dirs []string
	for _, dir := range strings.Split(path, "/") {
		if len(dir) > 0 {
			dirs = append(dirs, dir)
		}
	}

	return NewUrl(proto, fqdn, port, dirs)
}
