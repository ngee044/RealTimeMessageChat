package service

import (
	"context"

	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/repository"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/services"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/cache"
	apperrors "github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/errors"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/logger"
)

// MessageService handles message business logic
type MessageService struct {
	messageRepo repository.MessageRepository
	redis       *services.RedisService
}

// NewMessageService creates a new message service
func NewMessageService(
	messageRepo repository.MessageRepository,
	redis *services.RedisService,
) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		redis:       redis,
	}
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

	// 캐시는 best-effort 지만, 실패를 삼키면 갱신 전 값이 TTL 동안 계속 조회되므로 남긴다.
	if s.redis != nil {
		if err := s.redis.Set(ctx, cache.MessageStatusKey(messageID), status, cache.TTLMessageStatus); err != nil {
			logger.Warnf("Failed to cache message status (%s): %v", messageID, err)
		}
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

	// 삭제 후 캐시가 남으면 없는 메시지의 상태가 계속 조회되므로 실패를 반드시 기록한다.
	if s.redis != nil {
		if err := s.redis.Delete(ctx, cache.MessageStatusKey(messageID)); err != nil {
			logger.Warnf("Failed to evict message status cache (%s): %v", messageID, err)
		}
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
