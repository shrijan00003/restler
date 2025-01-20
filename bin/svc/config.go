package svc

import (
	"github.com/shrijan00003/restler/core/app"
	"github.com/shrijan00003/restler/core/utils"
)

func LoadConfig() (*app.Config, error) {
	config := &app.Config{}
	configPath := utils.Pwd() + "/" + "config.yaml"
	err := utils.LoadWithYaml(configPath, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
