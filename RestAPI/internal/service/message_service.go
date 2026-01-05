package service

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/models"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/repository"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/services"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/cache"
	apperrors "github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/errors"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/logger"
)

// MessageService handles message business logic
type MessageService struct {
	messageRepo repository.MessageRepository
	userRepo    repository.UserRepository
	rabbitMQ    *services.RabbitMQService
	redis       *services.RedisService
}

// NewMessageService creates a new message service
func NewMessageService(
	messageRepo repository.MessageRepository,
	userRepo repository.UserRepository,
	rabbitMQ *services.RabbitMQService,
	redis *services.RedisService,
) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		userRepo:    userRepo,
		rabbitMQ:    rabbitMQ,
		redis:       redis,
	}
}

// SendMessage sends a message through RabbitMQ and stores it in the database
func (s *MessageService) SendMessage(ctx context.Context, req *models.MessageRequest) (string, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return "", apperrors.Wrap(err, apperrors.ErrCodeValidation, "Validation failed", 400)
	}

	// Generate message ID
	messageID := uuid.New().String()

	// Convert to queue message
	queueMsg := req.ToQueueMessage(messageID)

	// Serialize to JSON
	msgBytes, err := queueMsg.ToJSON()
	if err != nil {
		logger.Errorf("Failed to serialize message: %v", err)
		return "", apperrors.Wrap(err, apperrors.ErrCodeInternal, "Failed to process message", 500)
	}

	// Store in database
	dbMessage := &repository.Message{
		MessageID: messageID,
		UserID:    req.UserID,
		Command:   req.Command,
		SubID:     sql.NullString{String: req.SubID, Valid: req.SubID != ""},
		Content:   req.Content,
		Priority:  req.Priority,
		Status:    "pending",
	}

	// Set metadata if provided
	if req.Metadata != nil {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err == nil {
			dbMessage.Metadata = metadataBytes
		}
	}

	if err := s.messageRepo.Create(ctx, dbMessage); err != nil {
		logger.Errorf("Failed to store message in database: %v", err)
		return "", apperrors.Wrap(err, apperrors.ErrCodeDatabaseError, "Failed to store message", 500)
	}

	// Publish to RabbitMQ
	if err := s.rabbitMQ.Publish(ctx, msgBytes); err != nil {
		logger.Errorf("Failed to publish message to RabbitMQ: %v", err)
		// Update message status to failed
		s.messageRepo.UpdateStatus(ctx, messageID, "failed")
		return "", apperrors.Wrap(err, apperrors.ErrCodeQueueError, "Failed to send message", 500)
	}

	// Update message status to sent
	s.messageRepo.UpdateStatus(ctx, messageID, "sent")

	// Update user's last seen
	if s.userRepo != nil {
		s.userRepo.UpdateLastSeen(ctx, req.UserID)
	}

	logger.Infof("Message sent successfully: %s (user: %s, command: %s)", messageID, req.UserID, req.Command)
	return messageID, nil
}

// GetMessage retrieves a message by ID
func (s *MessageService) GetMessage(ctx context.Context, messageID string) (*repository.Message, error) {
	message, err := s.messageRepo.GetByMessageID(ctx, messageID)
	if err != nil {
		logger.Warnf("Message not found: %s", messageID)
		return nil, apperrors.Wrap(err, apperrors.ErrCodeNotFound, "Message not found", 404)
	}

	return message, nil
}

// GetUserMessages retrieves messages for a specific user
func (s *MessageService) GetUserMessages(ctx context.Context, userID string, limit, offset int) ([]*repository.Message, int64, error) {
	messages, err := s.messageRepo.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		logger.Errorf("Failed to get user messages: %v", err)
		return nil, 0, apperrors.Wrap(err, apperrors.ErrCodeDatabaseError, "Failed to get messages", 500)
	}

	total, err := s.messageRepo.CountByUser(ctx, userID)
	if err != nil {
		logger.Errorf("Failed to count user messages: %v", err)
		return nil, 0, apperrors.Wrap(err, apperrors.ErrCodeDatabaseError, "Failed to count messages", 500)
	}

	return messages, total, nil
}

