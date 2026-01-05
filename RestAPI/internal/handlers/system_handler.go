package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hyunkyulee/RealTimeMessageChat/RestAPI/internal/services"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type SystemHandler struct {
	rabbitMQ *services.RabbitMQService
	redis    *services.RedisService
	db       *services.DatabaseService
	version  string
}

func NewSystemHandler(
	rabbitMQ *services.RabbitMQService,
	redis *services.RedisService,
	db *services.DatabaseService,
	version string,
) *SystemHandler {
	return &SystemHandler{
		rabbitMQ: rabbitMQ,
		redis:    redis,
		db:       db,
		version:  version,
	}
}

// @Summary Health check
// @Description Check API and dependency health status.
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /health [get]
func (h *SystemHandler) Health(c *gin.Context) {
	rabbitMQHealthy := h.rabbitMQ.IsHealthy()
	redisHealthy := true
	dbHealthy := true

	if h.redis != nil {
		redisHealthy = h.redis.IsHealthy(c.Request.Context())
	}

	if h.db != nil {
		dbHealthy = h.db.IsHealthy()
	}

	status := "healthy"
	httpStatus := http.StatusOK

	if !rabbitMQHealthy || !dbHealthy {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"status": status,
		"services": gin.H{
			"rabbitmq": rabbitMQHealthy,
			"redis":    redisHealthy,
			"database": dbHealthy,
		},
		"timestamp": time.Now().Unix(),
		"version":   h.version,
	})
}

// @Summary Service info
// @Description Returns service metadata and docs path.
// @Tags system
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router / [get]
func (h *SystemHandler) Root(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "RealTimeMessageChat REST API",
		"version": h.version,
		"status":  "running",
		"docs":    "/api/v1",
	})
}

// @Summary Prometheus metrics
// @Description Exposes Prometheus metrics when enabled.
// @Tags metrics
// @Produce text/plain
// @Success 200 {string} string
// @Router /metrics [get]
func MetricsHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}
