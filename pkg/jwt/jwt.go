package jwt

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
	UserID   string `json:"userid"`
	UserName string `json:"username"`
	jwt.StandardClaims
}

var jwtSecret = []byte("shortURL-secret-key")

// GenerateToken 生成Token
func GenerateToken(userId, userName string) (string, error) {
	claims := &Claims{
		UserID:   userId,
		UserName: userName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // 过期时间为24小时
			Issuer:    "AnYyy",                               // 签发者
			IssuedAt:  time.Now().Unix(),                     // 签发时间
		},
	}

	// 生成包含身份等信息的Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtSecret) // 加密Token并返回
}

// ParseToken 解析请求头中的Token
func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil // 得到密钥验证签名没有被篡改
		},
	)

	// 进行类型断言
	if claims, r := token.Claims.(*Claims); r {
		return claims, nil
	}
	return nil, err
}
