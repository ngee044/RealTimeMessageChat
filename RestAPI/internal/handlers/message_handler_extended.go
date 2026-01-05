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

type UpdateMessageStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// NewMessageHandlerExtended creates a new extended message handler
func NewMessageHandlerExtended(messageService *service.MessageService) *MessageHandlerExtended {
	return &MessageHandlerExtended{
		messageService: messageService,
	}
}

// GetMessage handles GET /messages/:messageID
// @Summary Get message by ID
// @Description Retrieve a single message by message ID.
// @Tags messages
// @Produce json
// @Param messageID path string true "Message ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/messages/{messageID} [get]
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
// @Summary Get messages by user
// @Description Retrieve messages for a specific user with pagination.
// @Tags messages
// @Produce json
// @Param userID path string true "User ID"
// @Param limit query int false "Pagination limit"
// @Param offset query int false "Pagination offset"
// @Success 200 {object} response.Response{data=response.PaginatedData}
// @Failure 500 {object} response.Response
// @Router /api/v1/users/{userID}/messages [get]
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
// @Summary Get recent messages
// @Description Retrieve recent messages with pagination.
// @Tags messages
// @Produce json
// @Param limit query int false "Pagination limit"
// @Param offset query int false "Pagination offset"
// @Success 200 {object} response.Response{data=response.PaginatedData}
// @Failure 500 {object} response.Response
// @Router /api/v1/messages/recent [get]
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
// @Summary Get messages by status
// @Description Retrieve messages filtered by status with pagination.
// @Tags messages
// @Produce json
// @Param status path string true "Message status"
// @Param limit query int false "Pagination limit"
// @Param offset query int false "Pagination offset"
// @Success 200 {object} response.Response{data=response.PaginatedData}
// @Failure 500 {object} response.Response
// @Router /api/v1/messages/status/{status} [get]
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
// @Summary Update message status
// @Description Update status for a message by ID.
// @Tags messages
// @Accept json
// @Produce json
// @Param messageID path string true "Message ID"
// @Param status body UpdateMessageStatusRequest true "Status update"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/messages/{messageID}/status [patch]
func (h *MessageHandlerExtended) UpdateMessageStatus(c *gin.Context) {
	messageID := c.Param("messageID")

	var req UpdateMessageStatusRequest

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
// @Summary Delete message
// @Description Delete a message by ID.
// @Tags messages
// @Produce json
// @Param messageID path string true "Message ID"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/messages/{messageID} [delete]
func (h *MessageHandlerExtended) DeleteMessage(c *gin.Context) {
	messageID := c.Param("messageID")

	if err := h.messageService.DeleteMessage(c.Request.Context(), messageID); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithMessage(c, "Message deleted successfully", nil)
}

// GetMessageStats handles GET /messages/stats
// @Summary Get message stats
// @Description Retrieve message counts by status.
// @Tags messages
// @Produce json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/messages/stats [get]
func (h *MessageHandlerExtended) GetMessageStats(c *gin.Context) {
	stats, err := h.messageService.GetMessageStats(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, stats)
}
