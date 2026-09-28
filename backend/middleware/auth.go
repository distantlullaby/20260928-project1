package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"mbti-backend/utils"
)

// Auth 必须登录，userID 写入上下文
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := parseToken(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}
		c.Set("userID", uid)
		c.Next()
	}
}

// OptionalAuth 可选登录：有合法 token 时写入 userID，无则匿名
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if uid, ok := parseToken(c); ok {
			c.Set("userID", uid)
		}
		c.Next()
	}
}

func parseToken(c *gin.Context) (uint, bool) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return 0, false
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0, false
	}
	claims, err := utils.ParseToken(parts[1])
	if err != nil {
		return 0, false
	}
	return claims.UserID, true
}

// CurrentUserID 从上下文取当前用户 ID
func CurrentUserID(c *gin.Context) uint {
	if v, ok := c.Get("userID"); ok {
		return v.(uint)
	}
	return 0
}
