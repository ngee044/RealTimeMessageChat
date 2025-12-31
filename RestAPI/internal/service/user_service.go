package service

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/repository"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/services"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/cache"
	apperrors "github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/errors"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/logger"
)

// UserService handles user business logic
type UserService struct {
	userRepo repository.UserRepository
	redis    *services.RedisService
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, redis *services.RedisService) *UserService {
	return &UserService{
		userRepo: userRepo,
		redis:    redis,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, userID, username, email string) (*repository.User, error) {
	// Check if user already exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		logger.Errorf("Failed to check user existence: %v", err)
		return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabaseError, "Failed to check user", 500)
	}

	if exists {
		return nil, apperrors.New(apperrors.ErrCodeDuplicateKey, "User already exists", 409)
	}

	user := &repository.User{
		UserID:   userID,
		Username: sql.NullString{String: username, Valid: username != ""},
		Email:    sql.NullString{String: email, Valid: email != ""},
		Status:   "offline",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		logger.Errorf("Failed to create user: %v", err)
		return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabaseError, "Failed to create user", 500)
	}

	// Cache user in Redis
	if s.redis != nil {
		s.cacheUser(ctx, user)
	}

	logger.Infof("User created successfully: %s", userID)
	return user, nil
}

// GetUser retrieves a user by user ID
func (s *UserService) GetUser(ctx context.Context, userID string) (*repository.User, error) {
	// Try to get from cache first
	if s.redis != nil {
		if user, err := s.getUserFromCache(ctx, userID); err == nil {
			return user, nil
		}
	}

	// Get from database
	user, err := s.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Warnf("User not found: %s", userID)
		return nil, apperrors.Wrap(err, apperrors.ErrCodeNotFound, "User not found", 404)
	}

	// Cache the user
	if s.redis != nil {
		s.cacheUser(ctx, user)
	}

	return user, nil
}

// UpdateUserStatus updates user status (online/offline/away)
func (s *UserService) UpdateUserStatus(ctx context.Context, userID, status string) error {
	// Validate status
	validStatuses := map[string]bool{"online": true, "offline": true, "away": true}
	if !validStatuses[status] {
		return apperrors.New(apperrors.ErrCodeValidation, "Invalid status", 400)
	}

	// Update in database
	if err := s.userRepo.UpdateStatus(ctx, userID, status); err != nil {
		logger.Errorf("Failed to update user status: %v", err)
		return apperrors.Wrap(err, apperrors.ErrCodeDatabaseError, "Failed to update status", 500)
	}

	// Update in cache
	if s.redis != nil {
		cacheKey := cache.UserStatusKey(userID)
		s.redis.Set(ctx, cacheKey, status, cache.TTLUserStatus)
	}

	logger.Infof("User status updated: %s -> %s", userID, status)
	return nil
}

// GetOnlineUsers retrieves all online users
func (s *UserService) GetOnlineUsers(ctx context.Context, limit, offset int) ([]*repository.User, error) {
	users, err := s.userRepo.ListByStatus(ctx, "online", limit, offset)
	if err != nil {
		logger.Errorf("Failed to get online users: %v", err)
		return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabaseError, "Failed to get online users", 500)
	}

	return users, nil
}

// ListUsers retrieves paginated list of users
func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]*repository.User, int64, error) {
	users, err := s.userRepo.List(ctx, limit, offset)
	if err != nil {
		logger.Errorf("Failed to list users: %v", err)
		return nil, 0, apperrors.Wrap(err, apperrors.ErrCodeDatabaseError, "Failed to list users", 500)
	}

	total, err := s.userRepo.Count(ctx)
	if err != nil {
		logger.Errorf("Failed to count users: %v", err)
		return nil, 0, apperrors.Wrap(err, apperrors.ErrCodeDatabaseError, "Failed to count users", 500)
	}

	return users, total, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		logger.Errorf("Failed to delete user: %v", err)
		return apperrors.Wrap(err, apperrors.ErrCodeDatabaseError, "Failed to delete user", 500)
	}

	// Remove from cache
	if s.redis != nil {
		s.redis.Delete(ctx, cache.UserKey(userID), cache.UserStatusKey(userID))
	}

	logger.Infof("User deleted: %s", userID)
	return nil
}

// UpdateLastSeen updates user's last seen timestamp
func (s *UserService) UpdateLastSeen(ctx context.Context, userID string) error {
	if err := s.userRepo.UpdateLastSeen(ctx, userID); err != nil {
		// Don't fail the request if this fails, just log it
		logger.Warnf("Failed to update last seen for user %s: %v", userID, err)
	}
	return nil
}

// cacheUser caches a user in Redis
func (s *UserService) cacheUser(ctx context.Context, user *repository.User) {
	userData, err := json.Marshal(user)
	if err != nil {
		logger.Warnf("Failed to marshal user for cache: %v", err)
		return
	}

	if err := s.redis.Set(ctx, cache.UserKey(user.UserID), userData, cache.TTLUserData); err != nil {
		logger.Warnf("Failed to cache user: %v", err)
	}
}

// getUserFromCache retrieves a user from Redis cache
func (s *UserService) getUserFromCache(ctx context.Context, userID string) (*repository.User, error) {
	data, err := s.redis.Get(ctx, cache.UserKey(userID))
	if err != nil {
		return nil, err
	}

	var user repository.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}

	return &user, nil
}
