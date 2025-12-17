package model

import "time"

// Tabel students
type Student struct {
	ID           string    `json:"id" db:"id"`
	
	UserID       string    `json:"userId" db:"user_id"`
	
	// Relasi (Tidak ada di kolom database, diisi manual via JOIN)
	User         *User     `json:"user,omitempty" db:"-"`
	
	// Kolom student_id di Database (NIM)
	StudentID    string    `json:"studentId" db:"student_id"` 
	
	ProgramStudy string    `json:"programStudy" db:"program_study"`
	AcademicYear string    `json:"academicYear" db:"academic_year"`
	
	AdvisorID    *string   `json:"advisorId" db:"advisor_id"`
	
	// Relasi (Tidak ada di kolom database)
	Advisor      *Lecturer `json:"advisor,omitempty" db:"-"`
	
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}