package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/WedhaWS/uasgosmt5/app/model"

	"github.com/sirupsen/logrus"
)

type UserRepository interface {
	Save(ctx context.Context, User *model.User) (*model.User, error)
	Update(ctx context.Context, User model.User) (*model.User, error)
	Delete(ctx context.Context, UserId int) error
	FindById(ctx context.Context, UserId int64) (*model.User, error)
	FindAll(ctx context.Context) (*[]model.User, error)
	FindByUsername(ctx context.Context, Username string) (*model.User, error)
}

type UserRepositoryImpl struct {
	DB  *sql.DB
	Log *logrus.Logger
}

func NewUserRepository(DB *sql.DB, Log *logrus.Logger) UserRepository {
	return &UserRepositoryImpl{
		DB:  DB,
		Log: Log,
	}
}

func (repo *UserRepositoryImpl) Save(ctx context.Context, User *model.User) (*model.User, error) {
	SQL := "INSERT INTO users (username, email, password_hash, full_name, role_id) VALUES ($1,$2,$3,$4,$5) returning id;"
	err := repo.DB.QueryRowContext(ctx, SQL, User.Username, User.Email, User.PasswordHash, User.FullName, User.RoleId).Scan(&User.ID)
	if err != nil {
		repo.Log.Fatalf("Error inserting user into database: %v", err)
		return nil, err
	}
	return User, nil
}

func (repo *UserRepositoryImpl) Update(ctx context.Context, User model.User) (*model.User, error) {
	SQL := "UPDATE users SET username = $1,email = $2 ,full_name = $3 ,role_id = $4 FROM users WHERE username = $5;"
	res, err := repo.DB.ExecContext(ctx, SQL,
		User.Username,
		User.Email,
		User.FullName,
		User.RoleId,
		User.ID,
	)

	if err != nil {
		repo.Log.Fatalf("Error updating user into database: %v", err)
		return nil, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		repo.Log.Fatalf("Error updating user into database: %v", err)
		return nil, err
	}

	if rows == 0 {
		return &model.User{}, nil
	} else {
		return &User, nil
	}
}

func (repo *UserRepositoryImpl) Delete(ctx context.Context, UserId int) error {
	SQL := "DELETE FROM users WHERE id = $1 ;"
	_, err := repo.DB.ExecContext(ctx, SQL, UserId)
	if err != nil {
		repo.Log.Fatalf("Error deleting user into database: %v", err)
		return err
	}
	return nil
}

func (repo *UserRepositoryImpl) FindById(ctx context.Context, UserId int64) (*model.User, error) {
	SQL := `SELECT u.id,u.email,u.username,u.full_name,r.name,
			COALESCE(
					JSON_AGG(
						JSON_BUILD_OBJECT('resource', p.resource, 'action', p.action)
					) FILTER (WHERE p.id IS NOT NULL), 
					'[]'
				) as permissions
			FROM users u 
    		INNER JOIN roles r ON u.role_id = r.id
			LEFT JOIN role_permissions rp ON u.role_id = rp.role_id
			LEFT JOIN permissions p ON rp.permission_id = p.id
			WHERE u.id = $1
			GROUP BY u.id,u.email,u.username,u.full_name,r.name;`

	var user model.User
	var permissions []byte
	err := repo.DB.QueryRowContext(ctx, SQL, UserId).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.FullName,
		&user.RoleName,
		&permissions)

	if err != nil {
		repo.Log.Fatalf("Error finding user by id into database: %v", err)
		return nil, err
	}

	return &user, nil
}

func (repo *UserRepositoryImpl) FindAll(ctx context.Context) (*[]model.User, error) {
	SQL := `SELECT u.id,u.email,u.username,u.full_name,r.name,
			COALESCE(
					JSON_AGG(
						JSON_BUILD_OBJECT('resource', p.resource, 'action', p.action)
					) FILTER (WHERE p.id IS NOT NULL), 
					'[]'
				) as permissions
			FROM users u 
    		INNER JOIN roles r ON u.role_id = r.id
			LEFT JOIN role_permissions rp ON u.role_id = rp.role_id
			LEFT JOIN permissions p ON rp.permission_id = p.id
			GROUP BY u.id,u.email,u.username,u.full_name,r.name;`
	rows, err := repo.DB.QueryContext(ctx, SQL)
	if err != nil {
		repo.Log.Fatalf("Error finding all users from database: %v", err)
	}
	defer rows.Close()
	var users []model.User
	for rows.Next() {
		var user model.User
		var permissions []byte
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.FullName,
			&user.RoleName,
			&permissions)
		if err != nil {
			repo.Log.Fatalf("Error finding all users from database: %v", err)
			return nil, err
		}
		if err := json.Unmarshal(permissions, &user.Permissions); err != nil {
			return nil, fmt.Errorf("unmarshal permissions: %w", err)
		}
		users = append(users, user)
	}

	return &users, nil
}

func (repo *UserRepositoryImpl) FindByUsername(ctx context.Context, Username string) (*model.User, error) {
	SQL := `SELECT u.id,u.username,u.full_name,u.password_hash,r.name,
			COALESCE(
        			TO_JSON(JSON_AGG(p.resource || ':' || p.action)),
       			 '[]'
    		) AS permissions
			FROM users u 
    		INNER JOIN roles r ON u.role_id = r.id
			LEFT JOIN role_permissions rp ON u.role_id = rp.role_id
			LEFT JOIN permissions p ON rp.permission_id = p.id
			WHERE u.username = $1 
			GROUP BY u.id,u.username,u.full_name,u.password_hash,r.name;`

	var user model.User
	var permStr string
	err := repo.DB.QueryRowContext(ctx, SQL, Username).Scan(
		&user.ID,
		&user.Username,
		&user.FullName,
		&user.PasswordHash,
		&user.RoleName,
		&permStr,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(permStr), &user.Permissions); err != nil {
		return nil, fmt.Errorf("unmarshal permissions: %w", err)
	}
	return &user, nil
}