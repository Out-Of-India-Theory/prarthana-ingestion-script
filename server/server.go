package server

import (
	"context"
	"fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/app"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/configuration"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/repository/es/prarthana"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/repository/mongo/prarthana_data"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/deity_ingestion"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/facade"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/prarthana_ingestion"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/search_ingestion"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/shlok_ingestion"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/shlok_translation"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/stotra_ingestion"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/service/zoho"
	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent/v3/newrelic"
	"net/http"
)

func InitServer(ctx context.Context, app *app.App, configuration *configuration.Configuration) {
	//repo initializations
	prarthanaDataMongoRepository := prarthana_data.InitPrarthanaDataMongoRepository(ctx, *configuration)
	prarthanaESRepository := prarthana.InitPrarthanaESRepository(ctx, configuration.ESConfig)

	zohoService := zoho.InitZohoService(ctx, configuration, &http.Client{})
	//service initializations
	shlokIngestionService := shlok_ingestion.InitShlokIngestionService(ctx, prarthanaDataMongoRepository, zohoService)
	stotraIngestionService := stotra_ingestion.InitStotraIngestionService(ctx, prarthanaDataMongoRepository, zohoService)
	prarthanaIngestionService := prarthana_ingestion.InitPrathanaIngestionService(ctx, prarthanaDataMongoRepository, zohoService)
	deityIngestionService := deity_ingestion.InitDeityIngestionService(ctx, prarthanaDataMongoRepository, zohoService)
	searchIngestionService := search_ingestion.InitSearchIngestionService(ctx, prarthanaDataMongoRepository, prarthanaESRepository)
	shlokTranslationService := shlok_translation.InitShlokTranslationService(ctx, zohoService)

	facadeService := facade.InitFacadeService(ctx, configuration, shlokIngestionService, stotraIngestionService, prarthanaIngestionService, deityIngestionService, zohoService, searchIngestionService, shlokTranslationService)
	registerMiddleware(app, configuration)
	registerRoutes(ctx, app, facadeService, configuration)

	app.StartHttpServer()
	err := app.StartMetricsServer()
	if err != nil {
		panic("Error while initializing http client")
	}

	<-make(chan int)
}

func registerMiddleware(app *app.App, configuration *configuration.Configuration) {
	newrelicApp, err := newrelic.NewApplication(
		newrelic.ConfigAppName(fmt.Sprintf("%s-%s", app.Config.AppName, app.Config.Env)),
		newrelic.ConfigLicense(app.Config.NewRelicLicense),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
	if err != nil {
		fmt.Println("Error while initializing new relic app")
		return
	}
	//app.Engine.Use(nrgin.Middleware(newrelicApp))
	app.Engine.Use(newrelicTransactionMiddleware(newrelicApp))
	app.Engine.Use(CORSMiddleware())
}

func newrelicTransactionMiddleware(newRelicApp *newrelic.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, "newRelicTransaction", newrelic.FromContext(c))
		c.Request = c.Request.Clone(ctx)
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, source")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
		c.Writer.Header().Set("source", "*")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
