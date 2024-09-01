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
	"os"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

func main() {
	// RESTLER_PATH path, where to run command to create api request
	var rsClientPath = os.Getenv("RESTLER_PATH")
	if rsClientPath == "" {
		rsClientPath = "./example/restlers" // TODO: Remove this after testing
	}

	app := &cli.App{
		Name:  "RSClient Application",
		Usage: "RestClient for developers",
		Action: func(cCtx *cli.Context) error {
			// consumer will send the request name in the args
			// TODO: if not we will run all requests on dir
			var requestDir = cCtx.Args().Get(0)
			if requestDir == "" {
				requestDir = "some"
			}
			// request path will have .request.yaml if not we will say that request file not found
			var requestDirPath = fmt.Sprintf("%s/requests/%s", rsClientPath, requestDir)
			if _, err := os.Stat(requestDirPath); os.IsNotExist(err) {
				log.Fatal("Request directory not found, please check the path")
			}

			var requestPath = fmt.Sprintf("%s/%s.request.yaml", requestDirPath, requestDir)
			fmt.Println("Processing Request: ",requestPath)
			if _, err := os.Stat(requestPath); os.IsNotExist(err) {
				log.Fatal("Request file not found, please check the path")
			}
			res, err := parseRequest(requestPath);
			if err != nil {
				log.Fatal(err)
			}
			body, err := readBody(res)
			if err != nil {
				log.Fatal(err)
			}

			responseBytes, err := prepareResponse(res, body);
			if err != nil {
				log.Fatal(err)
			}
		
			// TODO: support different output formats
			// If user chooses json format, we should convert the response body to json and write in different file than header.
			outputFilePath := fmt.Sprintf("%s/.%s.response.txt", requestDirPath, requestDir)
			os.WriteFile(outputFilePath, responseBytes, 0644)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

func prepareResponse(res *http.Response, body []byte)([]byte, error){
	var buffer bytes.Buffer
	buffer.WriteString("-------Header--------\n")

	for key, value := range res.Header {
		buffer.WriteString(fmt.Sprintf("%s: %s\n", key, value))
	}
	buffer.WriteString("\n")

	buffer.WriteString("-------Body--------\n")
	buffer.Write(body)

	return buffer.Bytes(), nil
}


func readBody(res * http.Response)([]byte, error){
  var reader io.ReadCloser;
  var err error;

  switch res.Header.Get("Content-Encoding") {
  case "gzip":{
	reader, err = gzip.NewReader(res.Body)
	if err!= nil {
		return nil, fmt.Errorf("error creating gzip reader : %v", err)
		}
  	}
	default:{
		reader = res.Body
	}
  }

  body, err := io.ReadAll(reader)
  if err != nil {
	return nil, fmt.Errorf("error reading response body : %v", err)
  }

  return body, nil
}

type Request struct {
	Name    string            `yaml:"Name"`
	URL     string            `yaml:"URL"`
	Method  string            `yaml:"Method"`
	Env     string            `yaml:"Env"`
	Headers map[string]string `yaml:"Headers"`
	Body    interface{}       `yaml:"Body"`
}

func parseRequest(requestPath string) (*http.Response, error) {
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

	var request Request
	err = yaml.Unmarshal(rawRequestContent, &request)
	if err != nil {
		fmt.Println("Error parsing request file, please check the format", err)
		return nil, err
	}
	err = validateRequest(&request)
	if err != nil {
		fmt.Println("Error validating request file, Error: ", err)
		return nil, err
	}

	res, err := processRequest(&request)
	if err != nil {
		fmt.Println("Error processing request: ", err)
		return nil, err
	}

	return res, nil
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
	// TODO: Add default env rather than making it required
	if r.Env == "" {
		return errors.New("Request Env is required")
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
	client := &http.Client{}
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
		return nil,fmt.Errorf("error making http request %s", err)
	}
	return httpResp, nil
}

