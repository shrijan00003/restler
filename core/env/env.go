package env

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/shrijan00003/restler/core/app"
	"github.com/shrijan00003/restler/core/logger"
	"github.com/shrijan00003/restler/core/utils"
)

func Terminate() {
	// unset environment variables if possible
}

func LoadEnv(config *app.Config) {
	load(config)

	// TODO: env side effects (need to find another place for this)
	logLevel := os.Getenv("RESTLER_LOG_LEVEL")
	if logLevel == "DEBUG" {
		logger.SetDebug()
	}
}

// LoadDefaultEnv function will load env files and set env variables
// May be we should be limited to .env and .env.local to make it simpler and faster.
func load(config *app.Config) {
	envFilePatterns := []string{config.EnvPath, filepath.Join(utils.Pwd(), ".env."+config.Env), ".env.local", ".env", "~/.env.local", "~/.env"}
	var envFiles []string

	for _, pattern := range envFilePatterns {
		matchedFiles, err := filepath.Glob(os.ExpandEnv(pattern))
		if err != nil {
			fmt.Println("[Restler Log]: Error loading .env file: ", err)
		}
		envFiles = append(envFiles, matchedFiles...)
	}

	err := godotenv.Load(envFiles...)
	if err != nil {
		fmt.Println("[restler info]: Error loading .env file: ", err)
	}

}

func GetCurrentEnvPath(a *app.App) string {
	if a.Config.EnvPath != "" || a.Config.Env != "" {
		if a.Config.EnvPath != "" {
			return a.Config.EnvPath
		}
		return filepath.Join(utils.Pwd(), ".env."+a.Config.Env)
	}

	if _, err := os.Stat(".env.local"); err == nil {
		return ".env.local"
	}

	if _, err := os.Stat(".env"); err == nil {
		return ".env"
	}

	return ""
}

// LoadEnvFileByName function will load env file by name
// dev will load .env.dev file from the current directory
func LoadEnvFileByName(envName string) error {
	envPath := filepath.Join(utils.Pwd(), ".env."+envName)
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		logger.Debug("[restler error]: env file does not exist", "[error]", err)
		return err
	}

	return OverLoadEnv(envPath)
}

func OverLoadEnv(envPath string) error {
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		logger.Debug("[restler error]: env file does not exist", "[error]", err)
		return err
	}
	err := godotenv.Overload(envPath)
	if err != nil {
		logger.Debug("[restler error]: env file can not be loaded", "[error]", err)
	}
	return nil
}

func GetEnv(key string) string {
	return os.Getenv(key)
}

func UpdateEnv(path string) error {
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

func UpdateEnvFile(a *app.App, values map[string]interface{}) error {
	envPath := GetCurrentEnvPath(a)
	if envPath == "" {
		fmt.Println("[error]: Env file path not found")
		return fmt.Errorf("env file path not found")
	}

	file, err := os.Open(envPath)
	if err != nil {
		fmt.Println("[error]: Error occurred while opening .env file: ", err)
		return err
	}
	defer file.Close()

	// read the file line by line
	scanner := bufio.NewScanner(file)
	var lines []string
	var updatedEnvMap = make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()
		for key, value := range values {
			if strings.HasPrefix(line, key) {
				line = fmt.Sprintf("%s=%s", key, value)
				updatedEnvMap[key] = value.(string)
			}
		}
		lines = append(lines, line)
	}

	for key, value := range values {
		if _, ok := updatedEnvMap[key]; !ok {
			lines = append(lines, fmt.Sprintf("%s=%s", key, value))
		}
	}

	err = os.WriteFile(envPath, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		fmt.Printf("[error]: Error occurred while updating %s file: %s\n", envPath, err)
		return err
	}
	return nil
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
