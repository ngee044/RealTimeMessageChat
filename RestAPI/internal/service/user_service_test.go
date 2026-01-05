package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	ctx := context.Background()

	t.Run("successful user creation", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, nil)

		userID := "test_user_123"
		username := "testuser"
		email := "test@example.com"

		mockRepo.On("Exists", ctx, userID).Return(false, nil)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*repository.User")).Return(nil)

		user, err := service.CreateUser(ctx, userID, username, email)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.UserID)
		assert.Equal(t, username, user.Username.String)
		assert.Equal(t, email, user.Email.String)
		assert.Equal(t, "offline", user.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user already exists", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, nil)

		userID := "existing_user"
		username := "existinguser"
		email := "existing@example.com"

		mockRepo.On("Exists", ctx, userID).Return(true, nil)

		user, err := service.CreateUser(ctx, userID, username, email)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "already exists")
		mockRepo.AssertExpectations(t)
	})

	t.Run("database error on exists check", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, nil)

		userID := "test_user"
		dbError := errors.New("database connection error")

		mockRepo.On("Exists", ctx, userID).Return(false, dbError)

		user, err := service.CreateUser(ctx, userID, "user", "email@test.com")

		assert.Error(t, err)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("database error on create", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, nil)

		userID := "test_user"
		dbError := errors.New("database insert error")

		mockRepo.On("Exists", ctx, userID).Return(false, nil)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*repository.User")).Return(dbError)

		user, err := service.CreateUser(ctx, userID, "user", "email@test.com")

		assert.Error(t, err)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUser(t *testing.T) {
	ctx := context.Background()

	t.Run("user found in database", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, nil)

		userID := "test_user_123"
		expectedUser := &repository.User{
			ID:       1,
			UserID:   userID,
			Username: sql.NullString{String: "testuser", Valid: true},
			Email:    sql.NullString{String: "test@example.com", Valid: true},
			Status:   "online",
		}

		mockRepo.On("GetByUserID", ctx, userID).Return(expectedUser, nil)

		user, err := service.GetUser(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.UserID)
		assert.Equal(t, "testuser", user.Username.String)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, nil)

		userID := "nonexistent_user"
		mockRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("user not found"))

		user, err := service.GetUser(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateUserStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("successful status update", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, nil)

		userID := "test_user_123"
		status := "online"

		mockRepo.On("UpdateStatus", ctx, userID, status).Return(nil)

		err := service.UpdateUserStatus(ctx, userID, status)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid status", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, nil)

		userID := "test_user_123"
		invalidStatus := "invalid_status"

		err := service.UpdateUserStatus(ctx, userID, invalidStatus)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid status")
		mockRepo.AssertNotCalled(t, "UpdateStatus")
	})

	t.Run("valid statuses", func(t *testing.T) {
		validStatuses := []string{"online", "offline", "away"}

		for _, status := range validStatuses {
			mockRepo := new(MockUserRepository)
			service := NewUserService(mockRepo, nil)

			userID := "test_user"
			mockRepo.On("UpdateStatus", ctx, userID, status).Return(nil)

			err := service.UpdateUserStatus(ctx, userID, status)

			assert.NoError(t, err)
			mockRepo.AssertExpectations(t)
		}
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, nil)

		userID := "test_user_123"
		status := "online"
		dbError := errors.New("database update error")

		mockRepo.On("UpdateStatus", ctx, userID, status).Return(dbError)

		err := service.UpdateUserStatus(ctx, userID, status)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, nil)

		userID := "test_user_123"

		mockRepo.On("Delete", ctx, userID).Return(nil)

		err := service.DeleteUser(ctx, userID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo, nil)

		userID := "test_user_123"
		dbError := errors.New("database delete error")

		mockRepo.On("Delete", ctx, userID).Return(dbError)

		err := service.DeleteUser(ctx, userID)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}
