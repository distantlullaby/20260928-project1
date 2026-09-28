package models

import "time"

// User 用户表
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:100;not null" json:"-"`
	Nickname     string    `gorm:"size:50" json:"nickname"`
	CoinBalance  int       `gorm:"not null;default:0" json:"coinBalance"`
	CreatedAt    time.Time `json:"createdAt"`
}

// AssessmentResult 测评结果表（Scores 存四维原始分，如 {"E":3,"I":4,...}）
type AssessmentResult struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"userId"`
	TypeCode  string    `gorm:"size:4;not null" json:"typeCode"`
	Scores    string    `gorm:"type:json;not null" json:"-"`
	CreatedAt time.Time `json:"createdAt"`
}

// InterpretationPost 求解读帖（悬赏币在发帖时已从发帖人余额冻结扣除）
type InterpretationPost struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	UserID             uint       `gorm:"index;not null" json:"userId"`
	AssessmentResultID uint       `gorm:"not null" json:"assessmentResultId"`
	Note               string     `gorm:"size:500" json:"note"`
	Bounty             int        `gorm:"not null;default:0" json:"bounty"`
	// Status: open 进行中 / settled 已采纳结算 / closed 已关闭
	Status             string     `gorm:"size:10;index;not null;default:open" json:"status"`
	AcceptedResponseID *uint      `json:"acceptedResponseId"`
	CreatedAt          time.Time  `gorm:"index" json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

// InterpretationResponse 解读回应表
type InterpretationResponse struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PostID    uint      `gorm:"index;not null" json:"postId"`
	UserID    uint      `gorm:"index;not null" json:"userId"`
	Content   string    `gorm:"size:1000;not null" json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

// CoinTransaction 测评币流水表（Change 为带符号增减，BalanceAfter 为变动后余额）
type CoinTransaction struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"userId"`
	Change       int       `gorm:"not null" json:"change"`
	BalanceAfter int       `gorm:"not null" json:"balanceAfter"`
	// Type: register_gift 注册赠送 / post_freeze 发帖冻结 / bounty_append 追加冻结 / settle_income 采纳收入
	Type         string    `gorm:"size:20;not null" json:"type"`
	RefID        *uint     `json:"refId"`
	Remark       string    `gorm:"size:100" json:"remark"`
	CreatedAt    time.Time `json:"createdAt"`
}
