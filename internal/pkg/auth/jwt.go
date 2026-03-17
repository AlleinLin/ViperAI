package auth

import (
	"time"

	"viperai/internal/config"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID  int64  `json:"user_id"`
	Account string `json:"account"`
	jwt.RegisteredClaims
}

func GenerateToken(userID int64, account string) (string, error) {
	cfg := config.Get().Auth

	claims := Claims{
		UserID:  userID,
		Account: account,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.Duration) * time.Hour)),
			Issuer:    cfg.Issuer,
			Subject:   cfg.Subject,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

func ParseToken(tokenString string) (*Claims, bool) {
	cfg := config.Get().Auth

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})

	if err != nil || !token.Valid {
		return nil, false
	}

	return claims, true
}
