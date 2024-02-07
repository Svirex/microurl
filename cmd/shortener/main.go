package main

import (
	"flag"
	"fmt"
	"net/url"

	"github.com/Svirex/microurl/internal/pkg/app"
)

func getHostFromAddr(addr string) (string, error) {
	a, err := url.Parse(addr)
	if err != nil {
		return "", err
	}
	return a.Host, nil
}

func main() {
	var host string
	var baseURL string

	flag.StringVar(&host, "a", "localhost:8080", "<host>:<port>")
	flag.StringVar(&baseURL, "b", "", "base url")
	flag.Parse()
	host, err := getHostFromAddr(host)
	if err != nil {
		panic(err)
	}
	if baseURL == "" {
		baseURL = host
	}

	fmt.Println(host, baseURL)
	app.Run(host, baseURL)
}
