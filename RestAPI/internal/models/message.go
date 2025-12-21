package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// MessageRequest represents the incoming message request from clients
type MessageRequest struct {
	UserID    string                 `json:"user_id" binding:"required"`
	Command   string                 `json:"command" binding:"required"`
	SubID     string                 `json:"sub_id,omitempty"`
	Content   string                 `json:"content" binding:"required"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Priority  int                    `json:"priority,omitempty"` // 1=high, 2=normal, 3=low
	Timestamp int64                  `json:"timestamp,omitempty"`
}

// MessageResponse represents the API response
type MessageResponse struct {
	Success   bool        `json:"success"`
	MessageID string      `json:"message_id,omitempty"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error"`
	Code      string `json:"code,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// QueueMessage represents the message structure sent to RabbitMQ
type QueueMessage struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Command   string                 `json:"command"`
	SubID     string                 `json:"sub_id,omitempty"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Priority  int                    `json:"priority"`
	Timestamp int64                  `json:"timestamp"`
	CreatedAt int64                  `json:"created_at"`
}

// Validate validates the message request
func (m *MessageRequest) Validate() error {
	if m.UserID == "" {
		return fmt.Errorf("user_id is required")
	}

	if m.Command == "" {
		return fmt.Errorf("command is required")
	}

	if m.Content == "" {
		return fmt.Errorf("content is required")
	}

	// Validate priority range
	if m.Priority != 0 && (m.Priority < 1 || m.Priority > 3) {
		return fmt.Errorf("priority must be between 1 and 3 when provided")
	}

	// Set default priority if not provided
	if m.Priority == 0 {
		m.Priority = 2 // Default to normal priority
	}

	return nil
}

// ToQueueMessage converts MessageRequest to QueueMessage
func (m *MessageRequest) ToQueueMessage(messageID string) *QueueMessage {
	now := time.Now().Unix()

	timestamp := m.Timestamp
	if timestamp == 0 {
		timestamp = now
	}

	return &QueueMessage{
		ID:        messageID,
		UserID:    m.UserID,
		Command:   m.Command,
		SubID:     m.SubID,
		Content:   m.Content,
		Metadata:  m.Metadata,
		Priority:  m.Priority,
		Timestamp: timestamp,
		CreatedAt: now,
	}
}

// ToJSON converts QueueMessage to JSON bytes
func (q *QueueMessage) ToJSON() ([]byte, error) {
	return json.Marshal(q)
}

// NewMessageResponse creates a new success response
func NewMessageResponse(messageID string, message string, data interface{}) *MessageResponse {
	return &MessageResponse{
		Success:   true,
		MessageID: messageID,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(error string, code string) *ErrorResponse {
	return &ErrorResponse{
		Success:   false,
		Error:     error,
		Code:      code,
		Timestamp: time.Now().Unix(),
	}
}
