package service

import (
	"context"

	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/models"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/repository"
)

// UserServiceInterface defines the interface for user business logic
type UserServiceInterface interface {
	CreateUser(ctx context.Context, userID, username, email string) (*repository.User, error)
	GetUser(ctx context.Context, userID string) (*repository.User, error)
	UpdateUserStatus(ctx context.Context, userID, status string) error
	GetOnlineUsers(ctx context.Context, limit, offset int) ([]*repository.User, error)
	ListUsers(ctx context.Context, limit, offset int) ([]*repository.User, int64, error)
	DeleteUser(ctx context.Context, userID string) error
	UpdateLastSeen(ctx context.Context, userID string) error
}

// MessageServiceInterface defines the interface for message business logic
type MessageServiceInterface interface {
	SendMessage(ctx context.Context, req *models.MessageRequest) (string, error)
	GetMessage(ctx context.Context, messageID string) (*repository.Message, error)
	GetUserMessages(ctx context.Context, userID string, limit, offset int) ([]*repository.Message, int64, error)
	GetRecentMessages(ctx context.Context, limit, offset int) ([]*repository.Message, int64, error)
	GetMessagesByStatus(ctx context.Context, status string, limit, offset int) ([]*repository.Message, int64, error)
	UpdateMessageStatus(ctx context.Context, messageID, status string) error
	MarkAsProcessed(ctx context.Context, messageID string) error
	DeleteMessage(ctx context.Context, messageID string) error
	GetMessageStats(ctx context.Context) (map[string]interface{}, error)
}

// Ensure implementations satisfy interfaces
var _ UserServiceInterface = (*UserService)(nil)
var _ MessageServiceInterface = (*MessageService)(nil)
