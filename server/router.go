package server

import (
	"context"
	"github.com/Out-Of-India-Theory/oit-go-commons/app"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/configuration"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/controller/ingestion"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/middleware"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/facade"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/zoho"
	"github.com/gin-gonic/gin"
	"net/http"
)

func registerRoutes(ctx context.Context, app *app.App, service facade.Service, configuration *configuration.Configuration) {
	basePath := app.Engine.Group("prarthana_script")
	app.Engine.GET("/health-check", ingestion.HealthCheck)
	authRepo := zoho.InitZohoService(ctx, configuration, &http.Client{})
	am := middleware.InitAuthMiddleware(configuration, authRepo)
	//prarthana-script
	{
		prarthanaIngestionController := ingestion.InitIngestionController(ctx, service, configuration)
		prarthanaIngestionV1 := basePath.Group("v1")
		prarthanaIngestionV1.POST("/shloks", am.ZohoAuthMiddleware(), prarthanaIngestionController.ShlokIngestion)
		prarthanaIngestionV1.POST("/stotras", am.ZohoAuthMiddleware(), prarthanaIngestionController.StotraIngestion)
		prarthanaIngestionV1.POST("/prarthanas", am.ZohoAuthMiddleware(), prarthanaIngestionController.PrarthanaIngestion)
		prarthanaIngestionV1.POST("/deities", am.ZohoAuthMiddleware(), prarthanaIngestionController.DeityIngestion)
		prarthanaIngestionV1.GET("/deities-search", prarthanaIngestionController.DeitySearchIngestion)
		prarthanaIngestionV1.GET("/prarthanas-search", prarthanaIngestionController.PrarthanaSearchIngestion)
	}
	app.Engine.LoadHTMLGlob("ingestion/*.html")
	app.Engine.GET("/ingestion/prarthana.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", configuration.UIConfig)
	})
}
