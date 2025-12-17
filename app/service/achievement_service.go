package service

import (
	"fmt"
	"math"
	"time"
	"github.com/WedhaWS/uasgosmt5/app/model"
	"github.com/WedhaWS/uasgosmt5/app/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gofiber/fiber/v2"
)

type AchievementService struct {
	achRepo  *repository.AchievementRepository
	userRepo *repository.UserRepository
}

func NewAchievementService(achRepo *repository.AchievementRepository, userRepo *repository.UserRepository) *AchievementService {
	return &AchievementService{
		achRepo:  achRepo,
		userRepo: userRepo,
	}
}

// ---------------------------------------------------------
// 5.4 ACHIEVEMENTS
// ---------------------------------------------------------

// POST /api/v1/achievements (FR-003: Submit Prestasi - Draft)
func (s *AchievementService) Submit(c *fiber.Ctx) error {
	// 1. Parse Input
	var req model.Achievement
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid input: " + err.Error()})
	}

	// 2. Ambil User ID dari Token (Middleware)
	userID := c.Locals("user_id").(string)

	// Validasi: Pastikan User adalah Mahasiswa dan punya profil Student
	student, err := s.userRepo.FindStudentByUserID(userID)
	if err != nil {
		// Error 404 ini muncul jika user login tapi tidak ada di tabel 'students'
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Student profile not found. Are you a student?"})
	}

	// 3. Siapkan Data MongoDB (Konten Detail)
	mongoData := model.Achievement{
		ID:              primitive.NewObjectID(), // Generate ID baru untuk Mongo
		StudentID:       student.ID,              // Link ke ID Student (Postgres UUID)
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         req.Details,     // Field dinamis (Juara, Lomba, dll)
		Attachments:     req.Attachments, // File bukti
		Tags:            req.Tags,
		Points:          req.Points, // [UPDATED] Mengambil poin dari input user (sebelumnya 0)
	}

	// 4. Siapkan Data PostgreSQL (Referensi Status)
	pgData := model.AchievementReference{
		StudentID: student.ID,
		Title:     req.Title, // Disimpan juga di SQL untuk searching/sorting
		Status:    "draft",   // Status Awal sesuai FR-003
	}

	// 5. Simpan ke Database (Hybrid Transaction di Repository)
	if err := s.achRepo.Create(c.Context(), &mongoData, &pgData); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to submit achievement: " + err.Error()})
	}

	// 6. Return Success Response
	return c.Status(201).JSON(model.WebResponse{
		Code:    201,
		Status:  "success",
		Message: "Prestasi berhasil disimpan sebagai draft",
		Data: fiber.Map{
			"referenceId": pgData.ID,                 // ID dari Postgres
			"mongoId":     pgData.MongoAchievementID, // ID dari Mongo
			"status":      pgData.Status,
			"points":      mongoData.Points, // Tampilkan poin yang tersimpan
		},
	})
}

// GET /api/v1/achievements
func (s *AchievementService) GetAll(c *fiber.Ctx) error {
	param := s.parsePagination(c)
	userRole := c.Locals("role").(string)
	userID := c.Locals("user_id").(string)

	var filterStudent, filterAdvisor string

	// Logic Filter Berdasarkan Role (RBAC Data Level)
	if userRole == "Mahasiswa" {
		// Mahasiswa HANYA boleh melihat prestasi miliknya sendiri
		student, _ := s.userRepo.FindStudentByUserID(userID)
		if student != nil {
			filterStudent = student.ID
		}
	} else if userRole == "Dosen Wali" {
		// Dosen Wali HANYA boleh melihat prestasi mahasiswa bimbingannya (FR-006)
		lecturer, _ := s.userRepo.FindLecturerByUserID(userID)
		if lecturer != nil {
			filterAdvisor = lecturer.ID
		}
	}
	// Admin melihat semua (filter kosong)

	data, total, err := s.achRepo.FindAll(param, filterStudent, filterAdvisor)
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	return s.sendPaginationResponse(c, data, total, param)
}

// GET /api/v1/achievements/:id
func (s *AchievementService) GetDetail(c *fiber.Ctx) error {
	id := c.Params("id")

	// Validasi Kepemilikan (Optional tapi disarankan)
	// Idealnya dicek apakah user berhak melihat detail ini

	ref, content, err := s.achRepo.FindDetail(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Achievement not found"})
	}
	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "Detail retrieved",
		Data: fiber.Map{
			"meta":    ref,
			"content": content,
		},
	})
}

