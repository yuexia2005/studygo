package utils

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 定义一个切片，用于签名和验证JWT的密钥
var jwtSecret []byte

// 初始化 JWT 密钥（在程序启动时调用）
func InitJWTSecret() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("环境变量 JWT_SECRET 未设置")
	}
	jwtSecret = []byte(secret)
	log.Println("JWT 密钥初始化成功")
}

// 定义一个结构体，声明，过期时间等，jwt.RegisteredClaims定义了一些标准字段
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
	//结构体定义大致如下（来自库源码）：
	/*type RegisteredClaims struct {
	    	Issuer    string    `json:"iss,omitempty"`
	    	Subject   string    `json:"sub,omitempty"`
	    	Audience  ClaimStrings `json:"aud,omitempty"`
	    	ExpiresAt *NumericDate `json:"exp,omitempty"`
	    	NotBefore *NumericDate `json:"nbf,omitempty"`
	    	IssuedAt  *NumericDate `json:"iat,omitempty"`
	    	ID        string    `json:"jti,omitempty"`
		}
	*/
}

// 定义函数GenerateToken,接收userID(uint类型),返回string（token字符串）和error
func GenerateToken(userID uint) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			//设置RegisteredClaims中的ExpiresAt字段为当前时间加上24小时后的时间点（即token有效期24小时）
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// 定义函数ParseToken，接收上述生成的tokenString（JWT字符串），返回uint（用户ID）和error。
func ParseToken(tokenString string) (uint, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return 0, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}
	return 0, errors.New("invalid token")

}
