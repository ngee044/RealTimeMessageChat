package cache

import "time"

// Default TTL values (can be overridden by configuration)
const (
	// User-related TTLs
	TTLUserData   = 1 * time.Hour
	TTLUserStatus = 24 * time.Hour

	// Message-related TTLs
	TTLMessageData   = 30 * time.Minute
	TTLMessageStatus = 1 * time.Hour

	// Session-related TTLs
	TTLSession       = 24 * time.Hour
	TTLRefreshToken  = 7 * 24 * time.Hour

	// Rate limiting TTLs
	TTLRateLimit = 1 * time.Minute

	// Default TTL
	TTLDefault = 15 * time.Minute
)

// TTLConfig holds configurable TTL values
type TTLConfig struct {
	UserData      time.Duration
	UserStatus    time.Duration
	MessageData   time.Duration
	MessageStatus time.Duration
	Session       time.Duration
	RefreshToken  time.Duration
	RateLimit     time.Duration
	Default       time.Duration
}

// DefaultTTLConfig returns default TTL configuration
func DefaultTTLConfig() *TTLConfig {
	return &TTLConfig{
		UserData:      TTLUserData,
		UserStatus:    TTLUserStatus,
		MessageData:   TTLMessageData,
		MessageStatus: TTLMessageStatus,
		Session:       TTLSession,
		RefreshToken:  TTLRefreshToken,
		RateLimit:     TTLRateLimit,
		Default:       TTLDefault,
	}
}
