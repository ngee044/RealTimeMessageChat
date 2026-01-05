package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// Default pagination values
const (
	DefaultLimit  = 20
	DefaultOffset = 0
	MaxLimit      = 100
	MinLimit      = 1
)

// Params holds pagination parameters
type Params struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// Result holds pagination result metadata
type Result struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	TotalPages int         `json:"total_pages"`
	Page       int         `json:"page"`
	HasNext    bool        `json:"has_next"`
	HasPrev    bool        `json:"has_prev"`
}

// ParseFromQuery extracts and validates pagination params from query string
func ParseFromQuery(c *gin.Context) Params {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(DefaultLimit)))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", strconv.Itoa(DefaultOffset)))

	return Normalize(limit, offset)
}

// ParseFromQueryWithDefaults extracts pagination params with custom defaults
func ParseFromQueryWithDefaults(c *gin.Context, defaultLimit, maxLimit int) Params {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", strconv.Itoa(DefaultOffset)))

	return NormalizeWithMax(limit, offset, maxLimit)
}

// Normalize validates and normalizes pagination params
func Normalize(limit, offset int) Params {
	return NormalizeWithMax(limit, offset, MaxLimit)
}

// NormalizeWithMax validates and normalizes pagination params with custom max
func NormalizeWithMax(limit, offset, maxLimit int) Params {
	if limit < MinLimit {
		limit = DefaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}

	return Params{
		Limit:  limit,
		Offset: offset,
	}
}

// NewResult creates a pagination result from items and total count
func NewResult(items interface{}, total int64, params Params) Result {
	totalPages := int(total) / params.Limit
	if int(total)%params.Limit > 0 {
		totalPages++
	}

	currentPage := (params.Offset / params.Limit) + 1
	hasNext := params.Offset+params.Limit < int(total)
	hasPrev := params.Offset > 0

	return Result{
		Items:      items,
		Total:      total,
		Limit:      params.Limit,
		Offset:     params.Offset,
		TotalPages: totalPages,
		Page:       currentPage,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}
}

// CalculateOffset calculates offset from page number
func CalculateOffset(page, limit int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * limit
}