// GetRecentMessages retrieves recent messages
func (s *MessageService) GetRecentMessages(ctx context.Context, limit, offset int) ([]*repository.Message, int64, error) {
	messages, err := s.messageRepo.ListRecent(ctx, limit, offset)
	if err != nil {
		logger.Errorf("Failed to get recent messages: %v", err)
		return nil, 0, apperrors.Wrap(err, apperrors.ErrCodeDatabaseError, "Failed to get messages", 500)
	}

	total, err := s.messageRepo.Count(ctx)
	if err != nil {
		logger.Errorf("Failed to count messages: %v", err)
		return nil, 0, apperrors.Wrap(err, apperrors.ErrCodeDatabaseError, "Failed to count messages", 500)
	}

	return messages, total, nil
}

// GetMessagesByStatus retrieves messages by status
func (s *MessageService) GetMessagesByStatus(ctx context.Context, status string, limit, offset int) ([]*repository.Message, int64, error) {
	messages, err := s.messageRepo.ListByStatus(ctx, status, limit, offset)
	if err != nil {
		logger.Errorf("Failed to get messages by status: %v", err)
		return nil, 0, apperrors.Wrap(err, apperrors.ErrCodeDatabaseError, "Failed to get messages", 500)
	}

	total, err := s.messageRepo.CountByStatus(ctx, status)
	if err != nil {
		logger.Errorf("Failed to count messages by status: %v", err)
		return nil, 0, apperrors.Wrap(err, apperrors.ErrCodeDatabaseError, "Failed to count messages", 500)
	}

	return messages, total, nil
}

// UpdateMessageStatus updates message status
func (s *MessageService) UpdateMessageStatus(ctx context.Context, messageID, status string) error {
	if err := s.messageRepo.UpdateStatus(ctx, messageID, status); err != nil {
		logger.Errorf("Failed to update message status: %v", err)
		return apperrors.Wrap(err, apperrors.ErrCodeDatabaseError, "Failed to update message status", 500)
	}

	// Cache the status update in Redis
	if s.redis != nil {
		s.redis.Set(ctx, cache.MessageStatusKey(messageID), status, cache.TTLMessageStatus)
	}

	return nil
}

// MarkAsProcessed marks a message as processed
func (s *MessageService) MarkAsProcessed(ctx context.Context, messageID string) error {
	if err := s.messageRepo.MarkAsProcessed(ctx, messageID); err != nil {
		logger.Errorf("Failed to mark message as processed: %v", err)
		return apperrors.Wrap(err, apperrors.ErrCodeDatabaseError, "Failed to mark message as processed", 500)
	}

	logger.Infof("Message marked as processed: %s", messageID)
	return nil
}

// DeleteMessage deletes a message
func (s *MessageService) DeleteMessage(ctx context.Context, messageID string) error {
	if err := s.messageRepo.Delete(ctx, messageID); err != nil {
		logger.Errorf("Failed to delete message: %v", err)
		return apperrors.Wrap(err, apperrors.ErrCodeDatabaseError, "Failed to delete message", 500)
	}

	// Remove from cache
	if s.redis != nil {
		s.redis.Delete(ctx, cache.MessageStatusKey(messageID))
	}

	logger.Infof("Message deleted: %s", messageID)
	return nil
}

// GetMessageStats retrieves message statistics
func (s *MessageService) GetMessageStats(ctx context.Context) (map[string]interface{}, error) {
	total, err := s.messageRepo.Count(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrCodeDatabaseError, "Failed to get stats", 500)
	}

	pending, _ := s.messageRepo.CountByStatus(ctx, "pending")
	sent, _ := s.messageRepo.CountByStatus(ctx, "sent")
	processed, _ := s.messageRepo.CountByStatus(ctx, "processed")
	failed, _ := s.messageRepo.CountByStatus(ctx, "failed")

	stats := map[string]interface{}{
		"total":     total,
		"pending":   pending,
		"sent":      sent,
		"processed": processed,
		"failed":    failed,
	}

	return stats, nil
}
