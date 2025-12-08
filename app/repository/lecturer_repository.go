package repository

import (
	"context"
	"database/sql"
	"github.com/WedhaWS/uasgosmt5/app/model"

	"github.com/sirupsen/logrus"
)

type LecturerRepository interface {
	Save(ctx context.Context, tx *sql.Tx, Lecturer *model.Lecturer) (*model.Lecturer, error)
	FindAll(ctx context.Context, tx *sql.Tx) ([]model.Lecturer, error)
	FindById(ctx context.Context, tx *sql.Tx, id string) (*model.Lecturer, error)
}

type LecturerRepositoryImpl struct {
	Log *logrus.Logger
}

func NewLecturerRepository(Log *logrus.Logger) LecturerRepository {
	return &LecturerRepositoryImpl{
		Log: Log,
	}
}


func (repo *LecturerRepositoryImpl) Save(ctx context.Context, tx *sql.Tx, Lecturer *model.Lecturer) (*model.Lecturer, error) {
	// Menggunakan placeholder $n untuk PostgreSQL
	SQL := "INSERT INTO lecturers (user_id, lecturer_id, department, created_at) VALUES ($1,$2,$3,NOW()) returning id"
	err := tx.QueryRowContext(ctx, SQL, Lecturer.UserID, Lecturer.LecturerID, Lecturer.Department).Scan(&Lecturer.ID)
	if err != nil {
		return nil, err
	}
	return Lecturer, nil
}

func (repo *LecturerRepositoryImpl) FindAll(ctx context.Context, tx *sql.Tx) ([]model.Lecturer, error) {
	//TODO implement me
	panic("implement me")
}

func (repo *LecturerRepositoryImpl) FindById(ctx context.Context, tx *sql.Tx, id string) (*model.Lecturer, error) {
	//TODO implement me
	panic("implement me")
}