// POST /api/v1/achievements/:id/submit (FR-004: Submit untuk Verifikasi)
func (s *AchievementService) RequestVerification(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	// 1. Ambil Profil Mahasiswa
	student, err := s.userRepo.FindStudentByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Student profile not found"})
	}

	// 2. Cek Existensi Achievement
	ref, _, err := s.achRepo.FindDetail(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Achievement not found"})
	}

	// 3. Validasi Kepemilikan (Harus milik mahasiswa yang login)
	if ref.StudentID != student.ID {
		return c.Status(403).JSON(model.WebResponse{Code: 403, Status: "error", Message: "Unauthorized: You do not own this achievement"})
	}

	// 4. Validasi Status (Hanya 'draft' yang bisa disubmit)
	if ref.Status != "draft" {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Only draft achievement can be submitted for verification"})
	}

	// 5. Update Status menjadi 'submitted'
	if err := s.achRepo.UpdateStatus(id, "submitted", "", "", 0); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to update status: " + err.Error()})
	}

	// 6. Create Notification untuk Dosen Wali
	if student.AdvisorID != nil && *student.AdvisorID != "" {
		// Log notification (bisa diganti dengan email/push notification service)
		fmt.Printf("[NOTIFICATION] Prestasi baru untuk verifikasi:\n")
		fmt.Printf("  - Student: %s (%s)\n", student.User.FullName, student.StudentID)
		fmt.Printf("  - Achievement: %s\n", ref.Title)
		fmt.Printf("  - Advisor ID: %s\n", *student.AdvisorID)
		fmt.Printf("  - Status: submitted\n")
		fmt.Printf("  - Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))

		// TODO: Implement actual notification service
		// - Email notification ke dosen wali
		// - Push notification ke mobile app
		// - In-app notification system
		// - SMS notification (optional)
	} else {
		fmt.Printf("[WARNING] Student %s tidak memiliki dosen wali yang ditugaskan\n", student.StudentID)
	}

	// 7. Return Updated Status
	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "Prestasi berhasil diajukan untuk verifikasi",
		Data: fiber.Map{
			"id":     id,
			"status": "submitted",
		},
	})
}

// POST /api/v1/achievements/:id/verify (FR-007: Verify Prestasi)
func (s *AchievementService) Verify(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Points int `json:"points"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid input: " + err.Error()})
	}

	// Validate points (must be positive)
	if req.Points <= 0 {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Points must be greater than 0"})
	}

	userID := c.Locals("user_id").(string)

	// 1. Pastikan user adalah Dosen Wali
	lecturer, err := s.userRepo.FindLecturerByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Lecturer profile not found"})
	}

	// 2. Dosen review prestasi detail - Get achievement info
	ref, content, err := s.achRepo.FindDetail(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Achievement not found"})
	}

	// 3. Validasi precondition: Status harus 'submitted'
	if ref.Status != "submitted" {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Only submitted achievements can be verified"})
	}

	// 4. Validasi ownership: Hanya dosen wali yang bisa verify prestasi mahasiswa bimbingannya
	if ref.Student.AdvisorID == nil || *ref.Student.AdvisorID != lecturer.ID {
		return c.Status(403).JSON(model.WebResponse{Code: 403, Status: "error", Message: "Unauthorized: You can only verify achievements of your advisees"})
	}

	// 5. Update status menjadi 'verified' dengan verified_by dan verified_at
	if err := s.achRepo.UpdateStatus(id, "verified", userID, "", req.Points); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to verify achievement: " + err.Error()})
	}

	// 6. Log verification
	fmt.Printf("[VERIFICATION] Achievement verified:\n")
	fmt.Printf("  - Student: %s (%s)\n", ref.Student.User.FullName, ref.Student.StudentID)
	fmt.Printf("  - Achievement: %s\n", ref.Title)
	fmt.Printf("  - Points awarded: %d\n", req.Points)
	fmt.Printf("  - Verified by: %s\n", lecturer.User.FullName)
	fmt.Printf("  - Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	// 7. Return updated status
	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "Prestasi berhasil diverifikasi",
		Data: fiber.Map{
			"id":         id,
			"status":     "verified",
			"points":     req.Points,
			"verifiedBy": userID,
			"verifiedAt": time.Now(),
			"achievement": fiber.Map{
				"title":     ref.Title,
				"type":      content.AchievementType,
				"student":   ref.Student.User.FullName,
				"studentId": ref.Student.StudentID,
			},
		},
	})
}

// POST /api/v1/achievements/:id/reject (FR-008: Reject Prestasi)
func (s *AchievementService) Reject(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Note string `json:"note"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid input: " + err.Error()})
	}

	// Validate rejection note (required and not empty)
	if req.Note == "" {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Rejection note is required"})
	}

	// Validate note length (reasonable limit)
	if len(req.Note) > 1000 {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Rejection note too long (max 1000 characters)"})
	}

	userID := c.Locals("user_id").(string)

	// 1. Pastikan user adalah Dosen Wali
	lecturer, err := s.userRepo.FindLecturerByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Lecturer profile not found"})
	}

	// 2. Get achievement info untuk validasi
	ref, content, err := s.achRepo.FindDetail(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Achievement not found"})
	}

	// 3. Validasi precondition: Status harus 'submitted'
	if ref.Status != "submitted" {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Only submitted achievements can be rejected"})
	}

	// 4. Validasi ownership: Hanya dosen wali yang bisa reject prestasi mahasiswa bimbingannya
	if ref.Student.AdvisorID == nil || *ref.Student.AdvisorID != lecturer.ID {
		return c.Status(403).JSON(model.WebResponse{Code: 403, Status: "error", Message: "Unauthorized: You can only reject achievements of your advisees"})
	}

	// 5. Update status menjadi 'rejected' dengan rejection_note
	if err := s.achRepo.UpdateStatus(id, "rejected", userID, req.Note, 0); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to reject achievement: " + err.Error()})
	}

	// 6. Create notification untuk mahasiswa
	fmt.Printf("[REJECTION] Achievement rejected:\n")
	fmt.Printf("  - Student: %s (%s)\n", ref.Student.User.FullName, ref.Student.StudentID)
	fmt.Printf("  - Achievement: %s\n", ref.Title)
	fmt.Printf("  - Rejection reason: %s\n", req.Note)
	fmt.Printf("  - Rejected by: %s\n", lecturer.User.FullName)
	fmt.Printf("  - Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	// TODO: Implement actual notification service
	// - Email notification ke mahasiswa
	// - Push notification ke mobile app
	// - In-app notification system

	// 7. Return updated status
	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "Prestasi berhasil ditolak",
		Data: fiber.Map{
			"id":            id,
			"status":        "rejected",
			"rejectionNote": req.Note,
			"rejectedBy":    userID,
			"rejectedAt":    time.Now(),
			"achievement": fiber.Map{
				"title":     ref.Title,
				"type":      content.AchievementType,
				"student":   ref.Student.User.FullName,
				"studentId": ref.Student.StudentID,
			},
		},
	})
}

