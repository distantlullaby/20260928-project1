package services

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"mbti-backend/models"
)

// 测评币相关业务错误
var (
	ErrInsufficientCoins = errors.New("可用测评币不足")
	ErrRequestNotOpen    = errors.New("该求解读已不可操作")
	ErrInterpretationNF  = errors.New("解读不存在或不属于该请求")
	ErrCannotAcceptSelf  = errors.New("不能采纳自己的解读")
)

// recordTxn 在流水表记账（须在事务内调用），以行锁读取最新余额
func recordTxn(tx *gorm.DB, userID uint, change int, txnType string, relatedID uint, remark string) (int, error) {
	var user models.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&user, userID).Error; err != nil {
		return 0, err
	}
	newBalance := user.CoinBalance + change
	if newBalance < 0 || user.FrozenBalance < 0 {
		return 0, ErrInsufficientCoins
	}
	if err := tx.Model(&user).Update("coin_balance", newBalance).Error; err != nil {
		return 0, err
	}
	txn := models.CoinTransaction{
		UserID:       userID,
		ChangeAmount: change,
		BalanceAfter: newBalance,
		Type:         txnType,
		RelatedID:    relatedID,
		Remark:       remark,
	}
	if err := tx.Create(&txn).Error; err != nil {
		return 0, err
	}
	return newBalance, nil
}

// GrantOnRegister 注册赠送初始测评币（独立事务）
func GrantOnRegister(db *gorm.DB, userID uint, amount int) error {
	return db.Transaction(func(tx *gorm.DB) error {
		_, err := recordTxn(tx, userID, amount, models.TxnRegister, 0, "注册赠送测评币")
		return err
	})
}

// freeze 冻结悬赏：可用扣减、冻结增加，并写一条负向流水
func freeze(tx *gorm.DB, userID uint, amount int, reqID uint, txnType, remark string) error {
	var user models.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&user, userID).Error; err != nil {
		return err
	}
	if user.CoinBalance < amount {
		return ErrInsufficientCoins
	}
	if err := tx.Model(&user).Updates(map[string]interface{}{
		"coin_balance":   gorm.Expr("coin_balance - ?", amount),
		"frozen_balance": gorm.Expr("frozen_balance + ?", amount),
	}).Error; err != nil {
		return err
	}
	txn := models.CoinTransaction{
		UserID:       userID,
		ChangeAmount: -amount,
		BalanceAfter: user.CoinBalance - amount,
		Type:         txnType,
		RelatedID:    reqID,
		Remark:       remark,
	}
	return tx.Create(&txn).Error
}

// unfreeze 释放冻结：冻结扣减，可选转入可用余额
func unfreeze(tx *gorm.DB, userID uint, amount int) error {
	var user models.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&user, userID).Error; err != nil {
		return err
	}
	if user.FrozenBalance < amount {
		return errors.New("冻结余额不足以释放")
	}
	return tx.Model(&user).Update("frozen_balance",
		gorm.Expr("frozen_balance - ?", amount)).Error
}

// CreateRequest 发布求解读：事务内创建请求并冻结悬赏币
func CreateRequest(db *gorm.DB, userID, resultID uint, title, content string, reward int) (*models.InterpretationRequest, error) {
	if reward <= 0 {
		return nil, errors.New("悬赏币必须大于 0")
	}
	req := &models.InterpretationRequest{}
	err := db.Transaction(func(tx *gorm.DB) error {
		// 校验测评结果归属
		var result models.AssessmentResult
		if err := tx.First(&result, resultID).Error; err != nil {
			return errors.New("测评结果不存在")
		}
		if result.UserID != userID {
			return errors.New("只能基于自己的测评结果发布求解读")
		}
		req.UserID = userID
		req.AssessmentResultID = resultID
		req.Title = title
		req.Content = content
		req.Reward = reward
		req.Status = models.StatusOpen
		if err := tx.Create(req).Error; err != nil {
			return err
		}
		return freeze(tx, userID, reward, req.ID, models.TxnFreeze, "发布求解读冻结悬赏")
	})
	if err != nil {
		return nil, err
	}
	return req, nil
}

// AppendReward 追加悬赏：再次冻结并累加到请求
func AppendReward(db *gorm.DB, userID, reqID uint, amount int) error {
	if amount <= 0 {
		return errors.New("追加币数必须大于 0")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var req models.InterpretationRequest
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&req, reqID).Error; err != nil {
			return errors.New("求解读不存在")
		}
		if req.UserID != userID {
			return errors.New("只能给自己的求解读追加悬赏")
		}
		if req.Status != models.StatusOpen {
			return ErrRequestNotOpen
		}
		if err := freeze(tx, userID, amount, reqID, models.TxnAppend, "追加悬赏冻结"); err != nil {
			return err
		}
		return tx.Model(&req).Update("reward", gorm.Expr("reward + ?", amount)).Error
	})
}

