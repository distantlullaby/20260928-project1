package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mbti-backend/config"
	"mbti-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	errInsufficientCoins = errors.New("insufficient coins")
	errPostNotOpen       = errors.New("post not open")
	errNotPostOwner      = errors.New("not post owner")
	errIsPostOwner       = errors.New("is post owner")
	errResponseNotFound  = errors.New("response not found")
)

type createPostRequest struct {
	AssessmentResultID uint   `json:"assessmentResultId" binding:"required"`
	Bounty             int    `json:"bounty" binding:"required"`
	Note               string `json:"note"`
}

// CreatePost 发布求解读：校验余额后在事务中冻结扣除悬赏币
func CreatePost(c *gin.Context) {
	userID := c.GetUint("userId")

	var req createPostRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Bounty <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "悬赏币数量需为正整数"})
		return
	}
	note := strings.TrimSpace(req.Note)
	if len([]rune(note)) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "备注不能超过 500 字"})
		return
	}

	var post models.InterpretationPost
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// 行锁锁定发帖人账户，防止并发超额冻结
		var user models.User
		if err := tx.Clauses(lockClause()).First(&user, userID).Error; err != nil {
			return err
		}
		// 校验测评结果属于本人
		var assessment models.AssessmentResult
		if err := tx.Where("id = ? AND user_id = ?", req.AssessmentResultID, userID).First(&assessment).Error; err != nil {
			return gorm.ErrRecordNotFound
		}
		if user.CoinBalance < req.Bounty {
			return errInsufficientCoins
		}

		post = models.InterpretationPost{
			UserID:             userID,
			AssessmentResultID: req.AssessmentResultID,
			Note:               note,
			Bounty:             req.Bounty,
			Status:             "open",
		}
		if err := tx.Create(&post).Error; err != nil {
			return err
		}

		newBalance := user.CoinBalance - req.Bounty
		if err := tx.Model(&user).Update("coin_balance", newBalance).Error; err != nil {
			return err
		}
		return tx.Create(&models.CoinTransaction{
			UserID:       userID,
			Change:       -req.Bounty,
			BalanceAfter: newBalance,
			Type:         "post_freeze",
			RefID:        &post.ID,
			Remark:       "发布求解读 · 冻结悬赏",
			CreatedAt:    time.Now(),
		}).Error
	})

	switch {
	case errors.Is(err, errInsufficientCoins):
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "测评币余额不足，无法冻结悬赏"})
		return
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先完成测评，再发布求解读"})
		return
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发布失败，请稍后重试"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"post": post})
}

type appendBountyRequest struct {
	Amount int `json:"amount" binding:"required"`
}

// AppendBounty 追加悬赏：再次冻结追加金额
func AppendBounty(c *gin.Context) {
	userID := c.GetUint("userId")
	postID := parseUintParam(c, "id")

	var req appendBountyRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "追加金额需为正整数"})
		return
	}

	var newBalance int
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var post models.InterpretationPost
		if err := tx.Clauses(lockClause()).First(&post, postID).Error; err != nil {
			return gorm.ErrRecordNotFound
		}
		if post.UserID != userID {
			return errNotPostOwner
		}
		if post.Status != "open" {
			return errPostNotOpen
		}

		var user models.User
		if err := tx.Clauses(lockClause()).First(&user, userID).Error; err != nil {
			return err
		}
		if user.CoinBalance < req.Amount {
			return errInsufficientCoins
		}

		if err := tx.Model(&post).Updates(map[string]interface{}{
			"bounty":     post.Bounty + req.Amount,
			"updated_at": time.Now(),
		}).Error; err != nil {
			return err
		}

		newBalance = user.CoinBalance - req.Amount
		if err := tx.Model(&user).Update("coin_balance", newBalance).Error; err != nil {
			return err
		}
		return tx.Create(&models.CoinTransaction{
			UserID:       userID,
			Change:       -req.Amount,
			BalanceAfter: newBalance,
			Type:         "bounty_append",
			RefID:        &post.ID,
			Remark:       "追加悬赏 · 冻结",
			CreatedAt:    time.Now(),
		}).Error
	})

	if err != nil {
		writePostActionError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "悬赏已追加", "coinBalance": newBalance})
}

type createResponseRequest struct {
	Content string `json:"content" binding:"required"`
}

