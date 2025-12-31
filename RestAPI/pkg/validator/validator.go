package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// Validator provides validation utilities
type Validator struct {
	errors map[string][]string
}

// New creates a new validator
func New() *Validator {
	return &Validator{
		errors: make(map[string][]string),
	}
}

// AddError adds a validation error for a field
func (v *Validator) AddError(field, message string) {
	v.errors[field] = append(v.errors[field], message)
}

// Check adds an error if condition is false
func (v *Validator) Check(condition bool, field, message string) {
	if !condition {
		v.AddError(field, message)
	}
}

// IsValid returns true if no validation errors
func (v *Validator) IsValid() bool {
	return len(v.errors) == 0
}

// Errors returns all validation errors
func (v *Validator) Errors() map[string][]string {
	return v.errors
}

// ErrorString returns a formatted error string
func (v *Validator) ErrorString() string {
	var msgs []string
	for field, errs := range v.errors {
		for _, err := range errs {
			msgs = append(msgs, fmt.Sprintf("%s: %s", field, err))
		}
	}
	return strings.Join(msgs, "; ")
}

// Required checks if a string is not empty
func Required(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MinLength checks minimum string length
func MinLength(value string, min int) bool {
	return len(strings.TrimSpace(value)) >= min
}

// MaxLength checks maximum string length
func MaxLength(value string, max int) bool {
	return len(strings.TrimSpace(value)) <= max
}

// InRange checks if an integer is within range
func InRange(value, min, max int) bool {
	return value >= min && value <= max
}

// Matches checks if a string matches a regex pattern
func Matches(value string, pattern *regexp.Regexp) bool {
	return pattern.MatchString(value)
}

// Email validates email format
func Email(value string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(value)
}

// AlphaNumeric checks if string contains only alphanumeric characters
func AlphaNumeric(value string) bool {
	alphaNumRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	return alphaNumRegex.MatchString(value)
}

// UUID validates UUID format
func UUID(value string) bool {
	uuidRegex := regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)
	return uuidRegex.MatchString(strings.ToLower(value))
}

// In checks if value is in the allowed list
func In[T comparable](value T, allowed ...T) bool {
	for _, v := range allowed {
		if value == v {
			return true
		}
	}
	return false
}

// NotIn checks if value is not in the disallowed list
func NotIn[T comparable](value T, disallowed ...T) bool {
	for _, v := range disallowed {
		if value == v {
			return false
		}
	}
	return true
}

// ValidatePriority validates message priority (1-3)
func ValidatePriority(priority int) bool {
	return InRange(priority, 1, 3)
}

// ValidateUserID validates user ID format
func ValidateUserID(userID string) bool {
	return Required(userID) && MinLength(userID, 3) && MaxLength(userID, 50) && AlphaNumeric(userID)
}

// ValidateCommand validates command format
func ValidateCommand(command string) bool {
	return Required(command) && MinLength(command, 2) && MaxLength(command, 50)
}

// ValidateContent validates message content
func ValidateContent(content string) bool {
	return Required(content) && MaxLength(content, 10000)
}
