package route

import (
	"github.com/WedhaWS/uasgosmt5/app/service"
	"github.com/WedhaWS/uasgosmt5/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(
	app *fiber.App,
	authService *service.AuthService,
	achService *service.AchievementService,
	authMiddleware *middleware.AuthMiddleware,
) {
	api := app.Group("/api/v1")

	// =================================================================
	// 5.1 Authentication
	// =================================================================
	auth := api.Group("/auth")
	auth.Post("/login", authService.Login)
	auth.Post("/refresh", authService.RefreshToken)
	auth.Post("/logout", authService.Logout)
	auth.Get("/profile", authMiddleware.AuthRequired(), authService.GetProfile)

	// =================================================================
	// 5.2 Users (Admin)
	// =================================================================
	users := api.Group("/users",
		authMiddleware.AuthRequired(),
		authMiddleware.PermissionRequired("user:manage"),
	)
	users.Get("/", authService.GetAllUsers)
	users.Get("/:id", authService.GetUserDetail)
	users.Post("/", authService.CreateUser)
	users.Put("/:id", authService.UpdateUser)
	users.Delete("/:id", authService.DeleteUser)
	users.Put("/:id/role", authService.UpdateUserRole)

	// =================================================================
	// 5.4 Achievements
	// =================================================================
	ach := api.Group("/achievements", authMiddleware.AuthRequired())

	// List (filtered by role)
	ach.Get("/", achService.GetAll)
	// Detail
	ach.Get("/:id", achService.GetDetail)
	// Create (Mahasiswa)
	ach.Post("/", authMiddleware.PermissionRequired("achievement:create"), achService.Submit)
	// Update (Mahasiswa)
	ach.Put("/:id", authMiddleware.PermissionRequired("achievement:update"), achService.Update)
	// Delete (Mahasiswa)
	ach.Delete("/:id", authMiddleware.PermissionRequired("achievement:delete"), achService.Delete)
	// Submit for verification
	ach.Post("/:id/submit", authMiddleware.PermissionRequired("achievement:create"), achService.RequestVerification)
	// Verify (Dosen Wali)
	ach.Post("/:id/verify", authMiddleware.PermissionRequired("achievement:verify"), achService.Verify)
	// Reject (Dosen Wali)
	ach.Post("/:id/reject", authMiddleware.PermissionRequired("achievement:verify"), achService.Reject)
	// Status history
	ach.Get("/:id/history", achService.GetHistory)
	// Upload files
	ach.Post("/:id/attachments", authMiddleware.PermissionRequired("achievement:update"), achService.UploadAttachment)

	// =================================================================
	// 5.5 Students & Lecturers
	// =================================================================
	students := api.Group("/students", authMiddleware.AuthRequired())
	students.Get("/", authService.GetAllStudents)
	students.Get("/:id", authService.GetStudentDetail)
	students.Get("/:id/achievements", achService.GetStudentAchievements)
	students.Put("/:id/advisor", authMiddleware.PermissionRequired("user:manage"), authService.UpdateStudentAdvisor)

	lecturers := api.Group("/lecturers", authMiddleware.AuthRequired())
	lecturers.Get("/", authService.GetAllLecturers)
	lecturers.Get("/:id/advisees", achService.GetAdviseeAchievements)

	// =================================================================
	// 5.8 Reports & Analytics
	// =================================================================
	reports := api.Group("/reports", authMiddleware.AuthRequired())
	reports.Get("/statistics", achService.GetStatistics)
	reports.Get("/student/:id", achService.GetStudentStatistics)
}
