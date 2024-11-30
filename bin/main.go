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
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shrijan00003/restler/bin/svc"
	mainEnv "github.com/shrijan00003/restler/core/env"
	"github.com/shrijan00003/restler/core/utils"

	"github.com/urfave/cli/v2"
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

type Config struct {
	Env        string `yaml:"Env"`
	envPath    string
	configPath string
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

const APP_VERSION = "v0.0.1-dev.9"

var restlerPath string

func main() {
	app := &cli.App{
		Name:    "Restler Application",
		Usage:   "Developer friendly rest client for developers only!!",
		Version: APP_VERSION,
		Commands: []*cli.Command{
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "Run request",
				Action: func(cCtx *cli.Context) error {
					return runAction(cCtx)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

// -------------------------

// -------------------------
// initialize restler project
// -------------------------
func initialize() {
	// initialize env
	initEnv()
	// initialize proxy
	svc.IntializeProxy()
}

func initEnv() error {
	mainEnv.LoadEnv()
	// parser.Parse()
	return nil
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

func runAction(cCtx *cli.Context) error {
	var reqPath = cCtx.Args().First()

	if reqPath == "" {
		log.Fatal("[Resterl Error]: Please provide request args like collection/request-name.yaml")
	}

	// TODO: env support

	if _, err := os.Stat(reqPath); os.IsNotExist(err) {
		log.Fatal("[Restler Error]: Request not found in path: ", reqPath)
	}

	pReq, err := parseRequest(reqPath)
	if err != nil {
		log.Fatal("[Restler Error]: Error processing your request, make sure you have valid format")
	}

	pRes, err := processRequest(pReq)
	if err != nil {
		log.Fatal("[Restler Error]: Error processing your request: ", err)
	}

	body, err := readBody(pRes)
	if err != nil {
		log.Fatal("[Restler error]: Error reading response body, Send PR :D", err)
	}

	responseBytes, err := prepareResponse(pReq, pRes, body)
	if err != nil {
		log.Fatal("[Restler Error]: We can't process your response, Fix and send PR :D", err)
	}

	// TODO: update env file if only it exists
	updateEnvPostScript(cCtx, pReq, pRes, body)

	outDir := filepath.Dir(reqPath)
	baseName := filepath.Base(reqPath)
	outName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	resName := fmt.Sprintf(".%s.%s.%s.res.md", outName, strings.ToLower(pReq.Method), strings.Replace(time.Now().Format("20060102150405.000000"), ".", "", 1))
	resFullPath := filepath.Join(outDir, resName)

	os.WriteFile(resFullPath, responseBytes, 0644)
	return nil
}

func restAction(cCtx *cli.Context, actionName ActionName, restlerPath string) error {
	var req = cCtx.Args().First()
	if req == "" {
		log.Fatal("[Restler Error]: No request provided! Please provide request name as argument.")
	}

	// update env if env flag is set
	// TODO: should be able to take env for nested structure
	envFlag := cCtx.String("env")
	if envFlag != "" {
		config.Env = envFlag
		err := loadWithYaml(fmt.Sprintf("%s/env/%s.yaml", restlerPath, envFlag), &env)
		if err != nil {
			return fmt.Errorf("[error]: Environment you have selected is not found in %s/env folder", restlerPath)
		}
	}

	var reqPath = fmt.Sprintf("%s/%s", restlerPath, req)
	if _, err := os.Stat(reqPath); os.IsNotExist(err) {
		log.Fatal("[Restler Error]: Request directory not found, please check the path. Request Directory Path: ", reqPath)
	}

	// `restler p posts` - process <restler_path>/posts/posts.post.yaml
	// `restler p ga0/posts` - process <restler_path>/ga0/posts/posts.post.yaml
	// `restler p ga0/auth/auth0/token` - process <restler_path>/ga0/auth/auth0/token/token.post.yaml

	reqName := req
	if strings.Contains(req, "/") {
		_paths := strings.Split(req, "/")
		reqName = _paths[len(_paths)-1]
	}

	// support request name with second argument
	if cCtx.Args().Get(1) != "" {
		reqName = cCtx.Args().Get(1)
	}

	// Note: request support with flag
	reqFlag := cCtx.String("request")
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

	updateEnvPostScript(cCtx, pReq, pRes, body)

	outputFilePath := fmt.Sprintf("%s/.%s.%s.res.md", reqPath, reqName, actionName)
	os.WriteFile(outputFilePath, responseBytes, 0644)
	return nil
}

func getNestedValue(data map[string]interface{}, keys string) (interface{}, bool) {
	parts := strings.Split(keys, "][")
	parts[0] = strings.TrimPrefix(parts[0], "[")
	parts[len(parts)-1] = strings.TrimSuffix(parts[len(parts)-1], "]")

	var current interface{} = data
	for _, key := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[key]
		} else if a, ok := current.([]interface{}); ok {
			index, err := strconv.Atoi(key)
			if err != nil || index < 0 || index >= len(a) {
				return nil, false
			}
			current = a[index]
		} else {
			return nil, false
		}
	}
	return current, true
}

func headerToMap(header http.Header) map[string]interface{} {
	result := make(map[string]interface{})
	for key, values := range header {
		if len(values) == 1 {
			result[key] = values[0]
		} else {
			result[key] = values
		}
	}
	return result
}

func updateEnvPostScript(c *cli.Context, req *Request, res *http.Response, body []byte) {
	if req.After == nil || req.After.Env == nil {
		return
	}

	var envBodyMap = map[string]string{}
	var envHeaderMap = map[string]string{}

	for envKey, valKeys := range req.After.Env {
		if valKeys != "" {
			if strings.HasPrefix(valKeys, "Body") {
				envBodyMap[envKey] = strings.TrimPrefix(valKeys, "Body")
			}
			if strings.HasPrefix(valKeys, "Header") {
				envHeaderMap[envKey] = strings.TrimPrefix(valKeys, "Header")
			}
		}
	}

	var envBodyValueMap = map[string]string{}
	var envHeaderValueMap = map[string]string{}

	for envKey, envBodyKey := range envBodyMap {
		var jsonBody map[string]interface{}
		if envBodyKey != "" {
			err := json.Unmarshal(body, &jsonBody)
			if err != nil {
				fmt.Println("[Update Env Log ] Body is not in JSON format: ", err)
				return
			}
			if val, ok := getNestedValue(jsonBody, envBodyKey); ok {
				envBodyValueMap[envKey] = fmt.Sprintf("%v", val)
			} else {
				fmt.Println("[Update Env Log] Value not found for key: ", envBodyKey)
				envBodyValueMap[envKey] = ""
			}
		}
	}

	headerMap := headerToMap(res.Header)
	for envKey, envHeaderKey := range envHeaderMap {
		if val, ok := getNestedValue(headerMap, envHeaderKey); ok {
			envHeaderValueMap[envKey] = fmt.Sprintf("%v", val)
		} else {
			fmt.Println("[Update Env Log] Value not found for key: ", envHeaderKey)
			envHeaderValueMap[envKey] = ""
		}
	}

	_, reqPath := utils.GetReqNamePath(c.Args().First())
	envPath := fmt.Sprintf("%s/env/%s.yaml", reqPath, config.Env)
	// this will override for new API
	if config.envPath != "" {
		envPath = config.envPath
	}
	newEnvMap := convertMap(env)
	mergeMaps(newEnvMap, convertMap(envBodyValueMap))
	mergeMaps(newEnvMap, convertMap(envHeaderValueMap))

	err := writeYAMLFile(envPath, newEnvMap)
	if err != nil {
		fmt.Println("[Restler Log]: Failed to write env file: ", err)
	}
}

func writeYAMLFile(filename string, data map[string]interface{}) error {
	out, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, out, 0644)
}

func convertMap[K comparable, V any](m map[K]V) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[fmt.Sprintf("%v", k)] = v
	}
	return result
}

func mergeMaps[K comparable](dest, src map[K]interface{}) {
	for key, value := range src {
		if v, ok := value.(map[K]interface{}); ok {
			if dv, ok := dest[key].(map[K]interface{}); ok {
				mergeMaps(dv, v)
				continue
			}
		}
		dest[key] = value
	}
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

	u, e := url.Parse(req.URL)
	if e != nil {
		return nil, fmt.Errorf("Not a valid url, error is %w", e)
	}

	// +++++++++++++++++++++++++++++++++++++++++++++
	// support for Parasm for search params
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

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error making http request %s", err)
	}

	return httpResp, nil
}
