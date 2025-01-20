package utils

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func GetReqNamePath(req string) (name string, path string) {
	restlerPath := os.Getenv("RESTLER_PATH")

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

func WriteYAMLFile(filename string, data map[string]interface{}) error {
	out, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, out, 0644)
}

func ConvertMap[K comparable, V any](m map[K]V) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[fmt.Sprintf("%v", k)] = v
	}
	return result
}

func MergeMaps[K comparable](dest, src map[K]interface{}) {
	for key, value := range src {
		if v, ok := value.(map[K]interface{}); ok {
			if dv, ok := dest[key].(map[K]interface{}); ok {
				MergeMaps(dv, v)
				continue
			}
		}
		dest[key] = value
	}
}

func GetNestedValue(data map[string]interface{}, keys string) (interface{}, bool) {
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

func HeaderToMap(header http.Header) map[string]interface{} {
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

func ReadBody(res *http.Response) ([]byte, error) {
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

func LoadWithYaml(configPath string, receiver interface{}) error {
	contentBytes, err := ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading file at %s", configPath)
	}

	err = yaml.Unmarshal(contentBytes, receiver)
	if err != nil {
		return fmt.Errorf("error occurred on parsing Yaml file : %s", configPath)
	}

	return nil
}

func ReadFile(path string) ([]byte, error) {
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

func Pwd() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("Error getting current directory", err)
	}
	return dir
}
