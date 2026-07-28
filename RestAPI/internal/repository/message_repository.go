package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// Message represents a message in the database
// Message mirrors the messages table defined in database/schema.sql.
// 생산자는 두 곳이다 — CommonModule/DBWorker(C++) 가 INSERT 하고 이 리포지토리가 조회·전이한다.
type Message struct {
	ID            int64        `db:"id" json:"id"`
	MessageID     string       `db:"message_id" json:"message_id"`
	UserID        string       `db:"user_id" json:"user_id"`
	SubID         string       `db:"sub_id" json:"sub_id"`
	Command       string       `db:"command" json:"command"`
	PublisherInfo string       `db:"publisher_info" json:"publisher_info"`
	ServerName    string       `db:"server_name" json:"server_name"`
	Content       string       `db:"content" json:"content"`
	IsEncrypted   bool         `db:"is_encrypted" json:"is_encrypted"`
	Status        string       `db:"status" json:"status"`
	CreatedAt     time.Time    `db:"created_at" json:"created_at"`
	ProcessedAt   sql.NullTime `db:"processed_at" json:"processed_at,omitempty"`
}

// MessageRepository defines message data access methods
type MessageRepository interface {
	Create(ctx context.Context, message *Message) error
	GetByMessageID(ctx context.Context, messageID string) (*Message, error)
	GetByID(ctx context.Context, id int64) (*Message, error)
	UpdateStatus(ctx context.Context, messageID string, status string) error
	MarkAsProcessed(ctx context.Context, messageID string) error
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*Message, error)
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*Message, error)
	ListRecent(ctx context.Context, limit, offset int) ([]*Message, error)
	Delete(ctx context.Context, messageID string) error
	Count(ctx context.Context) (int64, error)
	CountByUser(ctx context.Context, userID string) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
}

// messageRepository implements MessageRepository
type messageRepository struct {
	db *sqlx.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *sqlx.DB) MessageRepository {
	return &messageRepository{db: db}
}

// Create creates a new message
func (r *messageRepository) Create(ctx context.Context, message *Message) error {
	query := `
		INSERT INTO messages (message_id, user_id, sub_id, command, publisher_info, server_name, content, is_encrypted, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`

	return r.db.QueryRowxContext(
		ctx, query,
		message.MessageID, message.UserID, message.SubID, message.Command,
		message.PublisherInfo, message.ServerName, message.Content,
		message.IsEncrypted, message.Status,
	).Scan(&message.ID, &message.CreatedAt)
}

// GetByMessageID retrieves a message by message_id
func (r *messageRepository) GetByMessageID(ctx context.Context, messageID string) (*Message, error) {
	query := `
		SELECT id, message_id, user_id, sub_id, command, publisher_info,
		       server_name, content, is_encrypted, status, created_at, processed_at
		FROM messages
		WHERE message_id = $1
	`

	var message Message
	err := r.db.GetContext(ctx, &message, query, messageID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("message not found: %s", messageID)
	}
	return &message, err
}

// GetByID retrieves a message by ID
func (r *messageRepository) GetByID(ctx context.Context, id int64) (*Message, error) {
	query := `
		SELECT id, message_id, user_id, sub_id, command, publisher_info,
		       server_name, content, is_encrypted, status, created_at, processed_at
		FROM messages
		WHERE id = $1
	`

	var message Message
	err := r.db.GetContext(ctx, &message, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("message not found with ID: %d", id)
	}
	return &message, err
}

// UpdateStatus updates message status
func (r *messageRepository) UpdateStatus(ctx context.Context, messageID string, status string) error {
	query := `
		UPDATE messages
		SET status = $1
		WHERE message_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, status, messageID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("message not found: %s", messageID)
	}

	return nil
}

// MarkAsProcessed marks a message as processed
func (r *messageRepository) MarkAsProcessed(ctx context.Context, messageID string) error {
	query := `
		UPDATE messages
		SET status = 'processed', processed_at = CURRENT_TIMESTAMP
		WHERE message_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, messageID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("message not found: %s", messageID)
	}

	return nil
}

// ListByUser retrieves messages for a specific user
func (r *messageRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*Message, error) {
	query := `
		SELECT id, message_id, user_id, sub_id, command, publisher_info,
		       server_name, content, is_encrypted, status, created_at, processed_at
		FROM messages
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var messages []*Message
	err := r.db.SelectContext(ctx, &messages, query, userID, limit, offset)
	return messages, err
}

// ListByStatus retrieves messages by status
func (r *messageRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*Message, error) {
	query := `
		SELECT id, message_id, user_id, sub_id, command, publisher_info,
		       server_name, content, is_encrypted, status, created_at, processed_at
		FROM messages
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var messages []*Message
	err := r.db.SelectContext(ctx, &messages, query, status, limit, offset)
	return messages, err
}

// ListRecent retrieves recent messages
func (r *messageRepository) ListRecent(ctx context.Context, limit, offset int) ([]*Message, error) {
	query := `
		SELECT id, message_id, user_id, sub_id, command, publisher_info,
		       server_name, content, is_encrypted, status, created_at, processed_at
		FROM messages
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var messages []*Message
	err := r.db.SelectContext(ctx, &messages, query, limit, offset)
	return messages, err
}

// Delete deletes a message
func (r *messageRepository) Delete(ctx context.Context, messageID string) error {
	query := `DELETE FROM messages WHERE message_id = $1`

	result, err := r.db.ExecContext(ctx, query, messageID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("message not found: %s", messageID)
	}

	return nil
}

// Count returns the total number of messages
func (r *messageRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM messages`

	var count int64
	err := r.db.GetContext(ctx, &count, query)
	return count, err
}

// CountByUser returns the number of messages for a user
func (r *messageRepository) CountByUser(ctx context.Context, userID string) (int64, error) {
	query := `SELECT COUNT(*) FROM messages WHERE user_id = $1`

	var count int64
	err := r.db.GetContext(ctx, &count, query, userID)
	return count, err
}

// CountByStatus returns the number of messages by status
func (r *messageRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	query := `SELECT COUNT(*) FROM messages WHERE status = $1`

	var count int64
	err := r.db.GetContext(ctx, &count, query, status)
	return count, err
}
