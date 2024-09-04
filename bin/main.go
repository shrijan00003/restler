package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

type Request struct {
	Name    string            `yaml:"Name"`
	URL     string            `yaml:"URL"`
	Method  string            `yaml:"Method"`
	Headers map[string]string `yaml:"Headers"`
	Body    interface{}       `yaml:"Body"`
}

type Config struct {
	Env string `yaml:"Env"`
}

// global Config variable
var config Config

func (c *Config) DefaultConfig() {
	c.Env = "default"
}

func (c *Config) Terminate() {
	c = nil
}

// global environment map
var env map[string]string

// global proxy url
var gProxyUrl string

const APP_VERSION = "v0.0.1-dev.4"

func main() {
	// RESTLER_PATH path, where to run command to create api request
	var restlerPath = os.Getenv("RESTLER_PATH")
	if restlerPath == "" {
		fmt.Println("[Restler Log]:RESTLER_PATH is not set, defaulting to restler")
		restlerPath = "restler"
	}

	// Load proxy from env supports both HTTPS_PROXY and HTTP_PROXY
	gProxyUrl = os.Getenv("HTTPS_PROXY")
	if gProxyUrl == "" {
		gProxyUrl = os.Getenv("HTTP_PROXY")
		if gProxyUrl == "" {
			gProxyUrl = ""
		}
	}

	// Load configs
	err := loadWithYaml(fmt.Sprintf("%s/config.yaml", restlerPath), &config)
	if err != nil {
		fmt.Println("[Restler Log]:Failed to load config file, taking all defaults, err:", err)
		config.DefaultConfig()
	}

	// Load Environment
	err = loadWithYaml(fmt.Sprintf("%s/env/%s.yaml", restlerPath, config.Env), &env)
	if err != nil {
		fmt.Printf("[Restler Error]: Failed to load environment file! Make sure you have at least default.yaml file in %s/env folder to use environment variables in request!\n", restlerPath)
	}

	commonCommandFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "request",
			Aliases: []string{"r"},
			Usage:   "request to execute",
			Value:   "",
		},
	}

	app := &cli.App{
		Name:    "Restler Application",
		Usage:   "Developer friendly rest client for developers only!!",
		Version: APP_VERSION,
		Commands: []*cli.Command{
			{
				Name:    "post",
				Aliases: []string{"p"},
				Usage:   "Run post request",
				Flags:   commonCommandFlags,
				Action: func(cCtx *cli.Context) error {
					return restAction(cCtx, POST, restlerPath)
				},
			},
			{
				Name:    "get",
				Aliases: []string{"g"},
				Usage:   "Run get request",
				Flags:   commonCommandFlags,
				Action: func(cCtx *cli.Context) error {
					return restAction(cCtx, GET, restlerPath)
				},
			},
			{
				Name:    "put",
				Aliases: []string{"u"},
				Usage:   "Run put request",
				Flags:   commonCommandFlags,
				Action: func(cCtx *cli.Context) error {
					return restAction(cCtx, PUT, restlerPath)
				},
			},
			{
				Name:    "delete",
				Aliases: []string{"d"},
				Usage:   "Run delete request",
				Flags:   commonCommandFlags,
				Action: func(cCtx *cli.Context) error {
					return restAction(cCtx, DELETE, restlerPath)
				},
			},
			{
				Name:    "patch",
				Aliases: []string{"m"},
				Usage:   "Run patch request",
				Flags:   commonCommandFlags,
				Action: func(cCtx *cli.Context) error {
					return restAction(cCtx, PATCH, restlerPath)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

type ActionName string

const (
	POST    ActionName = "post"
	GET     ActionName = "get"
	PUT     ActionName = "put"
	DELETE  ActionName = "delete"
	PATCH   ActionName = "patch"
	OPTIONS ActionName = "options"
	HEAD    ActionName = "head"
)

func restAction(cCtx *cli.Context, actionName ActionName, restlerPath string) error {

	var req = cCtx.Args().Get(0)
	if req == "" {
		log.Fatal("[Restler Error]: No request provided! Please provide request name as argument. Request name is the name of the folder in requests folder.")
	}

	var reqPath = fmt.Sprintf("%s/requests/%s", restlerPath, req)
	if _, err := os.Stat(reqPath); os.IsNotExist(err) {
		log.Fatal("[Restler Error]: Request directory not found, please check the path. Request Directory Path: ", reqPath)
	}

	// Note: request support with flag
	reqFlag := cCtx.String("request")

	reqName := req
	if reqFlag != "" {
		reqName = reqFlag
	}

	var reqFullPath = fmt.Sprintf("%s/%s.%s.yaml", reqPath, reqName, actionName)
	fmt.Println("[Restler Log]: Processing Request: ", reqFullPath)

	if _, err := os.Stat(reqFullPath); os.IsNotExist(err) {
		log.Fatal("[Restler Error]: Request file not found, please check the path. Request File Path: ", reqFullPath)
	}

	pReq, err := parseRequest(reqFullPath)
	if err != nil {
		log.Fatal("[Restler error] Error parsing request: ", err)
	}

	pRes, err := processRequest(pReq)
	if err != nil {
		log.Fatal("[Restler error] Error processing request: ", err)
	}

	body, err := readBody(pRes)
	if err != nil {
		log.Fatal("[Restler error]: ", err)
	}

	responseBytes, err := prepareResponse(pReq, pRes, body)
	if err != nil {
		log.Fatal("[Restler error]: ", err)
	}

	outputFilePath := fmt.Sprintf("%s/.%s.%s.res.md", reqPath, reqName, actionName)
	os.WriteFile(outputFilePath, responseBytes, 0644)
	return nil
}

func prepareResponse(req *Request, res *http.Response, body []byte) ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("# Response For: %s \n", req.Name))
	buffer.WriteString(fmt.Sprintf("Status Code: %d, Status: %s\n", res.StatusCode, res.Status))
	buffer.WriteString("\n\n")
	buffer.WriteString("## Response Header: \n")

	for key, value := range res.Header {
		buffer.WriteString(fmt.Sprintf("%s: %s\n", key, value))
	}
	buffer.WriteString("\n\n")
	buffer.WriteString("## Response Body: \n")
	buffer.WriteString("```json\n")
	buffer.Write(body)
	buffer.WriteString("\n```")
	buffer.WriteString("\n\n")
	buffer.WriteString("## Original Request \n")
	buffer.WriteString(fmt.Sprintf("Method: %s, URL: %s\n", res.Request.Method, res.Request.URL))
	return buffer.Bytes(), nil
}

func readBody(res *http.Response) ([]byte, error) {
	var reader io.ReadCloser
	var err error

	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		{
			reader, err = gzip.NewReader(res.Body)
			if err != nil {
				return nil, fmt.Errorf("error creating gzip reader : %v", err)
			}
		}
	default:
		{
			reader = res.Body
		}
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading response body : %v", err)
	}

	return body, nil
}

func loadEnvInRequest(input string) string {
	re := regexp.MustCompile(`\{\s*\{\s*(\w+)\s*\}\s*\}`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		key := strings.Trim(match, "{} \t")
		if value, exists := env[key]; exists {
			return value
		}
		return match // Return the original if not found in env
	})
}

