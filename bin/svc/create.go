package svc

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

const samplePostRequestUrl = "https://raw.githubusercontent.com/shrijan00003/restler/main/sample/requests/sample.post.yaml"
const sampleGetRequestUrl = "https://raw.githubusercontent.com/shrijan00003/restler/main/sample/requests/sample.get.yaml"
const samplePutRequestUrl = "https://raw.githubusercontent.com/shrijan00003/restler/main/sample/requests/sample.put.yaml"
const sampleDeleteRequestUrl = "https://raw.githubusercontent.com/shrijan00003/restler/main/sample/requests/sample.delete.yaml"
const samplePatchRequestUrl = "https://raw.githubusercontent.com/shrijan00003/restler/main/sample/requests/sample.patch.yaml"

// FIXME: this will not be needed
var restlerPath = os.Getenv("RESTLER_PATH")

// -------------------------
// Restler create command
// -------------------------
// +++++++++++++++++++++++++
// create collection
// +++++++++++++++++++++++++
func createRestlerCollection(c *cli.Context) error {
	restlerPath := os.Getenv("RESTLER_PATH")
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

var CreateDefaultFile = createDefaultFiles

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
