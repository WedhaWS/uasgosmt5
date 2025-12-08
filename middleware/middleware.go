package middleware

import (
	"context"
	"github.com/WedhaWS/uasgosmt5/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AuthRequired(JWTsecret []byte) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Token akses diperlukan",
			})
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Token akses tidak valid",
			})
		}

		claims, err := utils.ValidateToken(tokenParts[1], JWTsecret)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Token akses tidak valid",
			})
		}

		ctx := context.WithValue(c.UserContext(), "user", claims)
		c.SetUserContext(ctx)

		return c.Next()
	}
}