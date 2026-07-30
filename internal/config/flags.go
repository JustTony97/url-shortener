package config

import "flag"

var RunAddr string
var RedirectBaseUrl string

func init() {
	flag.StringVar(&RunAddr, "a", ":8080", "Server address")
	flag.StringVar(&RedirectBaseUrl, "b", "http://localhost:8080", "Base URL for short links redirection")
}
