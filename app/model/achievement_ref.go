package model

import "time"

// Tabel achievement_references
type AchievementReference struct {
	ID                 string     `json:"id" db:"id"`

	// Foreign Key merujuk ke students.id (UUID)
	StudentID          string     `json:"studentId" db:"student_id"`
	
	// Relasi (Tidak ada di kolom database, diisi lewat JOIN manual)
	Student            *Student   `json:"student,omitempty" db:"-"`

	MongoAchievementID string     `json:"mongoAchievementId" db:"mongo_achievement_id"`

	// Field Title (Tambahan Modul 6 Search/Sort)
	Title              string     `json:"title" db:"title"`

	// Enum (draft, submitted, verified, rejected)
	Status             string     `json:"status" db:"status"`

	SubmittedAt        *time.Time `json:"submittedAt" db:"submitted_at"`
	VerifiedAt         *time.Time `json:"verifiedAt" db:"verified_at"`

	VerifiedBy         *string    `json:"verifiedBy" db:"verified_by"`
	
	// Relasi (Tidak ada di kolom database)
	Verifier           *User      `json:"verifier,omitempty" db:"-"`

	RejectionNote      string     `json:"rejectionNote" db:"rejection_note"`
	CreatedAt          time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time  `json:"updatedAt" db:"updated_at"`
}