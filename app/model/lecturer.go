package model

import "time"

// Tabel lecturers
type Lecturer struct {
	ID         string    `json:"id" db:"id"`
	
	UserID     string    `json:"userId" db:"user_id"`
	
	// Relasi (Tidak ada di kolom database, diisi manual via JOIN)
	User       *User     `json:"user,omitempty" db:"-"`
	
	LecturerID string    `json:"lecturerId" db:"lecturer_id"` // NIP
	Department string    `json:"department" db:"department"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
}