package main

import (
	"bufio"
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

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	mainEnv "github.com/shrijan00003/restler/core/env"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

const samplePostRequestUrl = "https://raw.githubusercontent.com/shrijan00003/restler/main/sample/requests/sample.post.yaml"
const sampleGetRequestUrl = "https://raw.githubusercontent.com/shrijan00003/restler/main/sample/requests/sample.get.yaml"
const samplePutRequestUrl = "https://raw.githubusercontent.com/shrijan00003/restler/main/sample/requests/sample.put.yaml"
const sampleDeleteRequestUrl = "https://raw.githubusercontent.com/shrijan00003/restler/main/sample/requests/sample.delete.yaml"
const samplePatchRequestUrl = "https://raw.githubusercontent.com/shrijan00003/restler/main/sample/requests/sample.patch.yaml"

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

func Main() {
	commonCommandFlags := []cli.Flag{
		&cli.StringFlag{
			Name:    "request",
			Aliases: []string{"r"},
			Usage:   "Select request to execute",
		},
		&cli.StringFlag{
			Name:    "env",
			Aliases: []string{"e"},
			Usage:   "Select env for request",
		},
	}

	app := &cli.App{
		Name:    "Restler Application",
		Usage:   "Developer friendly rest client for developers only!!",
		Version: APP_VERSION,
		Commands: []*cli.Command{
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "Initialize restler project",
				Action: func(cCtx *cli.Context) error {
					return initRestlerProject()
				},
			},
			{
				Name:    "test",
				Aliases: []string{"t"},
				Usage:   "Test the restler project for different purposes",
				Action: func(cCtx *cli.Context) error {
					return initEnv()
				},
			},
			{
				Name:    "create-collection",
				Aliases: []string{"cc"},
				Action: func(cCtx *cli.Context) error {
					initialize(cCtx)
					return createRestlerCollection(cCtx)
				},
			},
			{
				Name:    "create-request",
				Aliases: []string{"cr"},
				Action: func(cCtx *cli.Context) error {
					intializeCreatorAction()
					return createRequestFile(cCtx)
				},
			},
			{
				Name:    "create",
				Aliases: []string{"c"},
				Usage:   "Create restler structure",
				Subcommands: []*cli.Command{
					{
						Name:    "collection",
						Aliases: []string{"c"},
						Usage:   "Create collection",
						Action: func(cCtx *cli.Context) error {
							intializeCreatorAction()
							return createRestlerCollection(cCtx)
						},
					},
					{
						Name:    "request",
						Aliases: []string{"r"},
						Usage:   "Create request",
						Action: func(cCtx *cli.Context) error {
							intializeCreatorAction()
							return createRequestFile(cCtx)
						},
					},
				},
			},
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "Run request",
				Flags:   commonCommandFlags,
				Action: func(cCtx *cli.Context) error {
					initConfigs(cCtx)
					return runAction(cCtx)
				},
			},
			// @depreciated Use run command restler run
			{
				Name:    "post",
				Aliases: []string{"p"},
				Usage:   "Run post request",
				Flags:   commonCommandFlags,
				Action: func(cCtx *cli.Context) error {
					initialize(cCtx)
					return restAction(cCtx, POST, restlerPath)
				},
			},
			// @depreciated Use run command restler run
			{
				Name:    "get",
				Aliases: []string{"g"},
				Usage:   "Run get request",
				Flags:   commonCommandFlags,
				Action: func(cCtx *cli.Context) error {
					initialize(cCtx)
					return restAction(cCtx, GET, restlerPath)
				},
			},
			// @depreciated Use run command restler run
			{
				Name:    "put",
				Aliases: []string{"u"},
				Usage:   "Run put request",
				Flags:   commonCommandFlags,
				Action: func(cCtx *cli.Context) error {
					initialize(cCtx)
					return restAction(cCtx, PUT, restlerPath)
				},
			},
			// @depreciated Use run command restler run
			{
				Name:    "delete",
				Aliases: []string{"d"},
				Usage:   "Run delete request",
				Flags:   commonCommandFlags,
				Action: func(cCtx *cli.Context) error {
					initialize(cCtx)
					return restAction(cCtx, DELETE, restlerPath)
				},
			},
			// @depreciated Use run command restler run
			{
				Name:    "patch",
				Aliases: []string{"m"},
				Usage:   "Run patch request",
				Flags:   commonCommandFlags,
				Action: func(cCtx *cli.Context) error {
					initialize(cCtx)
					return restAction(cCtx, PATCH, restlerPath)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

// -------------------------
// Restler create command
// -------------------------
// +++++++++++++++++++++++++
// create collection
// +++++++++++++++++++++++++
func createRestlerCollection(c *cli.Context) error {
	fmt.Println("[Restler Log]: Creating restler collection", c.Args().First())
	// collection is basically restler structure
	// it will have env folder, config.yaml and sample request
	collectionPath := fmt.Sprintf("%s/%s", restlerPath, c.Args().First())
	if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
		err := os.MkdirAll(collectionPath, 0755)
		if err != nil {
			return fmt.Errorf("[error]: Error occurred while creating restler collection: err: %s", err)
		}
		err = createDefaultFiles(collectionPath)
		if err != nil {
			fmt.Println("[error]: Error occurred while creating default files: ", err)
			return err
		}
		return nil
	} else {
		fmt.Println("[info]: path exists, ignoring create restler collection")
	}
	return nil
}

// +++++++++++++++++++++++++
// create request file
// +++++++++++++++++++++++++
func createRequestFile(c *cli.Context) error {

	// user should be able to create request file with action name like `restler c r post`
	// or they can create request file with path like `restler c r collection1/collection2 post`
	path := restlerPath
	var action string
	var fileName string
	argsLen := c.Args().Len()
	// if no args provided, return error
	if argsLen == 0 {
		return errors.New("action name is required (post, get, put, delete, patch)")
	}
	// if only one arg provided, it should be action name and it should create sample request file
	// with action provided eg. <restler_path>/sample.post.yaml
	if argsLen == 1 {
		action = c.Args().Get(0)
	}

	// if two args provided, it should be path and action name
	if argsLen == 2 {
		path = fmt.Sprintf("%s/%s", restlerPath, c.Args().Get(0))
		action = c.Args().Get(1)
	}

	// if three args provided, it should be path, action name and file name
	if argsLen == 3 {
		path = fmt.Sprintf("%s/%s", restlerPath, c.Args().Get(0))
		action = c.Args().Get(1)
		fileName = c.Args().Get(2)
	}

	// // if restler cr col1/col2 post article
	if strings.Contains(action, "/") {
		path = fmt.Sprintf("%s/%s", restlerPath, action)
		action = c.Args().Get(1)
		if action == "" {
			return errors.New("action name is required (post, get, put, delete, patch)")
		}
		fileName = c.Args().Get(2)
	}

	if fileName == "" {
		fileName = "sample"
	}

	var url string
	switch action {
	case "post":
		url = samplePostRequestUrl
		return createSampleRequestFile(path, url, fmt.Sprintf("%s.post.yaml", fileName))
	case "get":
		url = sampleGetRequestUrl
		return createSampleRequestFile(path, url, fmt.Sprintf("%s.get.yaml", fileName))
	case "put":
		url = samplePutRequestUrl
		return createSampleRequestFile(path, url, fmt.Sprintf("%s.put.yaml", fileName))
	case "delete":
		url = sampleDeleteRequestUrl
		return createSampleRequestFile(path, url, fmt.Sprintf("%s.delete.yaml", fileName))
	case "patch":
		url = samplePatchRequestUrl
		return createSampleRequestFile(path, url, fmt.Sprintf("%s.patch.yaml", fileName))
	default:
		return errors.New("invalid action name, please use one of post, get, put, delete, patch")
	}
}

func createSampleRequestFile(path string, url string, fileName string) error {
	// fetch sample request file from github and save it to path
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get file content from %s, status code: %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("%s/%s", path, fileName)
	err = os.WriteFile(filePath, body, 0644)
	if err != nil {
		return err
	}

	return nil
}

func intializeRestlerPath() {
	restlerPath = os.Getenv("RESTLER_PATH")
	if restlerPath == "" {
		fmt.Println("[restler Log]:RESTLER_PATH is not set, defaulting to restler")
		restlerPath = "restler"
	}
}

// Load proxy from env supports both HTTPS_PROXY and HTTP_PROXY
func intializeProxy() {
	gProxyUrl = os.Getenv("HTTPS_PROXY")
	if gProxyUrl == "" {
		gProxyUrl = os.Getenv("HTTP_PROXY")
		if gProxyUrl == "" {
			gProxyUrl = ""
		}
	}
}

func intializeCreatorAction() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("[Restler Log]: Error loading .env file: ", err)
	}
	intializeRestlerPath()
}

// -------------------------

// -------------------------
// initialize restler project
// -------------------------
func initialize(c *cli.Context) {
	// RESTLER_PATH path, where to run command to create api request
	err := godotenv.Load()
	if err != nil {
		fmt.Println("[Restler Log]: Error loading .env file: ", err)
	}
	intializeRestlerPath()
	intializeProxy()

	_, reqPath := getReqNamePath(c.Args().First())
	// Load config from current request collection
	err = loadWithYaml(fmt.Sprintf("%s/config.yaml", reqPath), &config)
	if err != nil {
		fmt.Println("[restler log]:Failed to load config file, using default env, err:", err)
		config.DefaultConfig()
	}

	// Load Environment
	// TODO: should be able to take env for nested structure
	// May be merge env from parent folder
	// For now we will be using env from the current request collection
	err = loadWithYaml(fmt.Sprintf("%s/env/%s.yaml", reqPath, config.Env), &env)
	if err != nil {
		fmt.Printf("[restler Error]: Failed to load environment file! Make sure you have at least default.yaml file in %s/env folder to use environment variables in request!\n", restlerPath)
	}
}

func findFileRecursively(startPath string, fileName string) (string, error) {
	// get the absolute path of the starting dir
	currentPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working dir: %w", err)
	}
	pwd = filepath.Clean(pwd)

	for {
		currentDir := filepath.Dir(currentPath)
		filePath := filepath.Join(currentDir, fileName)
		if _, err := os.Stat(filePath); err == nil {
			return filePath, nil
		}

		if currentPath == pwd || currentPath == filepath.Dir(filePath) {
			break
		}

		currentPath = filepath.Dir(filePath)
	}

	return "", fmt.Errorf("%s not found", fileName)
}

