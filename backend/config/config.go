package config

import (
	"log"
	"time"

	"mbti-backend/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库连接
var DB *gorm.DB

// InitDB 初始化 MySQL 连接并自动迁移表结构
func InitDB() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/mbti_app?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取底层连接池失败: %v", err)
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := db.AutoMigrate(
		&models.User{},
		&models.AssessmentResult{},
		&models.InterpretationPost{},
		&models.InterpretationResponse{},
		&models.CoinTransaction{},
	); err != nil {
		log.Fatalf("数据表迁移失败: %v", err)
	}

	DB = db
	log.Println("数据库初始化完成")
}
