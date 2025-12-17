package main

import (
	"log"
	"os"

	"github.com/WedhaWS/uasgosmt5/app/repository"
	"github.com/WedhaWS/uasgosmt5/app/service"
	"github.com/WedhaWS/uasgosmt5/database"
	"github.com/WedhaWS/uasgosmt5/middleware"
	"github.com/WedhaWS/uasgosmt5/route"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	
)

func main() {
	// 1. Load Environment Variables
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: .env file not found, using system environment variables")
	}

	// 2. Initialize Database (Hybrid: Postgres & Mongo)
	// Config ini otomatis melakukan Manual Migration untuk Postgres (Native SQL)
	// Return type: *config.DatabaseInstances { Postgres: *sql.DB, Mongo: *mongo.Database }
	db := database.InitDB()

	// 3. Setup Repositories (Data Access Layer)
	// ---------------------------------------------------------
	// RoleRepo: Menggunakan *sql.DB (Postgres)
	roleRepo := repository.NewRoleRepository(db.Postgres)

	// UserRepo: Menggunakan *sql.DB (Postgres)
	userRepo := repository.NewUserRepository(db.Postgres)

	// AchRepo: Butuh DUA koneksi (Postgres *sql.DB & Mongo *mongo.Database)
	achRepo := repository.NewAchievementRepository(db.Postgres, db.Mongo)

	// 4. Setup Services (Business Logic Layer)
	// ---------------------------------------------------------
	// AuthService: Butuh UserRepo & RoleRepo
	authService := service.NewAuthService(userRepo, roleRepo)

	// AchService: Butuh AchRepo & UserRepo
	achService := service.NewAchievementService(achRepo, userRepo)

	// 5. Setup Middleware
	// ---------------------------------------------------------
	// AuthMiddleware: Butuh RoleRepo untuk validasi permission
	authMiddleware := middleware.NewAuthMiddleware(roleRepo)

	// 6. Initialize Fiber App
	// ---------------------------------------------------------
	app := fiber.New(fiber.Config{
		// Custom Error Handler
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"code":    code,
				"status":  "error",
				"message": err.Error(),
			})
		},
	})

	// Default Middlewares
	app.Use(logger.New()) // Log request ke terminal
	app.Use(cors.New())   // Enable CORS

	app.Static("/docs", "./docs")

	// 7. Static File Serving for Uploads
	// ---------------------------------------------------------
	app.Static("/uploads", "./uploads")

	// 8. Setup Routes (Wiring Semua Komponen)
	// ---------------------------------------------------------
	// Mengirimkan app, services, dan middleware ke router
	route.SetupRoutes(app, authService, achService, authMiddleware)

	// 9. Start Server
	// ---------------------------------------------------------
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("🚀 Server running on port " + port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}