func initEnv() error {
	mainEnv.LoadEnv()
	// parser.Parse()
	return nil
}

// initConfigs function is responsible for loading config and environments
// we will merge configs and envs after this version but for now we should load both
// recursively check for the config file, and its env folder
func initConfigs(c *cli.Context) {
	mainEnv.LoadEnv()
	intializeProxy()
	reqPath := c.Args().First()
	configPath, err := findFileRecursively(reqPath, "config.yaml")
	err = loadWithYaml(configPath, &config)

	if err != nil {
		fmt.Println("[restler log]:Failed to load config file, using default env, err:", err)
		config.DefaultConfig()
	} else {
		config.configPath = configPath
	}

	// TODO: remove env folder
	envPath := fmt.Sprintf("%s/env/%s.yaml", filepath.Dir(configPath), config.Env)
	err = loadWithYaml(envPath, &env)
	if err != nil {
		fmt.Printf("[restler Error]: Failed to load environment file! Make sure you have at least default.yaml file in %s/env folder to use environment variables in request!\n", restlerPath)
	} else {
		config.envPath = envPath
	}
}

func getReqNamePath(req string) (name string, path string) {
	if strings.Contains(req, "/") {
		_paths := strings.Split(req, "/")
		name = _paths[len(_paths)-1]
		path = fmt.Sprintf("%s/%s", restlerPath, strings.Join(_paths[:len(_paths)-1], "/"))
		return
	}

	name = req
	path = restlerPath
	return
}

