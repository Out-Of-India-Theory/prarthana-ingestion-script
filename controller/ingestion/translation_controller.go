package ingestion

import (
	"context"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/entity"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/util"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func (con *Controller) ShlokTranslationGeneration(c *gin.Context) {
	var request entity.IngestionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid request payload",
		})
		return
	}
	accessToken := c.Request.Header.Get("zoho-access-token")
	go func(startID, endID int, token string) {
		ctx := context.Background()
		ctx = util.SetZohoAccessTokenInContext(ctx, token)
		err := con.service.ShlokTranslationService().GenerateShlokaTranslation(ctx, startID, endID)
		if err != nil {
			log.Printf("❌ Background translation failed: %v", err)
		} else {
			log.Printf("✅ Background translation completed for ID range: %d - %d", startID, endID)
		}
	}(request.StartID, request.EndID, accessToken)

	c.JSON(http.StatusAccepted, gin.H{
		"status":  http.StatusAccepted,
		"message": "Translation started in background",
	})
}