package auth

import (
	"context"
	"fmt"
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

func InitZohoAuthController(ctx context.Context, service facade.Service, config *configuration.Configuration) *Controller {
	return &Controller{
		logger:  logging.WithContext(ctx),
		service: service,
		config:  config,
	}
}

func (con *Controller) GetAuthorizationURL(c *gin.Context) {
	state := "123"
	authURL := con.service.ZohoAuthService().GetAuthorizationURL(state)
	c.Redirect(http.StatusFound, authURL)
}

func (con *Controller) HandleAuthCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Missing authorization code",
		})
		return
	}
	ctx := context.Background()
	if err := con.service.ZohoAuthService().ExchangeCodeForTokens(ctx, code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": fmt.Sprintf("Failed to exchange code: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Authorization successful!",
	})
}

func (con *Controller) CheckTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if con.service.ZohoAuthService().IsTokenExpired() {
			if err := con.service.ZohoAuthService().RefreshAccessToken(); err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"status":  http.StatusUnauthorized,
					"message": "Token expired and failed to refresh",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func (con *Controller) ReadZohoSheet(c *gin.Context) {
	sheetID := c.Param("sheet-id")
	if sheetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Sheet ID is required",
		})
		return
	}
	sheetName := c.Query("sheet-name")
	if sheetName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Sheet name is required",
		})
		return
	}
	sheetData, err := con.service.ZohoAuthService().GetSheetData(sheetID, sheetName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": fmt.Sprintf("Failed to fetch sheet data: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"data":   sheetData,
	})
}