// init restler project
// init command should be able to set the RESTLER_PATH in .env file which will be loaded by restler.
// it should be creating default files and folders in the path.
type textInputModel struct {
	textInput textinput.Model
	err       error
}

type (
	errMsg error
)

func initialTextInputModel() textInputModel {
	ti := textinput.New()
	ti.Placeholder = "restler"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return textInputModel{
		textInput: ti,
		err:       nil,
	}
}

func (m textInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m textInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			executeInitCommand(m.textInput.Value())
			return m, tea.Quit
		}
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m textInputModel) View() string {
	return fmt.Sprintf(
		"Where do you want to initialize your restler project? \n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}
func initRestlerProject() error {
	p := tea.NewProgram(initialTextInputModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error occurred while initializing restler project: ", err)
		return err
	}
	return nil
}

// TODO: Will download the sample folder from github repo instead of creating each one one by one
func executeInitCommand(path string) error {
	// if path exists, thats it, otherwise ask if user wants to create it
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("[log]: Path doesn't exist, creating restler project in: ", path)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			fmt.Println("[error]: Error occurred while creating restler project: ", err)
			return err
		}
		err = createDefaultFiles(path)
		if err != nil {
			fmt.Println("[error]: Error occurred while creating default files: ", err)
			return err
		}
		updateEnv(path)
		return nil
	} else {
		fmt.Println("[info]: path exists, updating RESTLER_PATH env: ")
		updateEnv(path)
		return nil
	}

}

