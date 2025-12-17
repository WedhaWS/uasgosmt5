package model

import "time"

// Tabel users
type User struct {
	ID           string    `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	FullName     string    `json:"fullName" db:"full_name"`
	
	RoleID       string    `json:"roleId" db:"role_id"`
	
	// Relasi (Tidak ada di kolom database, diisi manual via JOIN)
	Role         *Role     `json:"role,omitempty" db:"-"`
	
	IsActive     bool      `json:"isActive" db:"is_active"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}