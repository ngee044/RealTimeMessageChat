package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/service"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/pagination"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/response"
)

// MessageHandlerExtended handles extended message-related HTTP requests
type MessageHandlerExtended struct {
	messageService *service.MessageService
}

// NewMessageHandlerExtended creates a new extended message handler
func NewMessageHandlerExtended(messageService *service.MessageService) *MessageHandlerExtended {
	return &MessageHandlerExtended{
		messageService: messageService,
	}
}

// GetMessage handles GET /messages/:messageID
func (h *MessageHandlerExtended) GetMessage(c *gin.Context) {
	messageID := c.Param("messageID")

	message, err := h.messageService.GetMessage(c.Request.Context(), messageID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, message)
}

// GetUserMessages handles GET /users/:userID/messages
func (h *MessageHandlerExtended) GetUserMessages(c *gin.Context) {
	userID := c.Param("userID")
	params := pagination.ParseFromQuery(c)

	messages, total, err := h.messageService.GetUserMessages(c.Request.Context(), userID, params.Limit, params.Offset)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Paginated(c, messages, total, params.Limit, params.Offset)
}

// GetRecentMessages handles GET /messages/recent
func (h *MessageHandlerExtended) GetRecentMessages(c *gin.Context) {
	params := pagination.ParseFromQuery(c)

	messages, total, err := h.messageService.GetRecentMessages(c.Request.Context(), params.Limit, params.Offset)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Paginated(c, messages, total, params.Limit, params.Offset)
}

// GetMessagesByStatus handles GET /messages/status/:status
func (h *MessageHandlerExtended) GetMessagesByStatus(c *gin.Context) {
	status := c.Param("status")
	params := pagination.ParseFromQuery(c)

	messages, total, err := h.messageService.GetMessagesByStatus(c.Request.Context(), status, params.Limit, params.Offset)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Paginated(c, messages, total, params.Limit, params.Offset)
}

// UpdateMessageStatus handles PATCH /messages/:messageID/status
func (h *MessageHandlerExtended) UpdateMessageStatus(c *gin.Context) {
	messageID := c.Param("messageID")

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request payload")
		return
	}

	if err := h.messageService.UpdateMessageStatus(c.Request.Context(), messageID, req.Status); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithMessage(c, "Message status updated successfully", nil)
}

// DeleteMessage handles DELETE /messages/:messageID
func (h *MessageHandlerExtended) DeleteMessage(c *gin.Context) {
	messageID := c.Param("messageID")

	if err := h.messageService.DeleteMessage(c.Request.Context(), messageID); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithMessage(c, "Message deleted successfully", nil)
}

// GetMessageStats handles GET /messages/stats
func (h *MessageHandlerExtended) GetMessageStats(c *gin.Context) {
	stats, err := h.messageService.GetMessageStats(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, stats)
}
