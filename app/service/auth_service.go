package service

import (
	"github.com/WedhaWS/uasgosmt5/app/model"
	"github.com/WedhaWS/uasgosmt5/app/repository"
	"github.com/WedhaWS/uasgosmt5/utils"

	"github.com/gofiber/fiber/v2"
)

type AuthService struct {
	userRepo *repository.UserRepository
	roleRepo *repository.RoleRepository
}

func NewAuthService(userRepo *repository.UserRepository, roleRepo *repository.RoleRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// =================================================================
// 5.1 AUTHENTICATION
// =================================================================

// POST /api/v1/auth/login
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid request body"})
	}

	// 1. Cari User
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return c.Status(401).JSON(model.WebResponse{Code: 401, Status: "error", Message: "Invalid email or password"})
	}

	// 2. Cek Password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return c.Status(401).JSON(model.WebResponse{Code: 401, Status: "error", Message: "Invalid email or password"})
	}

	// 3. Cek Status Aktif
	if !user.IsActive {
		return c.Status(403).JSON(model.WebResponse{Code: 403, Status: "error", Message: "User account is inactive"})
	}

	// 4. Ambil Permissions dari Role
	permsData, err := s.roleRepo.GetPermissionsByRoleID(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to load permissions"})
	}

	var permissions []string
	for _, p := range permsData {
		permissions = append(permissions, p.Name)
	}

	// 5. Generate Token
	token, err := utils.GenerateToken(user.ID, user.Role.Name, permissions)
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to generate token"})
	}

	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "Login successful",
		Data: fiber.Map{
			"token": token,
			"user": fiber.Map{
				"id":          user.ID,
				"username":    user.Username,
				"fullName":    user.FullName,
				"role":        user.Role.Name,
				"permissions": permissions,
			},
		},
	})
}

// POST /api/v1/auth/refresh
func (s *AuthService) RefreshToken(c *fiber.Ctx) error {
	return c.Status(501).JSON(model.WebResponse{Code: 501, Status: "error", Message: "Refresh Token Not Implemented"})
}

// POST /api/v1/auth/logout
func (s *AuthService) Logout(c *fiber.Ctx) error {
	// Karena menggunakan JWT (Stateless), logout cukup dilakukan di client side (hapus token).
	return c.JSON(model.WebResponse{Code: 200, Status: "success", Message: "Logged out successfully"})
}

// GET /api/v1/auth/profile
func (s *AuthService) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "User not found"})
	}

	user.PasswordHash = "" // Hide sensitive data

	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "Profile retrieved",
		Data:    user,
	})
}

// =================================================================
// 5.2 USERS MANAGEMENT (ADMIN)
// =================================================================

// GET /api/v1/users
func (s *AuthService) GetAllUsers(c *fiber.Ctx) error {
	users, err := s.userRepo.FindAll()
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "All users retrieved",
		Data:    users,
	})
}

// GET /api/v1/users/:id
func (s *AuthService) GetUserDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "User not found"})
	}

	user.PasswordHash = ""
	return c.JSON(model.WebResponse{
		Code:    200,
		Status:  "success",
		Message: "User detail retrieved",
		Data:    user,
	})
}

// POST /api/v1/users
func (s *AuthService) CreateUser(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"fullName"`
		RoleID   string `json:"roleId"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid input data"})
	}

	hashedPwd, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to hash password"})
	}

	user := model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPwd,
		FullName:     req.FullName,
		RoleID:       req.RoleID,
		IsActive:     true, // Default active
	}

	if err := s.userRepo.Create(&user); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	user.PasswordHash = ""
	return c.Status(201).JSON(model.WebResponse{
		Code:    201,
		Status:  "success",
		Message: "User created successfully",
		Data:    user,
	})
}