func findDotEnvFile() (string, error) {
	dotEnvPath := ".env"
	_, err := os.Stat(dotEnvPath)
	if os.IsNotExist(err) {
		dotEnvPath = ".env.local"
		_, err := os.Stat(dotEnvPath)
		if os.IsNotExist(err) {
			return "", err
		}
	}

	return dotEnvPath, nil
}

func updateEnv(path string) error {
	dotEnvPath, err := findDotEnvFile()
	if err != nil {
		fmt.Println("[log]: env file .env or .env.local not found, creating .env file :")
		if _, err := os.Create(".env"); err != nil {
			fmt.Println("[error]: Error occurred while creating .env file: ", err)
			return err
		}
		dotEnvPath = ".env"
	}

	return updateDotEnvFile(dotEnvPath, path)
}

func updateDotEnvFile(envPath, restlerPath string) error {
	file, err := os.Open(envPath)
	if err != nil {
		fmt.Println("[error]: Error occurred while opening .env file: ", err)
		return err
	}
	defer file.Close()

	// read the file line by line
	scanner := bufio.NewScanner(file)
	var lines []string
	var restlerPathFound bool
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "RESTLER_PATH") {
			// update RESTLER_PATH in .env file
			restlerPathFound = true
			line = fmt.Sprintf("RESTLER_PATH=%s", restlerPath)
		}
		lines = append(lines, line)
	}

	if !restlerPathFound {
		lines = append(lines, "RESTLER_PATH="+restlerPath)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("[error]: Error occurred while reading %s file: %s\n", envPath, err)
		return err
	}

	err = os.WriteFile(envPath, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		fmt.Printf("[error]: Error occurred while updating %s file: %s\n", envPath, err)
		return err
	}
	return nil
}

func createDefaultFiles(path string) error {
	// create config file
	configPath := fmt.Sprintf("%s/config.yaml", path)
	configFile, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer configFile.Close()

	// write default environment default to config file
	configFileContent := "Env: default"
	_, err = configFile.WriteString(configFileContent)
	if err != nil {
		return err
	}

	// create env folder
	envPath := fmt.Sprintf("%s/env", path)
	err = os.Mkdir(envPath, 0755)
	if err != nil {
		return err
	}

	// create default.yaml file in env folder
	defaultFile, err := os.Create(fmt.Sprintf("%s/default.yaml", envPath))
	if err != nil {
		return err
	}
	defer defaultFile.Close()

	// default file content on env/default.yaml
	// TODO: Update from the sample file
	defaultFileContent := "API_URL: https://jsonplaceholder.typicode.com/posts"
	_, err = defaultFile.WriteString(defaultFileContent)
	if err != nil {
		return err
	}

	// create requests folder with sample request
	requestsPath := fmt.Sprintf("%s/sample", path)
	err = os.MkdirAll(requestsPath, 0755)
	if err != nil {
		return err
	}

	// create sample request
	sampleRequestPath := fmt.Sprintf("%s/sample.post.yaml", requestsPath)
	sampleRequestFile, err := os.Create(sampleRequestPath)
	if err != nil {
		return err
	}
	defer sampleRequestFile.Close()

	sampleRequestFileContent, _ := getFileContent(samplePostRequestUrl)
	_, err = sampleRequestFile.WriteString(sampleRequestFileContent)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(".gitignore", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("[error]: Error occurred while opening .gitignore file: ", err)
		return err
	}
	defer file.Close()

	// TODO: update only if its not available on the .gitignore
	fileContent := "# Ignore response file\n**/.*.res.md\n\n# Ignore .env file\n.env\n.env.local\n"
	_, err = file.WriteString(fileContent)
	if err != nil {
		fmt.Println("[error]: Error occurred while writing to .gitignore file: ", err)
		return err
	}

	return nil
}

func getFileContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get file content from %s, status code: %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// INIT functionality ends here

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

	_, reqPath := getReqNamePath(c.Args().First())
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
