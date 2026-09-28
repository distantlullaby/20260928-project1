package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"mbti-backend/database"
	"mbti-backend/middleware"
	"mbti-backend/models"
	"mbti-backend/services"
	"mbti-backend/utils"
)

// ListQuestions 返回全部测评题（按维度、序号排列）
func ListQuestions(c *gin.Context) {
	var questions []models.Question
	if err := database.DB.Order("sort ASC, id ASC").Find(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取题目失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"questions": questions})
}

// ListProfiles 返回 16 型人格画像
func ListProfiles(c *gin.Context) {
	var profiles []models.TypeProfile
	if err := database.DB.Order("code ASC").Find(&profiles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取画像失败"})
		return
	}
	if code := c.Query("code"); code != "" {
		var p models.TypeProfile
		if err := database.DB.First(&p, "code = ?", code).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "类型不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"profile": p})
		return
	}
	c.JSON(http.StatusOK, gin.H{"profiles": profiles})
}

type submitAssessment struct {
	// Answers: [{question_id, pole: "A"|"B"}]
	Answers []struct {
		QuestionID uint   `json:"question_id"`
		Pole       string `json:"pole"`
	} `json:"answers" binding:"required,min=4"`
}

// SubmitAssessment 提交答卷，计分并保存结果（需登录）
func SubmitAssessment(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	var req submitAssessment
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "答卷数据不完整"})
		return
	}

	var questions []models.Question
	if err := database.DB.Find(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "题目加载失败"})
		return
	}
	qmap := map[uint]models.Question{}
	for _, q := range questions {
		qmap[q.ID] = q
	}

	picks := make([]string, 0, len(req.Answers))
	for _, a := range req.Answers {
		q, ok := qmap[a.QuestionID]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "包含无效题目"})
			return
		}
		switch a.Pole {
		case "A":
			picks = append(picks, q.OptionAPole)
		case "B":
			picks = append(picks, q.OptionBPole)
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "选项必须为 A 或 B"})
			return
		}
	}

	typeCode, dims := utils.ScoreMBTI(picks)
	dimsJSON, _ := json.Marshal(dims)

	result, err := services.SaveResult(database.DB, uid, typeCode, dimsJSON)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var profile models.TypeProfile
	_ = database.DB.First(&profile, "code = ?", typeCode).Error

	c.JSON(http.StatusOK, gin.H{
		"result_id":  result.ID,
		"type_code":  typeCode,
		"dimensions": dims,
		"profile":    profile,
	})
}

// MyResults 当前用户的测评历史
func MyResults(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	var results []models.AssessmentResult
	if err := database.DB.Where("user_id = ?", uid).
		Order("created_at DESC").Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}
	list := make([]gin.H, 0, len(results))
	for i := range results {
		r := results[i]
		var dims []utils.DimensionStat
		_ = json.Unmarshal([]byte(r.Dimensions), &dims)
		list = append(list, gin.H{
			"id": r.ID, "type_code": r.TypeCode,
			"dimensions": dims, "created_at": r.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"results": list})
}

// GetResult 单条测评结果（供求解读卡片正面使用）
func GetResult(c *gin.Context) {
	id := c.Param("id")
	var result models.AssessmentResult
	if err := database.DB.First(&result, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "测评结果不存在"})
		return
	}
	var dims []utils.DimensionStat
	_ = json.Unmarshal([]byte(result.Dimensions), &dims)
	var profile models.TypeProfile
	_ = database.DB.First(&profile, "code = ?", result.TypeCode).Error
	var owner models.User
	_ = database.DB.First(&owner, result.UserID).Error

	c.JSON(http.StatusOK, gin.H{
		"id": result.ID, "type_code": result.TypeCode,
		"dimensions": dims, "profile": profile,
		"created_at": result.CreatedAt,
		"owner":      userVO(&owner),
	})
}
