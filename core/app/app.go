package app

type Config struct {
	Env     string `yaml:"Env"`
	EnvPath string `yaml:"EnvPath"`
}

type App struct {
	ProxyUrl string
	Version  string
	Config   *Config
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
