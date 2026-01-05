package cache

import "fmt"

// Cache key prefixes
const (
	PrefixUser          = "user"
	PrefixUserStatus    = "user:status"
	PrefixMessage       = "message"
	PrefixMessageStatus = "message:status"
	PrefixSession       = "session"
	PrefixRateLimit     = "ratelimit"
)

// UserKey generates a cache key for user data
func UserKey(userID string) string {
	return fmt.Sprintf("%s:%s", PrefixUser, userID)
}

// UserStatusKey generates a cache key for user status
func UserStatusKey(userID string) string {
	return fmt.Sprintf("%s:%s", PrefixUserStatus, userID)
}

// MessageKey generates a cache key for message data
func MessageKey(messageID string) string {
	return fmt.Sprintf("%s:%s", PrefixMessage, messageID)
}

// MessageStatusKey generates a cache key for message status
func MessageStatusKey(messageID string) string {
	return fmt.Sprintf("%s:%s", PrefixMessageStatus, messageID)
}

// SessionKey generates a cache key for session data
func SessionKey(sessionID string) string {
	return fmt.Sprintf("%s:%s", PrefixSession, sessionID)
}

// RateLimitKey generates a cache key for rate limiting
func RateLimitKey(identifier string) string {
	return fmt.Sprintf("%s:%s", PrefixRateLimit, identifier)
}

// CustomKey generates a custom cache key with given prefix and parts
func CustomKey(prefix string, parts ...string) string {
	key := prefix
	for _, part := range parts {
		key = fmt.Sprintf("%s:%s", key, part)
	}
	return key
}
