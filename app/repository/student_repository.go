package repository

import (
	"context"
	"database/sql"
	"github.com/WedhaWS/uasgosmt5/app/model"

	"github.com/sirupsen/logrus"
)

type StudentRepository interface {
	Save(ctx context.Context, tx *sql.Tx, Student *model.Student) (*model.Student, error)
	FindAll(ctx context.Context, tx *sql.Tx) ([]model.Student, error)
	FindById(ctx context.Context, tx *sql.Tx, id string) (*model.Student, error)
}

type StudentRepositoryImpl struct {
	Log *logrus.Logger
}

func NewStudentRepository(Log *logrus.Logger) StudentRepository {
	return &StudentRepositoryImpl{
		Log: Log,
	}
}


func (repo *StudentRepositoryImpl) Save(ctx context.Context, tx *sql.Tx, Student *model.Student) (*model.Student, error) {
	// Menggunakan placeholder $n untuk PostgreSQL
	SQL := "INSERT INTO students (user_id, student_id, program_study, academic_year, advisor_id) VALUES ($1,$2,$3,$4,$5) returning id"
	err := tx.QueryRowContext(ctx, SQL, Student.UserID, Student.StudentID, Student.ProgramStudy, Student.AcademicYear, Student.AdvisorID).Scan(&Student.StudentID)
	if err != nil {
		return nil, err
	}

	return Student, nil
}

func (repo *StudentRepositoryImpl) FindAll(ctx context.Context, tx *sql.Tx) ([]model.Student, error) {
	//TODO implement me
	panic("implement me")
}

func (repo *StudentRepositoryImpl) FindById(ctx context.Context, tx *sql.Tx, id string) (*model.Student, error) {
	//TODO implement me
	panic("implement me")
}