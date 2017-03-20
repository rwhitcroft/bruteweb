package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Url struct {
	fqdn        string
	path        []string
	port        int
	proto       string
	location    string
	status_code int
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
	req, err := http.NewRequest("HEAD", u.Flatten(), nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := config.http_client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	u.status_code = resp.StatusCode
	if u.status_code == http.StatusFound || u.status_code == http.StatusMovedPermanently {
		u.location = resp.Header["Location"][0]
	}
}

func (u *Url) Flatten() string {
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
	url := new(Url)
	url.proto = proto
	url.fqdn = fqdn
	url.port = port
	url.path = path

	return url
}

func ParseURL(url string) *Url {
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