// DELETE /api/v1/achievements/:id (FR-005: Soft Delete Draft)
func (s *AchievementService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	// 1. Validasi user adalah mahasiswa
	student, err := s.userRepo.FindStudentByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Student profile not found"})
	}

	// 2. Cek existensi dan kepemilikan achievement
	ref, _, err := s.achRepo.FindDetail(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Achievement not found"})
	}

	// 3. Validasi kepemilikan
	if ref.StudentID != student.ID {
		return c.Status(403).JSON(model.WebResponse{Code: 403, Status: "error", Message: "Unauthorized: You do not own this achievement"})
	}

	// 4. Validasi status (hanya draft yang bisa dihapus)
	if ref.Status != "draft" {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Only draft achievements can be deleted"})
	}

	// 5. Soft delete achievement
	if err := s.achRepo.Delete(c.Context(), id); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to delete achievement: " + err.Error()})
	}

	// 6. Log deletion
	fmt.Printf("[SOFT DELETE] Achievement deleted:\n")
	fmt.Printf("  - Student: %s (%s)\n", student.User.FullName, student.StudentID)
	fmt.Printf("  - Achievement: %s\n", ref.Title)
	fmt.Printf("  - Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "Prestasi berhasil dihapus",
		Data: fiber.Map{
			"id":     id,
			"status": "deleted",
		},
	})
}

// ... (Method stub/placeholder lain seperti Update, GetHistory, UploadAttachment biarkan saja seperti sebelumnya) ...

func (s *AchievementService) Update(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Update Feature Not Implemented"})
}

func (s *AchievementService) GetHistory(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "History Not Implemented"})
}

func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
	achievementID := c.Params("id")

	// Parse multipart form
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "No file uploaded"})
	}

	// Validate file size (max 2MB)
	if file.Size > 2*1024*1024 {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "File size exceeds 2MB limit"})
	}

	// Validate file type
	allowedTypes := map[string]bool{
		"image/jpeg": true, "image/png": true, "image/jpg": true,
		"application/pdf": true, "application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	}

	if !allowedTypes[file.Header.Get("Content-Type")] {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid file type"})
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	filepath := fmt.Sprintf("./uploads/%s", filename)

	// Save file
	if err := c.SaveFile(file, filepath); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to save file"})
	}

	// Create attachment object
	attachment := model.AchievementAttachment{
		FileName:   file.Filename,
		FileURL:    fmt.Sprintf("/uploads/%s", filename),
		FileType:   file.Header.Get("Content-Type"),
		UploadedAt: time.Now(),
	}

	// Update achievement in MongoDB
	if err := s.achRepo.AddAttachment(c.Context(), achievementID, attachment); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to update achievement"})
	}

	return c.JSON(model.WebResponse{
		Code: 200, Status: "success",
		Message: "File uploaded successfully",
		Data:    attachment,
	})
}

