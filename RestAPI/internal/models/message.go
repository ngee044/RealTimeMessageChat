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

// QueueMessagePayload is the nested "message" object of the canonical queue schema.
// MainServerConsumer 는 message 가 object 일 것을 요구하고 MainServer·DBWorker 는
// message.content 를 읽는다. docker/publish-message.sh 가 만드는 형식과 동일하다.
type QueueMessagePayload struct {
	Command string `json:"command"`
	Content string `json:"content"`
	// 소비측(Consumer/MainServer/DBWorker) 어디도 읽지 않는 발행측 진단용 필드다.
	Timestamp string `json:"timestamp,omitempty"`
}

// QueuePublisherInformation carries producer-side metadata.
// DBWorker 가 publisher_info TEXT 로 그대로 영속화하므로 message_id 추적성이 유지된다.
type QueuePublisherInformation struct {
	MessageID string                 `json:"message_id"`
	Source    string                 `json:"source"`
	Priority  int                    `json:"priority"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt int64                  `json:"created_at"`
}

// QueueMessage represents the message structure sent to RabbitMQ.
// id 는 사용자 식별자다 — database/schema.sql 과 DBWorker 가 그렇게 정의한다.
// sub_id 는 Consumer 가 필수 문자열로 요구하므로 omitempty 를 쓰지 않는다.
type QueueMessage struct {
	ID                   string                    `json:"id"`
	SubID                string                    `json:"sub_id"`
	PublisherInformation QueuePublisherInformation `json:"publisher_information"`
	Message              QueueMessagePayload       `json:"message"`
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
		ID:    m.UserID,
		SubID: m.SubID,
		PublisherInformation: QueuePublisherInformation{
			MessageID: messageID,
			Source:    "restapi",
			Priority:  m.Priority,
			Metadata:  m.Metadata,
			CreatedAt: now,
		},
		Message: QueueMessagePayload{
			Command: m.Command,
			Content: m.Content,
			// 셸 도구 2종(publish-message.sh, test-integration.sh)이 RFC3339 문자열을 쓰므로 형식을 맞춘다.
			Timestamp: time.Unix(timestamp, 0).UTC().Format(time.RFC3339),
		},
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
