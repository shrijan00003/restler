package svc

import "os"

// TODO: Add support for NO_PROXY
func GetProxyURL() string {
	gProxyUrl := os.Getenv("HTTPS_PROXY")
	if gProxyUrl == "" {
		gProxyUrl = os.Getenv("HTTP_PROXY")
		if gProxyUrl == "" {
			gProxyUrl = ""
		}
	}

	return gProxyUrl
}
