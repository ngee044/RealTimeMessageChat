package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/service"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/logger"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/pagination"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/response"
	"github.com/sirupsen/logrus"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// UpdateStatusRequest represents the request to update user status
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// CreateUser handles POST /users
// @Summary Create user
// @Description Create a new user.
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User to create"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Warn("Invalid request payload")
		response.ValidationError(c, "Invalid request payload: "+err.Error())
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), req.UserID, req.Username, req.Email)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "User created successfully", user)
}

// GetUser handles GET /users/:userID
// @Summary Get user
// @Description Retrieve a user by ID.
// @Tags users
// @Produce json
// @Param userID path string true "User ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/{userID} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("userID")

	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, user)
}

// UpdateStatus handles PUT /users/:userID/status
// @Summary Update user status
// @Description Update a user's status.
// @Tags users
// @Accept json
// @Produce json
// @Param userID path string true "User ID"
// @Param status body UpdateStatusRequest true "Status update"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/{userID}/status [put]
func (h *UserHandler) UpdateStatus(c *gin.Context) {
	userID := c.Param("userID")

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request payload: "+err.Error())
		return
	}

	if err := h.userService.UpdateUserStatus(c.Request.Context(), userID, req.Status); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithMessage(c, "User status updated successfully", nil)
}

// ListUsers handles GET /users
// @Summary List users
// @Description Retrieve users with pagination.
// @Tags users
// @Produce json
// @Param limit query int false "Pagination limit"
// @Param offset query int false "Pagination offset"
// @Success 200 {object} response.Response{data=response.PaginatedData}
// @Failure 500 {object} response.Response
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	params := pagination.ParseFromQuery(c)

	users, total, err := h.userService.ListUsers(c.Request.Context(), params.Limit, params.Offset)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Paginated(c, users, total, params.Limit, params.Offset)
}

// GetOnlineUsers handles GET /users/online
// @Summary List online users
// @Description Retrieve online users with pagination.
// @Tags users
// @Produce json
// @Param limit query int false "Pagination limit"
// @Param offset query int false "Pagination offset"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/online [get]
func (h *UserHandler) GetOnlineUsers(c *gin.Context) {
	params := pagination.ParseFromQuery(c)

	users, err := h.userService.GetOnlineUsers(c.Request.Context(), params.Limit, params.Offset)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Online users don't have a total count in current implementation
	response.OK(c, map[string]interface{}{
		"users":  users,
		"limit":  params.Limit,
		"offset": params.Offset,
	})
}

// DeleteUser handles DELETE /users/:userID
// @Summary Delete user
// @Description Delete a user by ID.
// @Tags users
// @Produce json
// @Param userID path string true "User ID"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/{userID} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("userID")

	if err := h.userService.DeleteUser(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithMessage(c, "User deleted successfully", nil)
}
