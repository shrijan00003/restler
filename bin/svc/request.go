package svc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/shrijan00003/restler/core/app"
	"gopkg.in/yaml.v3"
)

type Request struct {
	Name    string            `yaml:"Name"`
	URL     string            `yaml:"URL"`
	Method  string            `yaml:"Method"`
	Headers map[string]string `yaml:"Headers"`
	Body    interface{}       `yaml:"Body"`
	After   *After            `yaml:"After"`
	Params  map[string]string `yaml:"Params"`
}

type After struct {
	Env map[string]string `yaml:"Env"`
}

func ParseRequest(reqPath string) (*Request, error) {
	rawReq, err := os.ReadFile(reqPath)
	if err != nil {
		return nil, err
	}

	replaced := os.ExpandEnv(string(rawReq))
	req := &Request{}

	err = yaml.Unmarshal([]byte(replaced), req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func ProcessRequest(req *Request, app *app.App) (*http.Response, error) {
	var proxyURL *url.URL = nil
	var transport *http.Transport = nil
	var client *http.Client = nil

	sProxyEnable := req.Headers["R-Proxy-Enable"]
	if sProxyEnable == "" {
		sProxyEnable = "Y"
	}

	if sProxyEnable == "N" {
		client = &http.Client{}
	} else {
		sProxyUrl := req.Headers["R-Proxy-Url"]

		if sProxyUrl == "" {
			sProxyUrl = app.ProxyUrl
		}

		if sProxyUrl != "" {
			var err error
			proxyURL, err = url.Parse(sProxyUrl)
			if err != nil {
				return nil, fmt.Errorf("error parsing proxy url, error: %s", err)
			}
		}

		if proxyURL != nil {
			transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
			client = &http.Client{
				Transport: transport,
			}
		} else {
			client = &http.Client{}
		}
	}

	u, e := url.Parse(req.URL)
	if e != nil {
		return nil, fmt.Errorf("Not a valid url, error is %w", e)
	}

	// +++++++++++++++++++++++++++++++++++++++++++++
	// support for Params for search params
	// +++++++++++++++++++++++++++++++++++++++++++++
	if req.Params != nil {
		q := u.Query()
		for key, val := range req.Params {
			q.Set(key, val)
		}
		u.RawQuery = q.Encode()
	}

	// +++++++++++++++++++++++++++++++++++++++++++++
	// support for application/x-www-form-urlencoded
	// +++++++++++++++++++++++++++++++++++++++++++++
	if req.Headers["Content-Type"] == "application/x-www-form-urlencoded" {
		// this will have support for single nested layer
		rawFormData := url.Values{}
		fmt.Println(req.Body)
		if reflect.TypeOf(req.Body).Kind() == reflect.Map {
			for key, val := range req.Body.(map[string]interface{}) {
				// Note: This structure only works if there is no nested values
				// we should be iterating if type of value is map or list
				value, ok := val.(string)
				if !ok {
					fmt.Println("[error] parsing body for [application/x-www-form-urlencoded]")
				}
				rawFormData.Add(key, value)
			}
		}
		encodedFormData := rawFormData.Encode()
		httpReq, err := http.NewRequest(req.Method, u.String(), strings.NewReader(encodedFormData))
		if err != nil {
			return nil, fmt.Errorf("error creating [application/x-www-form-urlencoded] request %s", err)
		}

		for key, value := range req.Headers {
			httpReq.Header.Set(key, value)
		}

		httpResp, err := client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("error making [application/x-www-form-urlencoded] http request %s", err)
		}
		return httpResp, nil

	}
	// +++++++++++++++++++++++++++++++++++++++++++++
	// json request flow
	// +++++++++++++++++++++++++++++++++++++++++++++
	var parsedBodyBytes []byte
	var err error
	if req.Body != nil {
		parsedBodyBytes, err = json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("error parsing request body %s", err)
		}
	}
	bodyReader := bytes.NewReader(parsedBodyBytes)
	httpReq, err := http.NewRequest(req.Method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("error creating http request %s", err)
	}
	startTime := time.Now()
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error making http request %s", err)
	}

	endTime := time.Now()
	app.RequestTime = endTime.Sub(startTime)

	return httpResp, nil
}
