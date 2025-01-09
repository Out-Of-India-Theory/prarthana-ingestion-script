package middleware

import (
	"github.com/Out-Of-India-Theory/prarthana-automated-script/configuration"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/service/zoho"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthMiddleware struct {
	config          *configuration.Configuration
	zohoAuthService zoho.Service
}

func InitAuthMiddleware(config *configuration.Configuration, zohoAuthService zoho.Service) *AuthMiddleware {
	return &AuthMiddleware{
		config:          config,
		zohoAuthService: zohoAuthService,
	}
}

func (am *AuthMiddleware) ZohoAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := am.zohoAuthService.RefreshAccessToken()
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to refresh access token"})
			c.Abort()
			return
		}
		c.Set("access-token", accessToken)
		c.Request.Header.Set("zoho-access-token", accessToken)
		c.Next()
	}
}
