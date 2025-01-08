package server

import (
	"context"
	"github.com/Out-Of-India-Theory/oit-go-commons/app"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/configuration"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/controller"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/service/facade"
)

func registerRoutes(ctx context.Context, app *app.App, service facade.Service, configuration *configuration.Configuration) {
	basePath := app.Engine.Group("prarthana_script")
	app.Engine.GET("/health-check", controller.HealthCheck)

	//prarthana-script
	{
		prarthanaIngestionController := controller.InitIngestionController(ctx, service, configuration)
		prarthanaIngestionV1 := basePath.Group("v1")
		prarthanaIngestionV1.POST("/shloks", prarthanaIngestionController.ShlokIngestion)
		prarthanaIngestionV1.POST("/stotras", prarthanaIngestionController.StotraIngestion)
		prarthanaIngestionV1.POST("/prarthanas", prarthanaIngestionController.PrarthanaIngestion)
		prarthanaIngestionV1.POST("/deities", prarthanaIngestionController.DeityIngestion)
	}
}
