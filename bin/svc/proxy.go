package svc

import "os"

// Load proxy from env supports both HTTPS_PROXY and HTTP_PROXY
var gProxyUrl string

func IntializeProxy() {
	gProxyUrl = os.Getenv("HTTPS_PROXY")
	if gProxyUrl == "" {
		gProxyUrl = os.Getenv("HTTP_PROXY")
		if gProxyUrl == "" {
			gProxyUrl = ""
		}
	}
}
