package ingestion

import (
	"context"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/configuration"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/entity"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/service/facade"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/util"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

type Controller struct {
	logger  *zap.Logger
	service facade.Service
	config  *configuration.Configuration
}

func InitIngestionController(ctx context.Context, service facade.Service, config *configuration.Configuration) *Controller {
	return &Controller{
		logger:  logging.WithContext(ctx),
		service: service,
		config:  config,
	}
}

func (con *Controller) ShlokIngestion(c *gin.Context) {
	ctx := c.Request.Context()
	ctx = util.SetZohoAccessTokenInContext(ctx, c.Request.Header.Get("zoho-access-token"))
	var request entity.IngestionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid request payload",
		})
		return
	}
	err := con.service.ShlokIngestionService().ShlokIngestion(ctx, request.StartID, request.EndID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Error processing request: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Successful",
		"data":    nil,
	})
}

func (con *Controller) StotraIngestion(c *gin.Context) {
	ctx := c.Request.Context()
	ctx = util.SetZohoAccessTokenInContext(ctx, c.Request.Header.Get("zoho-access-token"))
	var request entity.IngestionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid request payload",
		})
		return
	}
	_, err := con.service.StotraIngestionService().StotraIngestion(ctx, request.StartID, request.EndID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Error processing request: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Successful",
		"data":    nil,
	})
}

func (con *Controller) PrarthanaIngestion(c *gin.Context) {
	ctx := c.Request.Context()
	ctx = util.SetZohoAccessTokenInContext(ctx, c.Request.Header.Get("zoho-access-token"))
	var requestBody entity.IngestionRequest
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid request payload",
		})
		return
	}
	_, err := con.service.PrarthanaIngestionService().PrarthanaIngestion(ctx, requestBody.StartID, requestBody.EndID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Error processing request: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Successful",
		"data":    nil,
	})
}

func (con *Controller) DeityIngestion(c *gin.Context) {
	ctx := c.Request.Context()
	var requestBody struct {
		DeityCsvFilePath            string `json:"deity_csv_file_path"`
		PrarthanaToDeityCsvFilePath string `json:"prarthana_to_deity_csv_file_path"`
		StartID                     int    `json:"start_id"`
		EndID                       int    `json:"end_id"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid request payload",
		})
		return
	}
	_, err := con.service.DeityIngestionService().DeityIngestion(ctx, requestBody.PrarthanaToDeityCsvFilePath, requestBody.DeityCsvFilePath, requestBody.StartID, requestBody.EndID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Error processing request: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Successful",
		"data":    nil,
	})
}
