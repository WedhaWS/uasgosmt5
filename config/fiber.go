package config

import (
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

func NewFiber(config *viper.Viper) *fiber.App {
	app := fiber.New(fiber.Config{AppName: config.GetString("app.name"),
		ErrorHandler: ErrorHandler(),
		Prefork:      config.GetBool("app.prefork")})

	return app
}

func ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		return c.Status(code).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
}