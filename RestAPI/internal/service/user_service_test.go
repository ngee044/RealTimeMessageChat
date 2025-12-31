package service

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/repository"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *repository.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByUserID(ctx context.Context, userID string) (*repository.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*repository.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *repository.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateStatus(ctx context.Context, userID string, status string) error {
	args := m.Called(ctx, userID, status)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastSeen(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*repository.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.User), args.Error(1)
}

func (m *MockUserRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*repository.User, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func TestUserService_CreateUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo, nil)

	ctx := context.Background()
	userID := "test_user"
	username := "Test User"
	email := "test@example.com"

	// Mock expectations
	mockRepo.On("Exists", ctx, userID).Return(false, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*repository.User")).Return(nil)

	user, err := service.CreateUser(ctx, userID, username, email)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.UserID)
	assert.Equal(t, username, user.Username.String)

	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo, nil)

	ctx := context.Background()
	userID := "test_user"

	expectedUser := &repository.User{
		ID:       1,
		UserID:   userID,
		Username: sql.NullString{String: "Test User", Valid: true},
		Status:   "online",
	}

	mockRepo.On("GetByUserID", ctx, userID).Return(expectedUser, nil)

	user, err := service.GetUser(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.UserID)
	assert.Equal(t, "Test User", user.Username.String)

	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateUserStatus(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo, nil)

	ctx := context.Background()
	userID := "test_user"
	status := "online"

	mockRepo.On("UpdateStatus", ctx, userID, status).Return(nil)

	err := service.UpdateUserStatus(ctx, userID, status)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateUserStatus_InvalidStatus(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo, nil)

	ctx := context.Background()
	userID := "test_user"
	invalidStatus := "invalid"

	err := service.UpdateUserStatus(ctx, userID, invalidStatus)

	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "UpdateStatus")
}