// CreateResponse 帮 TA 解读：不能给自己的帖子回应
func CreateResponse(c *gin.Context) {
	userID := c.GetUint("userId")
	postID := parseUintParam(c, "id")

	var req createResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "解读内容不能为空"})
		return
	}
	content := strings.TrimSpace(req.Content)
	if len([]rune(content)) < 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "解读内容至少 5 个字"})
		return
	}
	if len([]rune(content)) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "解读内容不能超过 1000 字"})
		return
	}

	var response models.InterpretationResponse
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var post models.InterpretationPost
		if err := tx.First(&post, postID).Error; err != nil {
			return gorm.ErrRecordNotFound
		}
		if post.UserID == userID {
			return errIsPostOwner
		}
		if post.Status != "open" {
			return errPostNotOpen
		}
		// 同一用户对同一帖只能解读一次
		var count int64
		if err := tx.Model(&models.InterpretationResponse{}).
			Where("post_id = ? AND user_id = ?", postID, userID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errDuplicateResponse
		}

		response = models.InterpretationResponse{
			PostID:  postID,
			UserID:  userID,
			Content: content,
		}
		return tx.Create(&response).Error
	})

	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "求解读帖不存在"})
		case errors.Is(err, errIsPostOwner):
			c.JSON(http.StatusBadRequest, gin.H{"error": "不能给自己的测评提交解读"})
		case errors.Is(err, errPostNotOpen):
			c.JSON(http.StatusBadRequest, gin.H{"error": "该帖已结束，无法继续解读"})
		case errors.Is(err, errDuplicateResponse):
			c.JSON(http.StatusConflict, gin.H{"error": "你已经解读过这个帖子了"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "提交失败，请稍后重试"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"response": response})
}

// AcceptResponse 发起人确认采纳：帖状态置为 settled，冻结悬赏结算给解读人
func AcceptResponse(c *gin.Context) {
	userID := c.GetUint("userId")
	postID := parseUintParam(c, "id")
	responseID := parseUintParam(c, "responseId")

	var resolverBalance int
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var post models.InterpretationPost
		if err := tx.Clauses(lockClause()).First(&post, postID).Error; err != nil {
			return gorm.ErrRecordNotFound
		}
		if post.UserID != userID {
			return errNotPostOwner
		}
		if post.Status != "open" {
			return errPostNotOpen
		}

		var response models.InterpretationResponse
		if err := tx.First(&response, responseID).Error; err != nil {
			return errResponseNotFound
		}
		if response.PostID != postID {
			return errResponseNotFound
		}
		if response.UserID == userID {
			return errIsPostOwner
		}

		// 结算：悬赏币入账解读人
		var resolver models.User
		if err := tx.Clauses(lockClause()).First(&resolver, response.UserID).Error; err != nil {
			return err
		}
		resolverBalance = resolver.CoinBalance + post.Bounty
		if err := tx.Model(&resolver).Update("coin_balance", resolverBalance).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.CoinTransaction{
			UserID:       resolver.ID,
			Change:       post.Bounty,
			BalanceAfter: resolverBalance,
			Type:         "settle_income",
			RefID:        &post.ID,
			Remark:       "解读被采纳 · 悬赏收入",
			CreatedAt:    time.Now(),
		}).Error; err != nil {
			return err
		}

		// 帖子标记已结算
		acceptedID := response.ID
		return tx.Model(&post).Updates(map[string]interface{}{
			"status":                "settled",
			"accepted_response_id":  acceptedID,
			"updated_at":            time.Now(),
		}).Error
	})

	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "求解读帖不存在"})
		case errors.Is(err, errNotPostOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": "只有发帖人可以采纳解读"})
		case errors.Is(err, errPostNotOpen):
			c.JSON(http.StatusBadRequest, gin.H{"error": "该帖已结束，无法再次采纳"})
		case errors.Is(err, errResponseNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "解读不存在"})
		case errors.Is(err, errIsPostOwner):
			c.JSON(http.StatusBadRequest, gin.H{"error": "不能采纳自己的解读"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "结算失败，请稍后重试"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已采纳，悬赏币已结算给解读人"})
}

func writePostActionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "求解读帖不存在"})
	case errors.Is(err, errNotPostOwner):
		c.JSON(http.StatusForbidden, gin.H{"error": "只能操作自己发布的帖子"})
	case errors.Is(err, errPostNotOpen):
		c.JSON(http.StatusBadRequest, gin.H{"error": "该帖已结束"})
	case errors.Is(err, errInsufficientCoins):
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "测评币余额不足"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
	}
}

func parseUintParam(c *gin.Context, key string) uint {
	id, _ := strconv.ParseUint(c.Param(key), 10, 64)
	return uint(id)
}

var errDuplicateResponse = errors.New("duplicate response")
