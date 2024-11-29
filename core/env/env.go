package env

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/shrijan00003/restler/core/logger"
)

const RESTLER_ENV = "RESTLER_ENV"

func Init() {
	LoadDefaultEnv()

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
	// Init logger in main file
	logger.Init()
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
