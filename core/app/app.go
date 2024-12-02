package app

type App struct {
	ProxyUrl string
	Version  string
}

func NewApp(proxyUrl string, version string) *App {
	return &App{
		ProxyUrl: proxyUrl,
		Version:  version,
	}
}

func Terminate() {
	// TODO: implement
}
