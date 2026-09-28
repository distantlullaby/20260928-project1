package database

import (
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"mbti-backend/config"
	"mbti-backend/models"
)

var DB *gorm.DB

// Init 连接 MySQL 并自动迁移表结构
func Init() {
	var err error
	DB, err = gorm.Open(mysql.Open(config.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	sqlDB, _ := DB.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := DB.AutoMigrate(
		&models.User{},
		&models.Question{},
		&models.TypeProfile{},
		&models.AssessmentResult{},
		&models.InterpretationRequest{},
		&models.Interpretation{},
		&models.CoinTransaction{},
	); err != nil {
		log.Fatalf("自动迁移失败: %v", err)
	}
	log.Println("数据库连接并迁移完成")
}
