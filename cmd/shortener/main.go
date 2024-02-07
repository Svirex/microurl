package main

import (
	"flag"
	"fmt"

	"github.com/Svirex/microurl/internal/pkg/app"
)

func main() {
	var host string
	var baseURL string

	flag.StringVar(&host, "a", "localhost:8080", "<host>:<port>")
	flag.StringVar(&baseURL, "b", "", "base url")
	flag.Parse()
	if baseURL == "" {
		baseURL = host
	}
	fmt.Println(host, baseURL)
	app.Run(host, baseURL)
}