// AcceptInterpretation 采纳解读：解冻发起人的冻结额，把等额币结算给解读人
func AcceptInterpretation(db *gorm.DB, ownerID, reqID, interpretationID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var req models.InterpretationRequest
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&req, reqID).Error; err != nil {
			return errors.New("求解读不存在")
		}
		if req.UserID != ownerID {
			return errors.New("只能采纳自己求解读下的回应")
		}
		if req.Status != models.StatusOpen {
			return ErrRequestNotOpen
		}

		var inter models.Interpretation
		if err := tx.First(&inter, interpretationID).Error; err != nil {
			return ErrInterpretationNF
		}
		if inter.RequestID != reqID {
			return ErrInterpretationNF
		}
		if inter.UserID == ownerID {
			return ErrCannotAcceptSelf
		}

		reward := req.Reward

		// 1. 释放发起人冻结
		if err := unfreeze(tx, ownerID, reward); err != nil {
			return err
		}
		// 2. 标记结算（冻结以核销方式出账，不动发起人可用余额）
		if err := tx.Model(&models.InterpretationRequest{}).Where("id = ?", reqID).
			Updates(map[string]interface{}{
				"status":                      models.StatusSettled,
				"accepted_interpretation_id":  interpretationID,
			}).Error; err != nil {
			return err
		}
		// 3. 发起人记录核销流水
		ownerTxn := models.CoinTransaction{
			UserID: ownerID, ChangeAmount: 0, BalanceAfter: balanceOf(tx, ownerID),
			Type: models.TxnSettleOut, RelatedID: reqID,
			Remark: "采纳解读，悬赏币结算给对方",
		}
		if err := tx.Create(&ownerTxn).Error; err != nil {
			return err
		}
		// 4. 解读人入账
		if _, err := recordTxn(tx, inter.UserID, reward, models.TxnSettleIn, reqID,
			"解读被采纳获得悬赏"); err != nil {
			return err
		}
		return nil
	})
}

// CloseRequest 关闭尚未采纳的请求：全额退回冻结币
func CloseRequest(db *gorm.DB, userID, reqID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var req models.InterpretationRequest
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&req, reqID).Error; err != nil {
			return errors.New("求解读不存在")
		}
		if req.UserID != userID {
			return errors.New("只能关闭自己的求解读")
		}
		if req.Status != models.StatusOpen {
			return ErrRequestNotOpen
		}
		reward := req.Reward
		if err := unfreeze(tx, userID, reward); err != nil {
			return err
		}
		if err := tx.Model(&req).Update("status", models.StatusClosed).Error; err != nil {
			return err
		}
		_, err := recordTxn(tx, userID, reward, models.TxnRefund, reqID, "关闭求解读，悬赏币退回")
		return err
	})
}

func balanceOf(tx *gorm.DB, userID uint) int {
	var u models.User
	_ = tx.First(&u, userID).Error
	return u.CoinBalance
}

// SubmitInterpretation 提交解读回应
func SubmitInterpretation(db *gorm.DB, reqID, userID uint, content string) (*models.Interpretation, error) {
	if len([]rune(content)) < 10 {
		return nil, errors.New("解读内容至少 10 个字")
	}
	inter := &models.Interpretation{}
	err := db.Transaction(func(tx *gorm.DB) error {
		var req models.InterpretationRequest
		if err := tx.First(&req, reqID).Error; err != nil {
			return errors.New("求解读不存在")
		}
		if req.Status != models.StatusOpen {
			return ErrRequestNotOpen
		}
		if req.UserID == userID {
			return errors.New("不能解读自己的求解读")
		}
		var cnt int64
		if err := tx.Model(&models.Interpretation{}).
			Where("request_id = ? AND user_id = ?", reqID, userID).
			Count(&cnt).Error; err != nil {
			return err
		}
		if cnt > 0 {
			return errors.New("你已经解读过这条请求了")
		}
		inter.RequestID = reqID
		inter.UserID = userID
		inter.Content = content
		return tx.Create(inter).Error
	})
	if err != nil {
		return nil, err
	}
	return inter, nil
}

// SaveResult 保存测评结果并同步用户的当前人格类型
func SaveResult(db *gorm.DB, userID uint, typeCode string, dims []byte) (*models.AssessmentResult, error) {
	result := &models.AssessmentResult{}
	err := db.Transaction(func(tx *gorm.DB) error {
		result.UserID = userID
		result.TypeCode = typeCode
		result.Dimensions = string(dims)
		if err := tx.Create(result).Error; err != nil {
			return err
		}
		// 校验 JSON 合法
		var raw json.RawMessage = dims
		if !json.Valid(raw) {
			return errors.New("维度数据格式错误")
		}
		return tx.Model(&models.User{}).Where("id = ?", userID).
			Update("mbti_type", typeCode).Error
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
