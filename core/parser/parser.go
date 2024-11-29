package parser

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
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

func ParseViper() {
	viper.SetConfigFile("./core/parser/config.yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	headers := viper.GetStringMap("Headers")
	body := viper.GetStringMap("Body")
	after := viper.GetStringMap("After")

	fmt.Printf("Headers: %v\n", headers)
	fmt.Printf("Body: %v\n", body)
	fmt.Printf("After: %v\n", after)

	authorization := viper.GetString("Headers.Authorization")
	fmt.Printf("Authorization: %v\n", authorization)

	fmt.Println("URL: ", viper.GetString("URL"))
}
