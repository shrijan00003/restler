package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shrijan00003/restler/bin/svc"
	"github.com/shrijan00003/restler/core/app"
	"github.com/shrijan00003/restler/core/env"
	"github.com/shrijan00003/restler/core/logger"
	"github.com/shrijan00003/restler/core/utils"
	"gopkg.in/yaml.v3"

	"github.com/urfave/cli/v2"
)

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

const APP_VERSION = "v0.0.2-dev.0"

var restlerPath string
var a *app.App

func main() {
	defer func() {
		a.Terminate()
	}()
	run()
}

// -------------------------

// -------------------------
// initialize restler project
// -------------------------
func initialize() {
	defer func() {
		env.Terminate()
		logger.Terminate()
	}()

	logger.Init()

	// load config.yaml file in the root of restler project
	// TODO: support config flag to load config file from request
	pConfig, _ := svc.LoadConfig()

	// Load default env with .env and .env.local
	env.LoadEnv(pConfig)
	proxyURL := svc.GetProxyURL()
	a = app.NewApp(proxyURL, APP_VERSION, pConfig)
}

func run() {
	initialize()
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

	if _, err := os.Stat(reqPath); os.IsNotExist(err) {
		log.Fatal("[Restler Error]: Request not found in path: ", reqPath)
	}

	pReq, err := svc.ParseRequest(reqPath)
	if err != nil {
		logger.Debug("error processing request:", err)
		log.Fatal("[restler Error]: Error processing your request, make sure you have valid format")
	}

	pRes, err := svc.ProcessRequest(pReq, a)
	if err != nil {
		log.Fatal("[restler Error]: Error processing your request: ", err)
	}

	body, err := utils.ReadBody(pRes)
	if err != nil {
		log.Fatal("[restler error]: Error reading response body, Send PR :D", err)
	}

	responseBytes, err := prepareResponse(pReq, pRes, body)
	if err != nil {
		log.Fatal("[restler Error]: We can't process your response, Fix and send PR :D", err)
	}

	// TODO: update env file if only it exists
	updateEnvPostScript(pReq, pRes, body)

	outDir := filepath.Dir(reqPath)
	baseName := filepath.Base(reqPath)
	outName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	newDir := filepath.Join(outDir, ".res."+outName)
	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		os.Mkdir(newDir, 0755)
	}
	resName := fmt.Sprintf(".%s.%s.%s.res.md", outName, strings.ToLower(pReq.Method), strings.Replace(time.Now().Format("20060102150405.000000"), ".", "", 1))
	resFullPath := filepath.Join(newDir, resName)

	os.WriteFile(resFullPath, responseBytes, 0644)
	return nil
}

func updateEnvPostScript(req *svc.Request, res *http.Response, body []byte) {
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
			if val, ok := utils.GetNestedValue(jsonBody, envBodyKey); ok {
				envBodyValueMap[envKey] = fmt.Sprintf("%v", val)
			} else {
				fmt.Println("[Update Env Log] Value not found for key: ", envBodyKey)
				envBodyValueMap[envKey] = ""
			}
		}
	}

	headerMap := utils.HeaderToMap(res.Header)
	for envKey, envHeaderKey := range envHeaderMap {
		if val, ok := utils.GetNestedValue(headerMap, envHeaderKey); ok {
			envHeaderValueMap[envKey] = fmt.Sprintf("%v", val)
		} else {
			fmt.Println("[Update Env Log] Value not found for key: ", envHeaderKey)
			envHeaderValueMap[envKey] = ""
		}
	}

	// TODO: Verify if this works
	newEnvMap := make(map[string]interface{}, len(envBodyValueMap)+len(envHeaderValueMap))
	utils.MergeMaps(newEnvMap, utils.ConvertMap(envBodyValueMap))
	utils.MergeMaps(newEnvMap, utils.ConvertMap(envHeaderValueMap))
	err := env.UpdateEnvFile(a, newEnvMap)

	if err != nil {
		fmt.Println("[restler Log]: Failed to write env file: ", err)
	}
}

func getRequestBytes(req *svc.Request) ([]byte, error) {
	if req.Body != nil {
		return yaml.Marshal(req)
	}
	return nil, nil
}

func prepareResponse(req *svc.Request, res *http.Response, body []byte) ([]byte, error) {

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("# Response For: %s \n", req.Name))
	buffer.WriteString(fmt.Sprintf("Status Code: %d, Status: %s\n", res.StatusCode, res.Status))
	buffer.WriteString("\n\n")
	buffer.WriteString("## Request Time\n")
	buffer.WriteString(a.RequestTime.String())
	buffer.WriteString("\n\n## Response Header: \n")

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
	// ignoring errors here
	requestBytes, _ := getRequestBytes(req)
	buffer.WriteString("\n```yaml\n")
	buffer.Write(requestBytes)
	buffer.WriteString("\n```")
	return buffer.Bytes(), nil
}

func validateRequest(r *svc.Request) error {
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
