package handlers

import (
	"net/http"
	"strings"
	"time"

	"mbti-backend/config"
	"mbti-backend/models"
	"mbti-backend/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// InitialGiftCoins 注册赠送的初始测评币
const InitialGiftCoins = 100

type registerRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
}

// Register 注册：建用户 + 赠送初始测评币（事务）
func Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码不能为空"})
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if len(req.Username) < 3 || len(req.Username) > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名长度需在 3-20 个字符之间"})
		return
	}
	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码至少需要 6 位"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		nickname = req.Username
	}
	user := models.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Nickname:     nickname,
		CoinBalance:  InitialGiftCoins,
	}

	// 事务：创建用户 + 注册赠送流水
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			if strings.Contains(err.Error(), "Duplicate") {
				return errUserExists
			}
			return err
		}
		txn := models.CoinTransaction{
			UserID:       user.ID,
			Change:       InitialGiftCoins,
			BalanceAfter: InitialGiftCoins,
			Type:         "register_gift",
			Remark:       "注册赠送测评币",
			CreatedAt:    time.Now(),
		}
		return tx.Create(&txn).Error
	})
	if err != nil {
		if err == errUserExists {
			c.JSON(http.StatusConflict, gin.H{"error": "用户名已被注册"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败，请稍后重试"})
		return
	}

	token, _ := utils.GenerateToken(user.ID)
	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user":  userResponse(&user),
	})
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 登录并签发 JWT
func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码不能为空"})
		return
	}

	var user models.User
	if err := config.DB.Where("username = ?", strings.TrimSpace(req.Username)).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	token, _ := utils.GenerateToken(user.ID)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  userResponse(&user),
	})
}

// userResponse 用户信息（不含密码）
func userResponse(u *models.User) gin.H {
	return gin.H{
		"id":          u.ID,
		"username":    u.Username,
		"nickname":    u.Nickname,
		"coinBalance": u.CoinBalance,
		"createdAt":   u.CreatedAt,
	}
}
