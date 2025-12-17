package config

import (
	"os"
	"github.com/WedhaWS/uasgosmt5/helper" // Pastikan ini sesuai dengan nama modul Anda

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// NewFiberApp membuat dan mengonfigurasi instance aplikasi Fiber baru.
func NewFiberApp() *fiber.App {
	app := fiber.New(fiber.Config{
		// 1. Menambahkan ErrorHandler kustom
		// Ini akan menangkap semua error yang dikembalikan oleh service Anda
		// dan mengubahnya menjadi format JSON yang rapi menggunakan helper.ErrorResponse.
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError // Default 500
			if e, ok := err.(*fiber.Error); ok {
				// Menangkap error bawaan Fiber (spt 404 Not Found)
				code = e.Code
			}
			return ctx.Status(code).JSON(helper.ErrorResponse{
				Success: false,
				Message: err.Error(),
			})
		},

		// 2. Menambahkan BodyLimit (sesuai Modul 9)
		// Menaikkan batas default agar bisa menerima upload file.
		// 5 * 1024 * 1024 = 5MB. Ini harus lebih besar dari file terbesar Anda (2MB).
		BodyLimit: 5 * 1024 * 1024,
	})

	// Menggunakan middleware recover untuk mencegah crash
	app.Use(recover.New())

	// Menggunakan middleware CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, PATCH", // PATCH mungkin diperlukan untuk restore
	}))

	// Menggunakan middleware logger
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${ip}:${port} ${status} - ${method} ${path}\n",
		TimeFormat: "02-Jan-2006 15:04:05",
		Output:     os.Stdout,
	}))

	return app
}

