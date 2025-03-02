package server

import (
	"net/http"
	"strconv"

	"db_cli/internal/user"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine      *gin.Engine
	userService *user.UserService
}

func NewServer(userService *user.UserService) *Server {
	s := &Server{
		engine:      gin.Default(),
		userService: userService,
	}
	s.initRoutes()
	return s
}

func (s *Server) initRoutes() {
	s.engine.POST("/users", s.createUserHandler)
	s.engine.GET("/users/", s.getUsersHandler)
	s.engine.GET("/users/:id", s.getUserHandler)
	s.engine.PUT("/users/:id", s.updateUserHandler)
	s.engine.DELETE("/users/:id", s.deleteUserHandler)
}

func (s *Server) Run(addr string) error {
	return s.engine.Run(addr)
}

func (s *Server) createUserHandler(c *gin.Context) {
	var request struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.userService.CreateUser(request.Name, request.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

// get Users Handler
func (s *Server) getUsersHandler(c *gin.Context) {
	users, err := s.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// get User Handler
func (s *Server) getUserHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	user, err := s.userService.GetUser(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (s *Server) updateUserHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var request struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.userService.UpdateUser(uint(id), request.Name, request.Status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (s *Server) deleteUserHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := s.userService.DeleteUser(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
