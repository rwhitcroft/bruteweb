package main

import (
	"fmt"
	"net/http"
	"strconv"
)

const (
	COLOR_RED   = "\033[31m"
	COLOR_GREEN = "\033[32m"
	COLOR_CYAN  = "\033[36m"
	COLOR_RESET = "\033[0m"
	CLEAR_EOL   = "\033[K"
)

func print_hit(u *Url) {
	var color string
	var location string

	switch u.status_code {
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

	fmt.Println("\r" + color + strconv.Itoa(u.status_code) + " " + COLOR_RESET + u.Flatten() + location + CLEAR_EOL)
}

func print_status(u *Url) {
	fmt.Print("\r" + u.Flatten() + CLEAR_EOL)
}
