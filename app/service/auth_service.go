package service

import (
	"github.com/WedhaWS/uasgosmt5/app/model"
	"github.com/WedhaWS/uasgosmt5/app/repository"
	"github.com/WedhaWS/uasgosmt5/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type AuthService interface {
	Logout(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	RefreshToken(c *fiber.Ctx) error
}

func NewAuthService(repo repository.UserRepository, logout repository.AuthRepository, Log *logrus.Logger, secret []byte) AuthService {
	return &AuthServiceImpl{
		repo:   repo,
		Log:    Log,
		Auth:   logout,
		secret: secret,
	}
}

type AuthServiceImpl struct {
	repo     repository.UserRepository
	Auth     repository.AuthRepository
	validate *validator.Validate
	Log      *logrus.Logger
	secret   []byte
}

func (s *AuthServiceImpl) Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	ctx := c.UserContext()
	err := s.Auth.Logout(ctx, refreshToken)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	response := model.LogoutResponse{
		Message: "Logged out",
	}

	return c.JSON(model.WebResponse[model.LogoutResponse]{
		Data:   response,
		Status: "success",
	})

}

func (s *AuthServiceImpl) Login(c *fiber.Ctx) error {
	var request = new(model.LoginRequest)
	if err := c.BodyParser(request); err != nil {
		return fiber.ErrBadRequest
	}
	ctx := c.UserContext()
	User, err := s.repo.FindByUsername(ctx, request.Username)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if !utils.CheckPasswordHash(request.Password, User.PasswordHash) {
		return fiber.ErrUnauthorized
	}
	access, refresh, err := utils.GenerateToken(User, s.secret)
	if err != nil {

		return fiber.ErrInternalServerError
	}

	AuthResponse := &model.UserAuthResponse{
		ID:          User.ID,
		FullName:    User.FullName,
		Username:    User.Username,
		Role:        User.RoleName,
		Permissions: User.Permissions,
	}

	response := &model.LoginResponse{
		Token:        access,
		RefreshToken: refresh,
		User:         *AuthResponse,
	}

	return c.JSON(model.WebResponse[*model.LoginResponse]{
		Data:   response,
		Status: "success",
	})
}

func (s *AuthServiceImpl) RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	Claims, err := utils.ValidateToken(refreshToken, s.secret)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	ctx := c.UserContext()
	Access, err := s.Auth.RefreshToken(ctx, refreshToken, s.secret)
	if err != nil {
		return fiber.ErrUnauthorized
	}

	AuthResponse := &model.UserAuthResponse{
		ID:          Claims.UserID,
		FullName:    Claims.FullName,
		Username:    Claims.Username,
		Role:        Claims.Role,
		Permissions: Claims.Permissions,
	}

	response := &model.LoginResponse{
		Token:        Access,
		RefreshToken: refreshToken,
		User:         *AuthResponse,
	}

	return c.JSON(model.WebResponse[*model.LoginResponse]{
		Data:   response,
		Status: "success",
	})

}