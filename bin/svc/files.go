package svc

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

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