func loadWithYaml(configPath string, receiver interface{}) error {
	contentBytes, err := readFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading file at %s", configPath)
	}

	err = yaml.Unmarshal(contentBytes, receiver)
	if err != nil {
		return fmt.Errorf("error occurred on parsing Yaml file : %s", configPath)
	}

	return nil
}

func readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file", err)
		return nil, err
	}
	defer file.Close()
	rawContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file at path %s", path)
	}

	return rawContent, nil
}

func parseRequest(requestPath string) (*Request, error) {
	file, err := os.Open(requestPath)
	if err != nil {
		fmt.Println("Error opening file", err)
		return nil, err
	}
	defer file.Close()
	rawRequestContent, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file", err)
		return nil, err
	}
	rawRequestWithEnv := loadEnvInRequest(string(rawRequestContent))

	var request Request
	err = yaml.Unmarshal([]byte(rawRequestWithEnv), &request)
	if err != nil {
		fmt.Println("Error parsing request file, please check the format", err)
		return nil, err
	}
	err = validateRequest(&request)
	if err != nil {
		fmt.Println("Error validating request file, Error: ", err)
		return nil, err
	}

	return &request, nil
}

func validateRequest(r *Request) error {
	if r.Name == "" {
		return errors.New("Request name is required")
	}
	if r.URL == "" {
		return errors.New("Request URL is required")
	}
	if r.Method == "" {
		return errors.New("Request Method is required")
	}
	if r.Headers == nil {
		return errors.New("Request Headers is required")
	}
	var isBodyOptional = r.Method == "GET" || r.Method == "OPTIONS" || r.Method == "DELETE"

	if !isBodyOptional && r.Body == nil {
		return errors.New("Request Body is required for non-GET, OPTIONS, and DELETE requests")
	}

	return nil
}

func processRequest(req *Request) (*http.Response, error) {
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
			sProxyUrl = gProxyUrl
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

	var _bytes []byte
	var err error
	if req.Body != nil {
		_bytes, err = json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("error parsing request body %s", err)
		}
	}
	bodyReader := bytes.NewReader(_bytes)
	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("error creating http request %s", err)
	}

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error making http request %s", err)
	}
	return httpResp, nil
}
