package interfaces

import (
	"github.com/WedhaWS/uasgosmt5/app/model"
)

// RoleRepositoryInterface defines the interface for role repository
type RoleRepositoryInterface interface {
	GetPermissionsByRoleID(roleID string) ([]model.Permission, error)
	FindByName(name string) (*model.Role, error)
}
