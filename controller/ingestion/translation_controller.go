package ingestion

import (
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/entity"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (con *Controller) ShlokTranslationGeneration(c *gin.Context) {
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
