package ingestion

import (
	"context"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/configuration"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/service/facade"
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
	var requestBody struct {
		CsvFilePath string `json:"csv_file_path"`
		StartID     int    `json:"start_id"`
		EndID       int    `json:"end_id"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid request payload",
		})
		return
	}
	err := con.service.ShlokIngestionService().ShlokIngestion(ctx, requestBody.CsvFilePath, requestBody.StartID, requestBody.EndID)
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
	var requestBody struct {
		CsvFilePath string `json:"csv_file_path"`
		StartID     int    `json:"start_id"`
		EndID       int    `json:"end_id"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid request payload",
		})
		return
	}
	_, err := con.service.StotraIngestionService().StotraIngestion(ctx, requestBody.CsvFilePath, requestBody.StartID, requestBody.EndID)
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
	var requestBody struct {
		AdhyayaCsvFilePath          string `json:"adhyaya_csv_file_path"`
		VariantCsvFilePath          string `json:"variant_csv_file_path"`
		PrarthanaCsvFilePath        string `json:"prarthana_csv_file_path"`
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
	_, err := con.service.PrarthanaIngestionService().PrarthanaIngestion(ctx, requestBody.PrarthanaToDeityCsvFilePath, requestBody.AdhyayaCsvFilePath, requestBody.VariantCsvFilePath, requestBody.PrarthanaCsvFilePath, requestBody.StartID, requestBody.EndID)
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
