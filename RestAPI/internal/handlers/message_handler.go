package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/models"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/services"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/logger"
	"github.com/sirupsen/logrus"
)

// MessageHandler handles message-related HTTP requests
type MessageHandler struct {
	rabbitMQ *services.RabbitMQService
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(rabbitMQ *services.RabbitMQService) *MessageHandler {
	return &MessageHandler{
		rabbitMQ: rabbitMQ,
	}
}

// SendMessage handles the POST /api/v1/messages/send endpoint
// @Summary Send a message to RabbitMQ
// @Description Publishes a message to RabbitMQ queue for processing
// @Tags messages
// @Accept json
// @Produce json
// @Param message body models.MessageRequest true "Message to send"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/messages/send [post]
func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req models.MessageRequest

	// Bind JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithFields(logrus.Fields{
			"error":     err.Error(),
			"client_ip": c.ClientIP(),
		}).Warn("Invalid request payload")

		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Invalid request payload: "+err.Error(),
			"INVALID_PAYLOAD",
		))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		logger.WithFields(logrus.Fields{
			"error":   err.Error(),
			"user_id": req.UserID,
		}).Warn("Message validation failed")

		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Validation failed: "+err.Error(),
			"VALIDATION_ERROR",
		))
		return
	}

	// Generate unique message ID
	messageID := uuid.New().String()

	// Convert to queue message
	queueMsg := req.ToQueueMessage(messageID)

	// Serialize to JSON
	msgBytes, err := queueMsg.ToJSON()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error":      err.Error(),
			"message_id": messageID,
			"user_id":    req.UserID,
		}).Error("Failed to serialize message")

		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to process message",
			"SERIALIZATION_ERROR",
		))
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Publish to RabbitMQ
	if err := h.rabbitMQ.Publish(ctx, msgBytes); err != nil {
		logger.WithFields(logrus.Fields{
			"error":      err.Error(),
			"message_id": messageID,
			"user_id":    req.UserID,
			"command":    req.Command,
		}).Error("Failed to publish message to RabbitMQ")

		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Failed to send message",
			"PUBLISH_ERROR",
		))
		return
	}

	// Log success
	logger.WithFields(logrus.Fields{
		"message_id": messageID,
		"user_id":    req.UserID,
		"command":    req.Command,
		"priority":   req.Priority,
	}).Info("Message published successfully")

	// Return success response
	c.JSON(http.StatusOK, models.NewMessageResponse(
		messageID,
		"Message sent successfully",
		gin.H{
			"queue_name": h.rabbitMQ.QueueName(),
			"priority":   req.Priority,
		},
	))
}
