package parser

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type EnvConfig struct {
	URL     string            `yaml:"URL"`
	Headers map[string]string `yaml:"Headers"`
	Body    map[string]any    `yaml:"Body"`
	After   map[string]any    `yaml:"After"`
}

func Parse() {
	raw, err := os.ReadFile("./core/parser/config.yaml")
	if err != nil {
		panic(err)
	}

	replraced := os.ExpandEnv(string(raw))
	config := &EnvConfig{}

	err = yaml.Unmarshal([]byte(replraced), config)
	if err != nil {
		panic(err)
	}

	fmt.Printf("URL: %v\n", config.URL)
	fmt.Printf("Headers: %v\n", config.Headers)
	fmt.Printf("Body: %v\n", config.Body)
	fmt.Printf("After: %v\n", config.After)
	fmt.Println(config.After)
}
