package model

type UserCreateRequest struct {
	Username        string          `json:"username" validate:"required"`
	Email           string          `json:"email" validate:"required,email"`
	Password        string          `json:"password" validate:"required"`
	FullName        string          `json:"full_name" validate:"required"`
	RoleID          string          `json:"role_id" validate:"required"`
	StudentProfile  *StudentCreate  `json:"student_profile,omitempty"`
	LecturerProfile *LecturerCreate `json:"lecturer_profile,omitempty"`
}

type UserUpdateRequest struct {
	ID       string `json:"id"`
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	FullName string `json:"full_name" validate:"required"`
	RoleID   string `json:"role_id" validate:"required"`
}
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}

type LoginResponse struct {
	Token        string           `json:"token"`
	RefreshToken string           `json:"refresh_token"`
	User         UserAuthResponse `json:"user"`
}

type UserAuthResponse struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	FullName    string   `json:"full_name"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

type UserCreateResponse struct {
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	FullName        string    `json:"full_name"`
	RoleID          string    `json:"role_id"`
	StudentProfile  *Student  `json:"student_profile,omitempty"`
	LecturerProfile *Lecturer `json:"lecturer_profile,omitempty"`
}

type User struct {
	ID           string
	Username     string
	PasswordHash string
	Email        string
	FullName     string
	RoleId       string
	RoleName     string
	Permissions  []string
}