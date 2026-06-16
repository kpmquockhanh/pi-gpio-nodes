package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS middleware
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-API-Key, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// APIKeyMiddleware validates the API key for mutating endpoints
func APIKeyMiddleware(expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for GET requests (read-only)
		if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		key := c.GetHeader("X-API-Key")
		if key == "" {
			key = c.Query("api_key")
		}

		if key != expectedKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing api key"})
			return
		}

		c.Next()
	}
}

// LoggerMiddleware logs requests
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}
