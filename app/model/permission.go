package model

// Tabel permissions
type Permission struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Resource    string `json:"resource" db:"resource"`
	Action      string `json:"action" db:"action"`
	Description string `json:"description" db:"description"`
}

// Tabel role_permissions
type RolePermission struct {
	RoleID       string      `json:"roleId" db:"role_id"`
	PermissionID string      `json:"permissionId" db:"permission_id"`
	
	// Relasi (Tidak ada di kolom database, diisi manual via JOIN)
	Role         *Role       `json:"role,omitempty" db:"-"`
	Permission   *Permission `json:"permission,omitempty" db:"-"`
}

// Override nama tabel agar tidak di-pluralize otomatis (menjadi role_permissions)
func (RolePermission) TableName() string {
	return "role_permissions"
}