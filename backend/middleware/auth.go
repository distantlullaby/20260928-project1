package middleware

import (
	"net/http"
	"strings"

	"mbti-backend/utils"

	"github.com/gin-gonic/gin"
)

// Auth JWT 鉴权中间件，校验通过后把 userId 写入上下文
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}
		userID, err := utils.ParseToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录已过期，请重新登录"})
			return
		}
		c.Set("userId", userID)
		c.Next()
	}
}
