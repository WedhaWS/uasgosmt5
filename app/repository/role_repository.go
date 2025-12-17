package repository

import (
	"database/sql"
	"errors"
	"github.com/WedhaWS/uasgosmt5/app/model"
)

type RoleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Mencari Role berdasarkan nama (misal: untuk default role saat register)
func (r *RoleRepository) FindByName(name string) (*model.Role, error) {
	query := `
		SELECT id, name, description, created_at 
		FROM roles 
		WHERE name = $1 
		LIMIT 1`

	var role model.Role
	// Kita harus scan manual karena tidak ada GORM/SQLX
	// Pastikan urutan variabel Scan sama dengan urutan kolom di SELECT
	err := r.db.QueryRow(query, name).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("role not found")
		}
		return nil, err
	}

	return &role, nil
}

// Mengambil Permission yang dimiliki oleh sebuah Role (untuk Middleware RBAC)
func (r *RoleRepository) GetPermissionsByRoleID(roleID string) ([]model.Permission, error) {
	query := `
		SELECT p.id, p.name, p.resource, p.action, p.description
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = $1`

	rows, err := r.db.Query(query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []model.Permission

	for rows.Next() {
		var p model.Permission
		// Scan data per baris
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Resource,
			&p.Action,
			&p.Description,
		)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}

	// Cek error setelah loop (penting di database/sql)
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}