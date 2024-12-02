package env

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/shrijan00003/restler/core/logger"
)

const RESTLER_ENV = "RESTLER_ENV"

func Init() {
	LoadDefaultEnv()

	// TODO: env side effects (need to find another place for this)
	logLevel := os.Getenv("RESTLER_LOG_LEVEL")
	if logLevel == "DEBUG" {
		logger.SetDebug()
	}
}

func Terminate() {
	// unset environment variables if possible
}

func GetCurrentEnvPath() string {
	currentEnv := os.Getenv(RESTLER_ENV)
	restlerEnvDir, err := os.Getwd()

	if err != nil {
		logger.Debug("[restler error]", "error getting current directory", err)
	}
	restlerEnvPath := restlerEnvDir + "/.restler/.env"
	if currentEnv != "" {
		restlerEnvPath = restlerEnvDir + "/.restler/.env." + currentEnv
	}

	if _, err := os.Stat(restlerEnvPath); os.IsNotExist(err) {
		restlerEnvPath = restlerEnvDir + ".env"
	}

	return restlerEnvPath
}

func LoadRestlerEnv() {
	Init()

	restlerEnvPath := GetCurrentEnvPath()
	err := godotenv.Overload(restlerEnvPath)
	if err != nil {
		logger.Debug("[restler error]: env file can not be loaded", "[error]", err)
	}

}

func LoadEnv() {
	LoadRestlerEnv()
	apitoken := os.Getenv("API_KEY")
	if apitoken != "" {
		logger.Info("[restler info]: API_KEY is set " + apitoken)
	} else {
		logger.Info("[restler info]: API_KEY is not set")
	}
}

// LoadDefaultEnv function will load env files and set env variables
// May be we should be limited to .env and .env.local to make it simpler and faster.
func LoadDefaultEnv() {
	envFilePatterns := []string{".env.local", ".env", "~/.env.local", "~/.env"}
	var envFiles []string

	for _, pattern := range envFilePatterns {
		matchedFiles, err := filepath.Glob(os.ExpandEnv(pattern))
		if err != nil {
			fmt.Println("[Restler Log]: Error loading .env file: ", err)
		}
		// TODO: Do we need to check if the file exists?
		envFiles = append(envFiles, matchedFiles...)
	}

	err := godotenv.Load(envFiles...)
	if err != nil {
		fmt.Println("[restler info]: Error loading .env file: ", err)
	}

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
