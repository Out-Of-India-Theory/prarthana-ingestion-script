package main

import (
	"context"
	"fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/app"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/configuration"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/server"
)

func main() {
	configuration := configuration.GetConfig()
	ctx := context.Background()
	App, err := app.NewApp(ctx, configuration.ServerConfig)
	if err != nil {
		panic(fmt.Sprintf("Unable to initialize the app : %v", err))
	}
	server.InitServer(ctx, App, configuration)
}
