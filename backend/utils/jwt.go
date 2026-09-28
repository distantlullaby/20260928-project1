package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("mbti-app-secret-key-2026")

type Claims struct {
	UserID uint `json:"userId"`
	jwt.RegisteredClaims
}

// GenerateToken 签发 7 天有效的 JWT
func GenerateToken(userID uint) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}

// ParseToken 解析并校验 JWT
func ParseToken(tokenString string) (uint, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("签名算法不正确")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return 0, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}
	return 0, errors.New("无效的令牌")
}
