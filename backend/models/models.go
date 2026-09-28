package models

import "time"

// 求解读请求状态
const (
	StatusOpen    = "open"    // 悬赏中，等待/可采纳解读
	StatusSettled = "settled" // 已采纳结算
	StatusClosed  = "closed"  // 已关闭并退款
)

// 测评币流水类型
const (
	TxnRegister   = "register"     // 注册赠送
	TxnFreeze     = "freeze"       // 发布冻结
	TxnAppend     = "append_freeze" // 追加冻结
	TxnSettleIn   = "settle_in"    // 解读被采纳收入
	TxnSettleOut  = "settle_out"   // 发起人悬赏结算出账（冻结核销）
	TxnRefund     = "refund"       // 关闭请求退回
)

// User 用户
type User struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Username      string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	PasswordHash  string    `gorm:"size:255;not null" json:"-"`
	Nickname      string    `gorm:"size:50" json:"nickname"`
	CoinBalance   int       `gorm:"not null;default:0" json:"coin_balance"`   // 可用测评币
	FrozenBalance int       `gorm:"not null;default:0" json:"frozen_balance"` // 悬赏冻结中
	MBTIType      string    `gorm:"size:4;index" json:"mbti_type"`            // 最近一次测评类型
	CreatedAt     time.Time `json:"created_at"`
}

// Question MBTI 测评题（四个维度各 8 题，共 32 题）
type Question struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Dimension  string `gorm:"size:4;index" json:"dimension"` // EI / SN / TF / JP
	Text       string `gorm:"size:255;not null" json:"text"`
	OptionAText string `gorm:"size:255;not null" json:"option_a_text"`
	OptionAPole string `gorm:"size:2;not null" json:"option_a_pole"` // E/I/S/N/T/F/J/P
	OptionBText string `gorm:"size:255;not null" json:"option_b_text"`
	OptionBPole string `gorm:"size:2;not null" json:"option_b_pole"`
	Sort       int    `json:"sort"`
}

// TypeProfile 十六型人格画像
type TypeProfile struct {
	Code        string `gorm:"primaryKey;size:4" json:"code"` // 如 INTJ
	Nickname    string `gorm:"size:30" json:"nickname"`       // 如 建筑师
	Group       string `gorm:"size:20" json:"group"`          // 分析家/外交家/守护者/探险家
	GroupKey    string `gorm:"size:4" json:"group_key"`       // NT/NF/SJ/SP
	Color       string `gorm:"size:20" json:"color"`
	Description string `gorm:"type:text" json:"description"`
	Traits      string `gorm:"type:text" json:"traits"`     // 关键词，顿号分隔
	Strengths   string `gorm:"type:text" json:"strengths"`
	Weaknesses  string `gorm:"type:text" json:"weaknesses"`
	Careers     string `gorm:"type:text" json:"careers"`
}

// AssessmentResult 一次测评的结果
type AssessmentResult struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	TypeCode   string    `gorm:"size:4;index" json:"type_code"`
	Dimensions string    `gorm:"type:text" json:"dimensions"` // 四维占比 JSON
	CreatedAt  time.Time `json:"created_at"`
}

// InterpretationRequest 求解读请求
type InterpretationRequest struct {
	ID                       uint      `gorm:"primaryKey" json:"id"`
	UserID                   uint      `gorm:"index;not null" json:"user_id"`
	AssessmentResultID       uint      `gorm:"not null" json:"assessment_result_id"`
	Title                    string    `gorm:"size:100;not null" json:"title"`
	Content                  string    `gorm:"type:text" json:"content"`
	Reward                   int       `gorm:"not null;default:0" json:"reward"` // 当前悬赏总额（冻结）
	Status                   string    `gorm:"size:20;index;not null;default:open" json:"status"`
	AcceptedInterpretationID *uint     `json:"accepted_interpretation_id"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`

	User            *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Result          *AssessmentResult `gorm:"foreignKey:AssessmentResultID" json:"result,omitempty"`
	Interpretations []Interpretation  `gorm:"foreignKey:RequestID" json:"interpretations,omitempty"`
}

// Interpretation 他人对一条求解读的回应
type Interpretation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RequestID uint      `gorm:"index;not null" json:"request_id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// CoinTransaction 测评币流水（账目快照）
type CoinTransaction struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index;not null" json:"user_id"`
	ChangeAmount  int       `gorm:"not null" json:"change_amount"` // 正为收入，负为支出
	BalanceAfter  int       `gorm:"not null" json:"balance_after"` // 变动后可用余额
	Type          string    `gorm:"size:30;not null;index" json:"type"`
	RelatedID     uint      `json:"related_id"` // 关联请求/解读 ID
	Remark        string    `gorm:"size:255" json:"remark"`
	CreatedAt     time.Time `json:"created_at"`
}
