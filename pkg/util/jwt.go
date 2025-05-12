package util

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"mind-weaver/config"
)

// JWTClaims 自定义JWT Claims
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	UserRole string `json:"user_role"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(jwtInfo *JWTClaims, cfg config.JWT) (string, error) {
	// 设置过期时间
	expiresTime := time.Now().Add(time.Duration(cfg.ExpiresIn) * time.Hour)
	jwtInfo.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiresTime),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "yuewen365",
	}
	// claims := JWTClaims{
	// 	UserID:   userID,
	// 	Username: username,
	// 	RegisteredClaims: jwt.RegisteredClaims{
	// 		ExpiresAt: jwt.NewNumericDate(expiresTime),
	// 		IssuedAt:  jwt.NewNumericDate(time.Now()),
	// 		NotBefore: jwt.NewNumericDate(time.Now()),
	// 		Issuer:    "yuewen365",
	// 	},
	// }

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtInfo)
	return token.SignedString([]byte(cfg.Secret))
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string, cfg config.JWT) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
