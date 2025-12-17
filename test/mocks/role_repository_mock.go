package mocks

import (
	"github.com/WedhaWS/uasgosmt5/app/model"

	"github.com/stretchr/testify/mock"
)

type MockRoleRepository struct {
	mock.Mock
}

func NewMockRoleRepository() *MockRoleRepository {
	return &MockRoleRepository{}
}

func (m *MockRoleRepository) GetPermissionsByRoleID(roleID string) ([]model.Permission, error) {
	args := m.Called(roleID)
	return args.Get(0).([]model.Permission), args.Error(1)
}

func (m *MockRoleRepository) FindByName(name string) (*model.Role, error) {
	args := m.Called(name)
	return args.Get(0).(*model.Role), args.Error(1)
}
