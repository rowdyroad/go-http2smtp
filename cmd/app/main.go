package main

import (
	yconfig "github.com/rowdyroad/go-yaml-config"
	app "go-http2smtp/pkg/app"
)

func main() {
	var cfg app.Config
	yconfig.LoadConfig(&cfg, "configs/app.yaml", nil)
	app := app.NewApp(cfg)
	defer app.Close()
	app.Run()
}
