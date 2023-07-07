package mocks

import (
	user "Makhkets/internal/user/repository"
	"context"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateUser(ctx context.Context, user *user.UserDTO) (*user.UserDTO, error) {
	//TODO implement me
	panic("implement me")
}
func (m *MockRepository) FindOne(ctx context.Context, id string) (*user.User, error) {
	//TODO implement me
	panic("implement me")
}
func (m *MockRepository) Delete(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}
func (m *MockRepository) FindLoginUser(ctx context.Context, username, password string) (*user.User, error) {
	//TODO implement me
	panic("implement me")
}
func (m *MockRepository) UpdateUsername(ctx context.Context, id, username string) error {
	//TODO implement me
	panic("implement me")
}
func (m *MockRepository) UpdatePassword(ctx context.Context, id, oldPassword, newPassword string) error {
	//TODO implement me
	panic("implement me")
}
func (m *MockRepository) GetUser(ctx context.Context, id string) (*user.User, error) {
	//TODO implement me
	panic("implement me")
}
func (m *MockRepository) ChangeStatus(ctx context.Context, id, status string) error {
	//TODO implement me
	panic("implement me")
}
func (m *MockRepository) ChangePermission(ctx context.Context, id string, permission bool) error {
	//TODO implement me
	panic("implement me")
}
func (m *MockRepository) GetRefreshSession(ctx context.Context, fingerprint string) (*user.RefreshSession, error) {
	//TODO implement me
	panic("implement me")
}
func (m *MockRepository) DeleteRefreshSession(ctx context.Context, key, id string) error {
	//TODO implement me
	panic("implement me")
}
func (m *MockRepository) GetRefreshSessionsByUserId(ctx context.Context, userId string) ([]*user.RefreshSession, error) {
	//TODO implement me
	panic("implement me")
}

func NewMockRepository() user.Repository {
	return new(MockRepository)
}

func (m *MockRepository) SaveRefreshSession(ctx context.Context, rs *user.RefreshSession) error {
	args := m.Called(ctx, rs)
	return args.Error(0)
}
