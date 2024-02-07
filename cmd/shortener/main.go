package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/Svirex/microurl/internal/pkg/app"
)

func getHostFromAddr(addr string) string {
	if strings.HasPrefix(addr, "http://") {
		return addr[7:]
	} else if strings.HasPrefix(addr, "https://") {
		return addr[8:]
	}
	return addr
}

func getBaseURL(addr, baseURL string) string {
	if baseURL != "" {
		if strings.HasPrefix(baseURL, "http://") {
			return baseURL
		} else {
			return "http://" + baseURL
		}
	}
	host := getHostFromAddr(addr)
	baseURL = "http://" + host
	return baseURL
}

func main() {
	var addr string
	var baseURL string

	flag.StringVar(&addr, "a", "localhost:8080", "<host>:<port>")
	flag.StringVar(&baseURL, "b", "", "base url")
	flag.Parse()
	host := getHostFromAddr(addr)
	baseURL = getBaseURL(addr, baseURL)
	fmt.Println(addr, host, baseURL)
	app.Run(host, baseURL)
}