// PUT /api/v1/users/:id
func (s *AuthService) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	
	// Struct input parsial untuk update
	var req struct {
		FullName string `json:"fullName"`
		Username string `json:"username"`
		Email    string `json:"email"`
		IsActive bool   `json:"isActive"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid input"})
	}

	// 1. Cari user dulu untuk memastikan ada
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "User not found"})
	}

	// 2. Update field
	user.FullName = req.FullName
	user.Username = req.Username
	user.Email = req.Email
	user.IsActive = req.IsActive

	// 3. Simpan perubahan
	if err := s.userRepo.Update(user); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	user.PasswordHash = "" // Hide sensitive data
	return c.JSON(model.WebResponse{Code: 200, Status: "success", Message: "User updated successfully", Data: user})
}

// DELETE /api/v1/users/:id
func (s *AuthService) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	
	// Prevent delete self (Optional logic)
	myID := c.Locals("user_id").(string)
	if id == myID {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Cannot delete yourself"})
	}

	if err := s.userRepo.Delete(id); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to delete user: " + err.Error()})
	}

	return c.JSON(model.WebResponse{Code: 200, Status: "success", Message: "User deleted successfully"})
}

// PUT /api/v1/users/:id/role
func (s *AuthService) UpdateUserRole(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct { RoleID string `json:"roleId"` }
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid input"})
	}

	if err := s.userRepo.UpdateRole(id, req.RoleID); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	return c.JSON(model.WebResponse{Code: 200, Status: "success", Message: "Role assigned successfully"})
}

// =================================================================
// 5.5 STUDENTS & LECTURERS (PROFILING)
// =================================================================

func (s *AuthService) GetAllStudents(c *fiber.Ctx) error {
	students, err := s.userRepo.FindAllStudents()
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}
	return c.JSON(model.WebResponse{Code: 200, Status: "success", Data: students})
}

// POST /api/v1/students (Set Student Profile - Admin Only)
func (s *AuthService) SetStudentProfile(c *fiber.Ctx) error {
	var req struct {
		UserID       string `json:"userId"`
		StudentID    string `json:"studentId"` // NIM
		ProgramStudy string `json:"programStudy"`
		AcademicYear string `json:"academicYear"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid input"})
	}

	student := model.Student{
		UserID:       req.UserID,
		StudentID:    req.StudentID,
		ProgramStudy: req.ProgramStudy,
		AcademicYear: req.AcademicYear,
	}

	if err := s.userRepo.SaveStudent(&student); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to set profile: " + err.Error()})
	}

	return c.JSON(model.WebResponse{Code: 200, Status: "success", Message: "Student profile updated"})
}

// GET /api/v1/students/:id (Detail Student)
func (s *AuthService) GetStudentDetail(c *fiber.Ctx) error {
	// Di repo FindStudentByUserID mengambil berdasarkan user_id, 
	// tapi biasanya route /students/:id mengirim user_id (karena lebih umum dipakai untuk admin).
	// Jika :id adalah ID tabel students, kita butuh method repo lain.
	// Asumsi saat ini :id adalah user_id.
	userID := c.Params("id")
	
	student, err := s.userRepo.FindStudentByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(model.WebResponse{Code: 404, Status: "error", Message: "Student profile not found"})
	}
	
	return c.JSON(model.WebResponse{Code: 200, Status: "success", Data: student})
}

// PUT /api/v1/students/:id/advisor (Set Advisor)
func (s *AuthService) UpdateStudentAdvisor(c *fiber.Ctx) error {
	studentUUID := c.Params("id") // Ini ID tabel Students (UUID Primary Key)
	var req struct { AdvisorID string `json:"advisorId"` } // ID tabel Lecturers
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid input"})
	}

	if err := s.userRepo.AssignAdvisor(studentUUID, req.AdvisorID); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}

	return c.JSON(model.WebResponse{Code: 200, Status: "success", Message: "Advisor assigned successfully"})
}

func (s *AuthService) GetAllLecturers(c *fiber.Ctx) error {
	lecturers, err := s.userRepo.FindAllLecturers()
	if err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: err.Error()})
	}
	return c.JSON(model.WebResponse{Code: 200, Status: "success", Data: lecturers})
}

// POST /api/v1/lecturers (Set Lecturer Profile - Admin Only)
func (s *AuthService) SetLecturerProfile(c *fiber.Ctx) error {
	var req struct {
		UserID     string `json:"userId"`
		LecturerID string `json:"lecturerId"` // NIP
		Department string `json:"department"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.WebResponse{Code: 400, Status: "error", Message: "Invalid input"})
	}

	lecturer := model.Lecturer{
		UserID:     req.UserID,
		LecturerID: req.LecturerID,
		Department: req.Department,
	}

	if err := s.userRepo.SaveLecturer(&lecturer); err != nil {
		return c.Status(500).JSON(model.WebResponse{Code: 500, Status: "error", Message: "Failed to set profile: " + err.Error()})
	}

	return c.JSON(model.WebResponse{Code: 200, Status: "success", Message: "Lecturer profile updated"})
}