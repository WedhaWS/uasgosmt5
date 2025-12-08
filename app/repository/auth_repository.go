package repository

import (
	"context"
	"errors"
	"github.com/WedhaWS/uasgosmt5/app/model"
	"github.com/WedhaWS/uasgosmt5/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type AuthRepository interface {
	Logout(ctx context.Context, RefreshToken string) error
	RefreshToken(ctx context.Context, RefreshToken string, jwtKey []byte) (string, error)
}

type AuthRepositoryImplements struct {
	DB  *redis.Client
	Log *logrus.Logger
}

func NewLogoutRepository(DB *redis.Client, Log *logrus.Logger) AuthRepository {
	return &AuthRepositoryImplements{
		DB:  DB,
		Log: Log,
	}
}

func (l *AuthRepositoryImplements) Logout(ctx context.Context, RefreshToken string) error {
	exp, err := utils.ExtractExpiration(RefreshToken)
	if err != nil {
		return err
	}

	ttl := time.Until(exp)

	return l.DB.Set(ctx, "blacklist:"+RefreshToken, "1", ttl).Err()
}

func (l *AuthRepositoryImplements) RefreshToken(ctx context.Context, refreshToken string, jwtKey []byte) (string, error) {

	claims, err := utils.ValidateToken(refreshToken, jwtKey)
	if err != nil {
		return "", err
	}

	isBlacklisted, _ := l.DB.Get(ctx, "blacklist:"+refreshToken).Result()
	if isBlacklisted != "" {
		return "", errors.New("refresh token banned")
	}

	accessExp := time.Now().Add(15 * time.Minute)

	newClaims := model.Claims{
		UserID:      claims.UserID,
		Username:    claims.Username,
		FullName:    claims.FullName,
		Role:        claims.Role,
		Permissions: claims.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExp),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	accessString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return accessString, nil
}