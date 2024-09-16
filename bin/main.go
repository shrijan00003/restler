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
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

type Request struct {
	Name       string            `yaml:"Name"`
	URL        string            `yaml:"URL"`
	Method     string            `yaml:"Method"`
	Headers    map[string]string `yaml:"Headers"`
	Body       interface{}       `yaml:"Body"`
	After      *After            `yaml:"After"`	
}

type After struct{
	Env map[string]string `yaml:"Env"`
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
const APP_VERSION = "v0.0.1-dev.6"

var restlerPath string



func main() {
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
				Name: "init",
				Aliases: []string{"i"},
				Usage: "Initialize restler project",
				Action: func(cCtx *cli.Context) error {
					return initRestlerProject()
				},
			},
			{
				Name:    "post",
				Aliases: []string{"p"},
				Usage:   "Run post request",
				Flags:   commonCommandFlags,
				Action: func(cCtx *cli.Context) error {
					initialize()
					return restAction(cCtx, POST, restlerPath)
				},
			},
			{
				Name:    "get",
				Aliases: []string{"g"},
				Usage:   "Run get request",
				Flags:   commonCommandFlags,
				Action: func(cCtx *cli.Context) error {
					initialize()
					return restAction(cCtx, GET, restlerPath)
				},
			},
			{
				Name:    "put",
				Aliases: []string{"u"},
				Usage:   "Run put request",
				Flags:   commonCommandFlags,
				Action: func(cCtx *cli.Context) error {
					initialize()
					return restAction(cCtx, PUT, restlerPath)
				},
			},
			{
				Name:    "delete",
				Aliases: []string{"d"},
				Usage:   "Run delete request",
				Flags:   commonCommandFlags,
				Action: func(cCtx *cli.Context) error {
					initialize()
					return restAction(cCtx, DELETE, restlerPath)
				},
			},
			{
				Name:    "patch",
				Aliases: []string{"m"},
				Usage:   "Run patch request",
				Flags:   commonCommandFlags,
				Action: func(cCtx *cli.Context) error {
					initialize()
					return restAction(cCtx, PATCH, restlerPath)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

// initialize restler project
func initialize(){
	// RESTLER_PATH path, where to run command to create api request
	err := godotenv.Load()
	if err != nil {
		fmt.Println("[Restler Log]: Error loading .env file: ", err)
	}
	restlerPath = os.Getenv("RESTLER_PATH")
	if restlerPath == "" {
		fmt.Println("[restler Log]:RESTLER_PATH is not set, defaulting to restler")
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
	err = loadWithYaml(fmt.Sprintf("%s/config.yaml", restlerPath), &config)
	if err != nil {
		fmt.Println("[restler log]:Failed to load config file, using default env, err:", err)
		config.DefaultConfig()
	}

	// Load Environment
	err = loadWithYaml(fmt.Sprintf("%s/env/%s.yaml", restlerPath, config.Env), &env)
	if err != nil {
		fmt.Printf("[restler Error]: Failed to load environment file! Make sure you have at least default.yaml file in %s/env folder to use environment variables in request!\n", restlerPath)
	}
}

// init restler project
// init command should be able to set the RESTLER_PATH in .env file which will be loaded by restler. 
// it should be creating default files and folders in the path. 
type textInputModel struct{
	textInput textinput.Model
	err error
	message string
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
	p:= tea.NewProgram(initialTextInputModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error occurred while initializing restler project: ", err)
		return err;
	}
	return nil;
}
// TODO: Will download the sample folder from github repo instead of creating each one one by one
func executeInitCommand(path string) error {
	// if path exists, thats it, otherwise ask if user wants to create it 
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("[log]: Path doesn't exist, creating restler project in: ", path)
		err:= os.MkdirAll(path, 0755)
		if err != nil {
			fmt.Println("[error]: Error occurred while creating restler project: ", err)
			return err;
		}
		err = createDefaultFiles(path)
		if err != nil {
			fmt.Println("[error]: Error occurred while creating default files: ", err)
			return err;
		}
		updateEnv(path);
		return nil;
	}else{
		fmt.Println("[info]: path exists, updating RESTLER_PATH env: ")
		updateEnv(path);
		return nil;
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
		if _, err:= os.Create(".env") ; err!= nil {
			fmt.Println("[error]: Error occurred while creating .env file: ", err)
			return err
		}
		dotEnvPath = ".env"
	}

	return updateDotEnvFile(dotEnvPath, path)
}

func updateDotEnvFile(envPath, restlerPath string) error {
   file,err := os.Open(envPath)
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
	return nil;
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
	defaultFileContent := "API_URL: https://jsonplaceholder.typicode.com/posts"
	_, err = defaultFile.WriteString(defaultFileContent)
	if err != nil {
		return err
	}
	
	// create requests folder with sample request
	requestsPath := fmt.Sprintf("%s/requests/sample", path)
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
	// TODO: Write sample post request command to this file
	// read content from the github repo and write to the file
	// https://raw.githubusercontent.com/shrijan00003/restler/main/restler/requests/posts/posts.post.yaml
	sampleRequestFileContent, _ := getFileContent("https://raw.githubusercontent.com/shrijan00003/restler/main/restler/requests/posts/posts.post.yaml")
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
	fileContent := "# Ignore response file\n**/.*.res.md\n\n# Ignore .env file\n.env\n.env.local\n"
	_, err = file.WriteString(fileContent)
	if err != nil {
		fmt.Println("[error]: Error occurred while writing to .gitignore file: ", err)
		return err
	}

	return nil;
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

func restAction(cCtx *cli.Context, actionName ActionName, restlerPath string) error {
	var req = cCtx.Args().Get(0)
	if req == "" {
		log.Fatal("[Restler Error]: No request provided! Please provide request name as argument. Request name is the name of the folder in requests folder.")
	}

	// update env if env flag is set
	envFlag := cCtx.String("env")
	if envFlag != "" {
		config.Env = envFlag
		err := loadWithYaml(fmt.Sprintf("%s/env/%s.yaml", restlerPath, envFlag), &env)
		if err != nil {
			return fmt.Errorf("restler >>[Error]: Environment you have selected is not found in %s/env folder", restlerPath)
		}
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

	updateEnvPostScript(pReq, pRes, body)

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


func updateEnvPostScript(req *Request, res *http.Response, body []byte){
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
				return;
			}
			if val, ok := getNestedValue(jsonBody, envBodyKey); ok {
				envBodyValueMap[envKey] = fmt.Sprintf("%v", val)
			}else{
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

	envPath := fmt.Sprintf("%s/env/%s.yaml", restlerPath, config.Env)
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
