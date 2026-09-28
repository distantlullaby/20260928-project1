package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"mbti-backend/database"
	"mbti-backend/middleware"
	"mbti-backend/models"
	"mbti-backend/services"
	"mbti-backend/utils"
)

// requestVO 组装一条求解读（含测评结果+画像+解读回应）给前端
func requestVO(r *models.InterpretationRequest, includeInterps bool) gin.H {
	vo := gin.H{
		"id": r.ID, "user_id": r.UserID,
		"assessment_result_id": r.AssessmentResultID,
		"title": r.Title, "content": r.Content,
		"reward": r.Reward, "status": r.Status,
		"accepted_interpretation_id": r.AcceptedInterpretationID,
		"created_at": r.CreatedAt, "updated_at": r.UpdatedAt,
	}
	if r.User != nil {
		vo["user"] = userVO(r.User)
	}
	if r.Result != nil {
		var dims []utils.DimensionStat
		_ = json.Unmarshal([]byte(r.Result.Dimensions), &dims)
		vo["assessment"] = gin.H{
			"result_id":  r.Result.ID,
			"type_code":  r.Result.TypeCode,
			"dimensions": dims,
			"created_at": r.Result.CreatedAt,
		}
		var profile models.TypeProfile
		if err := database.DB.First(&profile, "code = ?", r.Result.TypeCode).Error; err == nil {
			vo["profile"] = profile
		}
	}
	if includeInterps {
		list := make([]gin.H, 0, len(r.Interpretations))
		for i := range r.Interpretations {
			it := r.Interpretations[i]
			item := gin.H{
				"id": it.ID, "request_id": it.RequestID,
				"content": it.Content, "created_at": it.CreatedAt,
				"accepted": r.AcceptedInterpretationID != nil && *r.AcceptedInterpretationID == it.ID,
			}
			if it.User != nil {
				item["user"] = userVO(it.User)
			}
			list = append(list, item)
		}
		vo["interpretations"] = list
	} else {
		vo["interpretation_count"] = len(r.Interpretations)
	}
	return vo
}

// ListRequests 首页求解读列表
// query: status=open|settled|all, mine=1, page, page_size
func ListRequests(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}

	q := database.DB.Model(&models.InterpretationRequest{})
	if c.Query("mine") == "1" {
		q = q.Where("user_id = ?", middleware.CurrentUserID(c))
	} else if st := c.DefaultQuery("status", "open"); st != "all" {
		q = q.Where("status = ?", st)
	}

	var total int64
	q.Count(&total)

	var reqs []models.InterpretationRequest
	q.Preload("User").
		Preload("Result").
		Preload("Interpretations").
		Preload("Interpretations.User").
		Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&reqs)

	list := make([]gin.H, 0, len(reqs))
	for i := range reqs {
		list = append(list, requestVO(&reqs[i], false))
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "page": page, "page_size": size, "list": list})
}

// GetRequest 求解读详情（含全部解读）
func GetRequest(c *gin.Context) {
	id := c.Param("id")
	var req models.InterpretationRequest
	if err := database.DB.
		Preload("User").
		Preload("Result").
		Preload("Interpretations").
		Preload("Interpretations.User").
		First(&req, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "求解读不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"request": requestVO(&req, true)})
}

type createRequestReq struct {
	AssessmentResultID uint   `json:"assessment_result_id" binding:"required"`
	Title              string `json:"title" binding:"required,min=2,max=100"`
	Content            string `json:"content"`
	Reward             int    `json:"reward" binding:"required,min=1"`
}

// CreateRequest 发布求解读并冻结悬赏
func CreateRequest(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	var req createRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "标题 2-100 字，悬赏币至少 1 枚"})
		return
	}
	r, err := services.CreateRequest(database.DB, uid, req.AssessmentResultID,
		req.Title, req.Content, req.Reward)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, services.ErrInsufficientCoins) {
			status = http.StatusPaymentRequired
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": r.ID, "reward": r.Reward})
}

type appendRewardReq struct {
	Amount int `json:"amount" binding:"required,min=1"`
}

// AppendReward 追加悬赏币
func AppendReward(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	id := c.Param("id")
	reqID, _ := strconv.ParseUint(id, 10, 64)
	var req appendRewardReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "追加币数至少 1 枚"})
		return
	}
	err := services.AppendReward(database.DB, uid, uint(reqID), req.Amount)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, services.ErrInsufficientCoins) {
			status = http.StatusPaymentRequired
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	ok(c, "悬赏已追加并冻结")
}

type interpretReq struct {
	Content string `json:"content" binding:"required,min=10"`
}

// SubmitInterpretation 帮 TA 解读
func SubmitInterpretation(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	id := c.Param("id")
	reqID, _ := strconv.ParseUint(id, 10, 64)
	var req interpretReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "解读内容至少 10 个字"})
		return
	}
	inter, err := services.SubmitInterpretation(database.DB, uint(reqID), uid, req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": inter.ID, "message": "解读已提交，等待发起人采纳"})
}

// AcceptInterpretation 发起人确认采纳，结算悬赏
func AcceptInterpretation(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	reqID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	interID, _ := strconv.ParseUint(c.Param("interpId"), 10, 64)
	err := services.AcceptInterpretation(database.DB, uid, uint(reqID), uint(interID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ok(c, "已采纳，悬赏币已结算给对方")
}

// CloseRequest 关闭请求并退回冻结币
func CloseRequest(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	reqID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := services.CloseRequest(database.DB, uid, uint(reqID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ok(c, "已关闭，冻结的悬赏币已退回")
}

// InterpretedByMe 我解读过的请求（个人中心-我的解读）
func InterpretedByMe(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	var reqs []models.InterpretationRequest
	database.DB.Distinct("interpretation_requests.*").
		Joins("JOIN interpretations ON interpretations.request_id = interpretation_requests.id").
		Where("interpretations.user_id = ?", uid).
		Preload("User").Preload("Result").Preload("Interpretations").Preload("Interpretations.User").
		Order("interpretation_requests.created_at DESC").
		Find(&reqs)
	list := make([]gin.H, 0, len(reqs))
	for i := range reqs {
		list = append(list, requestVO(&reqs[i], true))
	}
	c.JSON(http.StatusOK, gin.H{"list": list})
}

func ok(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{"message": msg})
}
