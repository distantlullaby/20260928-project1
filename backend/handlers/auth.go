package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"mbti-backend/config"
	"mbti-backend/database"
	"mbti-backend/middleware"
	"mbti-backend/models"
	"mbti-backend/services"
	"mbti-backend/utils"

	"golang.org/x/crypto/bcrypt"
)

type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Nickname string `json:"nickname" binding:"max=20"`
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func userVO(u *models.User) gin.H {
	nick := u.Nickname
	if nick == "" {
		nick = u.Username
	}
	return gin.H{
		"id": u.ID, "username": u.Username, "nickname": nick,
		"coin_balance": u.CoinBalance, "frozen_balance": u.FrozenBalance,
		"mbti_type": u.MBTIType, "created_at": u.CreatedAt,
	}
}

// Register 注册并赠送初始测评币
func Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名 3-20 位，密码 6-32 位"})
		return
	}

	var existing models.User
	err := database.DB.Where("username = ?", req.Username).First(&existing).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已被注册"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加密失败"})
		return
	}
	nick := req.Nickname
	if nick == "" {
		nick = req.Username
	}
	user := models.User{
		Username: req.Username, PasswordHash: string(hash),
		Nickname: nick,
	}
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return services.GrantOnRegister(tx, user.ID, config.InitialCoins)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败: " + err.Error()})
		return
	}

	token, _ := utils.GenerateToken(user.ID, user.Username)

	// 事务后重新读取，确保响应包含赠送后的真实余额
	var fresh models.User
	if err := database.DB.First(&fresh, user.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": userVO(&fresh)})
}

// Login 登录
func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入用户名和密码"})
		return
	}
	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	token, _ := utils.GenerateToken(user.ID, user.Username)
	c.JSON(http.StatusOK, gin.H{"token": token, "user": userVO(&user)})
}

// Me 当前登录用户信息
func Me(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	var user models.User
	if err := database.DB.First(&user, uid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": userVO(&user)})
}
