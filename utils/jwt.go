package utils

import (
	"errors" // [1] Tambahkan import ini
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type JwtClaims struct {
	UserID      string   `json:"user_id"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

func GenerateToken(userID string, role string, permissions []string) (string, error) {
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("rahasia_default_jangan_dipakai_production")
	}

	claims := JwtClaims{
		UserID:      userID,
		Role:        role,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ParseToken(tokenString string) (*JwtClaims, error) {
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("rahasia_default_jangan_dipakai_production")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid // Ini ada di v5
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// Validasi Claims
	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid {
		return claims, nil
	}

	// [2] Perbaikan: Gunakan errors.New karena jwt.ErrTokenInvalid tidak ada
	return nil, errors.New("token invalid") 
}