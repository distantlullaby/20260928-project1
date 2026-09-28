package handlers

import (
	"net/http"
	"strconv"
	"time"

	"mbti-backend/config"
	"mbti-backend/models"

	"github.com/gin-gonic/gin"
)

// userBrief 用户简要信息
type userBrief struct {
	ID       uint   `json:"id"`
	Nickname string `json:"nickname"`
}

// responseItem 解读回应（附带解读人信息）
type responseItem struct {
	models.InterpretationResponse
	Resolver userBrief `json:"resolver"`
}

// postItem 求解读帖列表/详情项，正面为测评结果+类型画像，回应对他人解读视角
type postItem struct {
	models.InterpretationPost
	Author     userBrief                 `json:"author"`
	Assessment gin.H                     `json:"assessment"`
	Responses  []responseItem            `json:"responses"`
	Accepted   *models.InterpretationResponse `json:"accepted,omitempty"`
}

// ListPosts 首页求解读列表：?status=open|settled&mine=1
func ListPosts(c *gin.Context) {
	var userID uint
	if uid, exists := c.Get("userId"); exists {
		userID = uid.(uint)
	}

	q := config.DB.Model(&models.InterpretationPost{})
	if c.Query("status") != "" {
		q = q.Where("status = ?", c.Query("status"))
	} else {
		q = q.Where("status IN ?", []string{"open", "settled"})
	}
	if c.Query("mine") == "1" {
		q = q.Where("user_id = ?", userID)
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	var posts []models.InterpretationPost
	if err := q.Order("created_at DESC").Limit(limit).Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取列表失败"})
		return
	}

	items := make([]postItem, 0, len(posts))
	for i := range posts {
		items = append(items, assemblePostItem(&posts[i], userID))
	}
	c.JSON(http.StatusOK, gin.H{"posts": items})
}

// GetPost 帖子详情
func GetPost(c *gin.Context) {
	postID := parseUintParam(c, "id")
	var userID uint
	if uid, exists := c.Get("userId"); exists {
		userID = uid.(uint)
	}

	var post models.InterpretationPost
	if err := config.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "求解读帖不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"post": assemblePostItem(&post, userID)})
}

// assemblePostItem 组装帖子的作者、测评画像、解读回应
func assemblePostItem(post *models.InterpretationPost, currentUserID uint) postItem {
	item := postItem{InterpretationPost: *post, Responses: []responseItem{}}

	var author models.User
	if err := config.DB.First(&author, post.UserID).Error; err == nil {
		item.Author = userBrief{ID: author.ID, Nickname: author.Nickname}
	}

	var assessment models.AssessmentResult
	if err := config.DB.First(&assessment, post.AssessmentResultID).Error; err == nil {
		item.Assessment = formatAssessment(&assessment)
	}

	var responses []models.InterpretationResponse
	config.DB.Where("post_id = ?", post.ID).Order("created_at ASC").Find(&responses)

	userCache := map[uint]userBrief{}
	for _, r := range responses {
		brief, ok := userCache[r.UserID]
		if !ok {
			var u models.User
			if err := config.DB.First(&u, r.UserID).Error; err == nil {
				brief = userBrief{ID: u.ID, Nickname: u.Nickname}
			} else {
				brief = userBrief{ID: r.UserID, Nickname: "已注销用户"}
			}
			userCache[r.UserID] = brief
		}
		ri := responseItem{InterpretationResponse: r, Resolver: brief}
		item.Responses = append(item.Responses, ri)
		if post.AcceptedResponseID != nil && r.ID == *post.AcceptedResponseID {
			accepted := r
			item.Accepted = &accepted
		}
	}
	return item
}

// GetMyPostOverview 个人中心：我发布的 / 我解读的帖子摘要
func GetMyPostOverview(c *gin.Context) {
	userID := c.GetUint("userId")

	var mine []models.InterpretationPost
	config.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&mine)
	mineItems := make([]gin.H, 0, len(mine))
	for i := range mine {
		p := &mine[i]
		var respCount int64
		config.DB.Model(&models.InterpretationResponse{}).Where("post_id = ?", p.ID).Count(&respCount)
		mineItems = append(mineItems, gin.H{
			"id":           p.ID,
			"bounty":       p.Bounty,
			"status":       p.Status,
			"responseCount": respCount,
			"createdAt":    p.CreatedAt,
		})
	}

	var myResponses []models.InterpretationResponse
	config.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&myResponses)
	resolvedItems := make([]gin.H, 0, len(myResponses))
	for _, r := range myResponses {
		var post models.InterpretationPost
		if err := config.DB.First(&post, r.PostID).Error; err != nil {
			continue
		}
		var author models.User
		config.DB.First(&author, post.UserID)
		earned := post.AcceptedResponseID != nil && *post.AcceptedResponseID == r.ID
		resolvedItems = append(resolvedItems, gin.H{
			"responseId":  r.ID,
			"postId":      post.ID,
			"postStatus":  post.Status,
			"bounty":      post.Bounty,
			"earned":      earned,
			"author":      userBrief{ID: author.ID, Nickname: author.Nickname},
			"content":     r.Content,
			"createdAt":   r.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"myPosts":     mineItems,
		"myResponses": resolvedItems,
	})
}

// GetCoinTransactions 测评币流水
func GetCoinTransactions(c *gin.Context) {
	userID := c.GetUint("userId")

	var txns []models.CoinTransaction
	if err := config.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(100).Find(&txns).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取流水失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"transactions": txns})
}

// GetMe 个人中心主页信息
func GetMe(c *gin.Context) {
	userID := c.GetUint("userId")

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	var latest models.AssessmentResult
	var latestData gin.H
	if err := config.DB.Where("user_id = ?", userID).Order("created_at DESC").First(&latest).Error; err == nil {
		latestData = formatAssessment(&latest)
	}

	var assessmentCount int64
	config.DB.Model(&models.AssessmentResult{}).Where("user_id = ?", userID).Count(&assessmentCount)
	var postCount, settledCount int64
	config.DB.Model(&models.InterpretationPost{}).Where("user_id = ?", userID).Count(&postCount)
	config.DB.Model(&models.InterpretationPost{}).Where("user_id = ? AND status = ?", userID, "settled").Count(&settledCount)
	var earnedTotal int64
	row := config.DB.Model(&models.CoinTransaction{}).Where("user_id = ? AND type = ?", userID, "settle_income").
		Select("COALESCE(SUM(`change`), 0)").Row()
	_ = row.Scan(&earnedTotal)

	c.JSON(http.StatusOK, gin.H{
		"user":             userResponse(&user),
		"latestAssessment": latestData,
		"stats": gin.H{
			"assessmentCount": assessmentCount,
			"postCount":       postCount,
			"settledCount":    settledCount,
			"earnedTotal":     earnedTotal,
		},
		"serverTime": time.Now(),
	})
}
