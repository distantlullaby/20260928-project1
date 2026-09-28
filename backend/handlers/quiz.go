package handlers

import (
	"encoding/json"
	"net/http"

	"mbti-backend/config"
	"mbti-backend/mbti"
	"mbti-backend/models"

	"github.com/gin-gonic/gin"
)

// GetQuestions 返回题库
func GetQuestions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"questions": mbti.Questions})
}

type submitRequest struct {
	// Answers 题目 ID -> 选项（"A" 或 "B"）
	Answers map[int]string `json:"answers" binding:"required"`
}

type dimensionScore struct {
	Positive    string `json:"positive"`
	Negative    string `json:"negative"`
	PosCount    int    `json:"posCount"`
	NegCount    int    `json:"negCount"`
	PosPercent  int    `json:"posPercent"`
	NegPercent  int    `json:"negPercent"`
	Dominant    string `json:"dominant"`
}

// SubmitAssessment 提交答案，计算四维倾向占比与人格类型
func SubmitAssessment(c *gin.Context) {
	userID := c.GetUint("userId")

	var req submitRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Answers) != len(mbti.Questions) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请完成全部题目后再提交"})
		return
	}

	// counts: 每一极的得分
	counts := map[string]int{"E": 0, "I": 0, "S": 0, "N": 0, "T": 0, "F": 0, "J": 0, "P": 0}
	for _, q := range mbti.Questions {
		ans, ok := req.Answers[q.ID]
		if !ok || (ans != "A" && ans != "B") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "存在未作答或非法的题目"})
			return
		}
		if ans == "A" {
			counts[q.PolA]++
		} else {
			counts[q.PolB]++
		}
	}

	typeCode := ""
	dimensions := make([]dimensionScore, 0, 4)
	for _, d := range mbti.Dims {
		pos, neg := d.Poles[0], d.Poles[1]
		posCount, negCount := counts[pos], counts[neg]
		total := posCount + negCount
		dominant := pos
		if negCount > posCount {
			dominant = neg
		}
		posPercent, negPercent := percentPair(posCount, total)
		typeCode += dominant
		dimensions = append(dimensions, dimensionScore{
			Positive:   pos,
			Negative:   neg,
			PosCount:   posCount,
			NegCount:   negCount,
			PosPercent: posPercent,
			NegPercent: negPercent,
			Dominant:   dominant,
		})
	}

	scoresJSON, _ := json.Marshal(counts)
	result := models.AssessmentResult{
		UserID:   userID,
		TypeCode: typeCode,
		Scores:   string(scoresJSON),
	}
	if err := config.DB.Create(&result).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "测评结果保存失败"})
		return
	}

	profile := mbti.Profiles[typeCode]
	c.JSON(http.StatusCreated, gin.H{
		"resultId":   result.ID,
		"typeCode":   typeCode,
		"dimensions": dimensions,
		"profile":    profile,
	})
}

// GetLatestAssessment 获取当前用户最近一次测评结果
func GetLatestAssessment(c *gin.Context) {
	userID := c.GetUint("userId")

	var result models.AssessmentResult
	if err := config.DB.Where("user_id = ?", userID).Order("created_at DESC").First(&result).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"result": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": formatAssessment(&result)})
}

func formatAssessment(r *models.AssessmentResult) gin.H {
	counts := map[string]int{}
	_ = json.Unmarshal([]byte(r.Scores), &counts)

	dimensions := make([]dimensionScore, 0, 4)
	for _, d := range mbti.Dims {
		pos, neg := d.Poles[0], d.Poles[1]
		posCount, negCount := counts[pos], counts[neg]
		total := posCount + negCount
		dominant := pos
		if negCount > posCount {
			dominant = neg
		}
		posPercent, negPercent := percentPair(posCount, total)
		dimensions = append(dimensions, dimensionScore{
			Positive:   pos,
			Negative:   neg,
			PosCount:   posCount,
			NegCount:   negCount,
			PosPercent: posPercent,
			NegPercent: negPercent,
			Dominant:   dominant,
		})
	}
	return gin.H{
		"resultId":   r.ID,
		"typeCode":   r.TypeCode,
		"dimensions": dimensions,
		"profile":    mbti.Profiles[r.TypeCode],
		"createdAt":  r.CreatedAt,
	}
}

// percentPair 返回积极极百分比与消极极百分比，四舍五入后保证两极之和为 100
func percentPair(posCount, total int) (int, int) {
	if total == 0 {
		return 0, 0
	}
	posExact := float64(posCount) * 100 / float64(total)
	posPercent := int(posExact + 0.5)
	if posPercent < 0 {
		posPercent = 0
	}
	if posPercent > 100 {
		posPercent = 100
	}
	return posPercent, 100 - posPercent
}
