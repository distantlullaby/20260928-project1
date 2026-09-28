package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mbti-backend/database"
	"mbti-backend/middleware"
	"mbti-backend/models"
)

// MyTransactions 我的测评币流水
func MyTransactions(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	var txns []models.CoinTransaction
	if err := database.DB.Where("user_id = ?", uid).
		Order("created_at DESC, id DESC").Limit(200).Find(&txns).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取流水失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"transactions": txns})
}
