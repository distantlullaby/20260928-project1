package config

import (
	"os"
	"strconv"
)

var (
	// AppAddr HTTP 监听地址（8080 被同机其他项目占用，默认 127.0.0.1:8090）
	AppAddr = "127.0.0.1:8090"
	// JWTSecret JWT 签名密钥
	JWTSecret = "mbti-assess-secret-2026"
	// InitialCoins 注册赠送的初始测评币
	InitialCoins = 200
)

// DSN 返回 MySQL 连接串，可用环境变量覆盖
func DSN() string {
	host := getenv("DB_HOST", "127.0.0.1")
	port := getenv("DB_PORT", "3306")
	user := getenv("DB_USER", "root")
	pass := getenv("DB_PASS", "123456")
	name := getenv("DB_NAME", "mbti_assess")
	return user + ":" + pass + "@tcp(" + host + ":" + port + ")/" + name +
		"?charset=utf8mb4&parseTime=True&loc=Local"
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

// InitialCoinsFromEnv 允许通过环境变量调整注册赠送币数
func init() {
	if v := os.Getenv("INITIAL_COINS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			InitialCoins = n
		}
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		JWTSecret = v
	}
	if v := os.Getenv("APP_ADDR"); v != "" {
		AppAddr = v
	}
}
