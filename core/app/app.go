package app

import "time"

type Config struct {
	Env     string `yaml:"Env"`
	EnvPath string `yaml:"EnvPath"`
}

type App struct {
	ProxyUrl    string
	Version     string
	Config      *Config
	RequestTime time.Duration
}

func NewApp(proxyUrl string, version string, config *Config) *App {
	return &App{
		ProxyUrl: proxyUrl,
		Version:  version,
		Config:   config,
	}
}

func (a *App) Terminate() {
	a = nil
}
