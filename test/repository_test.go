package test

import (
	"database/sql"
	"testing"
	"time"
	"github.com/WedhaWS/uasgosmt5/app/model"
	"github.com/WedhaWS/uasgosmt5/app/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_FindByEmail(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Create repository
	userRepo := repository.NewUserRepository(db)

	t.Run("User found successfully", func(t *testing.T) {
		email := "test@example.com"
		now := time.Now()

		// Set up mock expectations
		rows := sqlmock.NewRows([]string{
			"id", "username", "email", "password_hash", "full_name", "role_id", "is_active", "created_at", "updated_at",
			"role_id", "role_name", "role_description",
		}).AddRow(
			"user-123", "testuser", email, "hashed_password", "Test User", "role-123", true, now, now,
			"role-123", "Admin", "Administrator",
		)

		mock.ExpectQuery(`SELECT .+ FROM users u JOIN roles r ON u.role_id = r.id WHERE u.email = \$1`).
			WithArgs(email).
			WillReturnRows(rows)

		// Execute
		user, err := userRepo.FindByEmail(email)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "user-123", user.ID)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "Test User", user.FullName)
		assert.True(t, user.IsActive)
		assert.NotNil(t, user.Role)
		assert.Equal(t, "Admin", user.Role.Name)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("User not found", func(t *testing.T) {
		email := "nonexistent@example.com"

		// Set up mock to return no rows
		mock.ExpectQuery(`SELECT .+ FROM users u JOIN roles r ON u.role_id = r.id WHERE u.email = \$1`).
			WithArgs(email).
			WillReturnError(sql.ErrNoRows)

		// Execute
		user, err := userRepo.FindByEmail(email)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database error", func(t *testing.T) {
		email := "error@example.com"

		// Set up mock to return database error
		mock.ExpectQuery(`SELECT .+ FROM users u JOIN roles r ON u.role_id = r.id WHERE u.email = \$1`).
			WithArgs(email).
			WillReturnError(sql.ErrConnDone)

		// Execute
		user, err := userRepo.FindByEmail(email)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, sql.ErrConnDone, err)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_Create(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Create repository
	userRepo := repository.NewUserRepository(db)

	t.Run("User created successfully", func(t *testing.T) {
		now := time.Now()
		user := &model.User{
			Username:     "newuser",
			Email:        "newuser@example.com",
			PasswordHash: "hashed_password",
			FullName:     "New User",
			RoleID:       "role-123",
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		// Set up mock expectations
		rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("user-456", now, now)

		mock.ExpectQuery(`INSERT INTO users \(username, email, password_hash, full_name, role_id, is_active, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8\) RETURNING id, created_at, updated_at`).
			WithArgs(user.Username, user.Email, user.PasswordHash, user.FullName, user.RoleID, user.IsActive, user.CreatedAt, user.UpdatedAt).
			WillReturnRows(rows)

		// Execute
		err := userRepo.Create(user)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, "user-456", user.ID)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database error during creation", func(t *testing.T) {
		user := &model.User{
			Username:     "erroruser",
			Email:        "error@example.com",
			PasswordHash: "hashed_password",
			FullName:     "Error User",
			RoleID:       "role-123",
			IsActive:     true,
		}

		// Set up mock to return error
		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)

		// Execute
		err := userRepo.Create(user)

		// Assertions
		assert.Error(t, err)
		assert.Equal(t, sql.ErrConnDone, err)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_FindStudentByUserID(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Create repository
	userRepo := repository.NewUserRepository(db)

	t.Run("Student found with advisor", func(t *testing.T) {
		userID := "user-123"
		now := time.Now()

		// Set up mock expectations
		rows := sqlmock.NewRows([]string{
			"student_id", "user_id", "student_number", "program_study", "academic_year", "advisor_id", "created_at",
			"user_id", "username", "full_name", "email",
			"lecturer_id", "lecturer_number", "department",
			"advisor_user_id", "advisor_name",
		}).AddRow(
			"student-123", userID, "123456789", "Teknik Informatika", "2023/2024", "lecturer-123", now,
			userID, "student1", "John Doe", "john@example.com",
			"lecturer-123", "198001012005011001", "Informatika",
			"user-lecturer", "Dr. Smith",
		)

		mock.ExpectQuery(`SELECT .+ FROM students s JOIN users u ON s.user_id = u.id LEFT JOIN lecturers l ON s.advisor_id = l.id LEFT JOIN users au ON l.user_id = au.id WHERE s.user_id = \$1`).
			WithArgs(userID).
			WillReturnRows(rows)

		// Execute
		student, err := userRepo.FindStudentByUserID(userID)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, student)
		assert.Equal(t, "student-123", student.ID)
		assert.Equal(t, userID, student.UserID)
		assert.Equal(t, "123456789", student.StudentID)
		assert.Equal(t, "Teknik Informatika", student.ProgramStudy)
		assert.NotNil(t, student.User)
		assert.Equal(t, "John Doe", student.User.FullName)
		assert.NotNil(t, student.Advisor)
		assert.Equal(t, "lecturer-123", student.Advisor.ID)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Student not found", func(t *testing.T) {
		userID := "nonexistent-user"

		// Set up mock to return no rows
		mock.ExpectQuery(`SELECT .+ FROM students s`).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		// Execute
		student, err := userRepo.FindStudentByUserID(userID)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, student)
		assert.Contains(t, err.Error(), "student profile not found")

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_Update(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Create repository
	userRepo := repository.NewUserRepository(db)

	t.Run("User updated successfully", func(t *testing.T) {
		user := &model.User{
			ID:       "user-123",
			Username: "updateduser",
			Email:    "updated@example.com",
			FullName: "Updated User",
			IsActive: true,
		}

		// Set up mock expectations
		mock.ExpectExec(`UPDATE users SET full_name = \$1, username = \$2, email = \$3, is_active = \$4, updated_at = \$5 WHERE id = \$6`).
			WithArgs(user.FullName, user.Username, user.Email, user.IsActive, sqlmock.AnyArg(), user.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Execute
		err := userRepo.Update(user)

		// Assertions
		assert.NoError(t, err)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database error during update", func(t *testing.T) {
		user := &model.User{
			ID:       "user-123",
			Username: "erroruser",
			Email:    "error@example.com",
			FullName: "Error User",
			IsActive: true,
		}

		// Set up mock to return error
		mock.ExpectExec(`UPDATE users`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)

		// Execute
		err := userRepo.Update(user)

		// Assertions
		assert.Error(t, err)
		assert.Equal(t, sql.ErrConnDone, err)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_Delete(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Create repository
	userRepo := repository.NewUserRepository(db)

	t.Run("User deleted successfully", func(t *testing.T) {
		userID := "user-123"

		// Set up mock expectations for cascade deletes
		mock.ExpectExec(`DELETE FROM students WHERE user_id = \$1`).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(`DELETE FROM lecturers WHERE user_id = \$1`).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Execute
		err := userRepo.Delete(userID)

		// Assertions
		assert.NoError(t, err)

		// Verify all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
