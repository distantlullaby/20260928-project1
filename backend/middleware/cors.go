package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS 允许 Vite 开发服务器跨域访问
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
