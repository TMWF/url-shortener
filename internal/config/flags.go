package config

import "flag"

var ServerHost string
var BaseURL string

func ParseFlags() {
	flag.StringVar(&ServerHost, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&BaseURL, "b", "http://localhost:8080", "address and port to run server")
	flag.Parse()
}
