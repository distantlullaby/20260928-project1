package handlers

import (
	"gorm.io/gorm/clause"
)

// lockClause 返回 SELECT ... FOR UPDATE 行锁子句（事务内使用）
func lockClause() clause.Expression {
	return clause.Locking{Strength: "UPDATE"}
}
