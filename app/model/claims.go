package model

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	FullName    string   `json:"full_name"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}