func (s *AchievementService) GetAdviseeAchievements(c *fiber.Ctx) error {
	// 1. Ambil User ID Dosen dari Token
	userID := c.Locals("user_id").(string)

	// 2. Cari Profil Dosen (Lecturer) berdasarkan User ID
	lecturer, err := s.userRepo.FindLecturerByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Lecturer profile not found"})
	}

	// 3. Parse Parameter Pagination (Page, Limit, Search, Sort)
	param := s.parsePagination(c)

	// 4. Panggil Repo FindAll dengan Filter AdvisorID
	data, total, err := s.achRepo.FindAll(param, "", lecturer.ID)
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	// 5. Return Response dengan Pagination
	return s.sendPaginationResponse(c, data, total, param)
}

func (s *AchievementService) GetStudentAchievements(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Student Achievement List Not Implemented"})
}

// ---------------------------------------------------------
// 5.8 REPORTS & ANALYTICS
// ---------------------------------------------------------

func (s *AchievementService) GetStatistics(c *fiber.Ctx) error {
	userRole := c.Locals("role").(string)
	userID := c.Locals("user_id").(string)

	// Role-based statistics
	if userRole == "Dosen Wali" {
		// Dosen Wali hanya melihat statistik mahasiswa bimbingannya
		return s.getAdvisorStatistics(c, userID)
	} else if userRole == "Mahasiswa" {
		// Mahasiswa hanya melihat statistik sendiri
		student, err := s.userRepo.FindStudentByUserID(userID)
		if err != nil {
			return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Student profile not found"})
		}
		stats, err := s.achRepo.GetStudentStatistics(c.Context(), student.ID)
		if err != nil {
			return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to generate stats: " + err.Error()})
		}
		return c.JSON(model.WebResponse{Code: 200, Status: "success", Message: "Your Statistics", Data: stats})
	}

	// Admin melihat statistik keseluruhan
	stats, err := s.achRepo.GetStatistics(c.Context())
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to generate stats: " + err.Error()})
	}
	return c.JSON(model.WebResponse{Code: 200, Status: "success", Message: "Overall Statistics", Data: stats})
}

// Helper method untuk statistik dosen wali
func (s *AchievementService) getAdvisorStatistics(c *fiber.Ctx, userID string) error {
	// Get lecturer profile
	_, err := s.userRepo.FindLecturerByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Lecturer profile not found"})
	}

	// TODO: Implement advisor-specific statistics
	// This would aggregate statistics from all advisee students
	// For now, return overall stats (can be enhanced later)
	stats, err := s.achRepo.GetStatistics(c.Context())
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to generate stats: " + err.Error()})
	}

	return c.JSON(model.WebResponse{
		Code: 200, Status: "success",
		Message: "Advisee Statistics",
		Data:    stats,
	})
}

func (s *AchievementService) GetStudentStatistics(c *fiber.Ctx) error {
	// ID di sini bisa berupa UserID atau StudentID.
	requestID := c.Params("id") // Bisa "me" atau UUID
	userID := c.Locals("user_id").(string)
	userRole := c.Locals("role").(string)

	var targetStudentID string

	if requestID == "me" || (userRole == "Mahasiswa" && requestID == userID) {
		// Lihat statistik diri sendiri
		student, err := s.userRepo.FindStudentByUserID(userID)
		if err != nil {
			return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Student profile not found"})
		}
		targetStudentID = student.ID
	} else if userRole == "Admin" || userRole == "Dosen Wali" {
		// Admin/Dosen lihat statistik mahasiswa lain
		student, err := s.userRepo.FindStudentByUserID(requestID)
		if err != nil {
			return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Student not found"})
		}
		targetStudentID = student.ID
	} else {
		return c.Status(403).JSON(model.WebResponse{Code: 403, Status: "error", Message: "Forbidden"})
	}

	stats, err := s.achRepo.GetStudentStatistics(c.Context(), targetStudentID)
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "Student Statistics",
		Data:    stats,
	})
}

// HELPER
func (s *AchievementService) parsePagination(c *fiber.Ctx) model.PaginationParam {
	return model.PaginationParam{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),
		SortBy: c.Query("sortBy", "created_at"),
		Order:  c.Query("order", "desc"),
		Search: c.Query("search", ""),
	}
}

func (s *AchievementService) sendPaginationResponse(c *fiber.Ctx, data interface{}, total int64, param model.PaginationParam) error {
	totalPages := int(math.Ceil(float64(total) / float64(param.Limit)))
	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "Data retrieved",
		Data:    data,
		Meta: &model.MetaInfo{
			Page:      param.Page,
			Limit:     param.Limit,
			TotalData: total,
			TotalPage: totalPages,
			SortBy:    param.SortBy,
			Order:     param.Order,
			Search:    param.Search,
		},
	})
}
