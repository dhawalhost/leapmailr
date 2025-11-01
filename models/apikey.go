package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// APIKeyPair represents a public/private key pair for SDK authentication
type APIKeyPair struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Name        string     `gorm:"type:varchar(255);not null" json:"name"`
	Description string     `gorm:"type:text" json:"description,omitempty"`
	PublicKey   string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"public_key"`
	PrivateKey  string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"private_key"` // Only shown once on creation
	Permissions string     `gorm:"type:jsonb" json:"permissions"`                             // JSON array of permissions
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	RateLimit   int        `gorm:"default:100" json:"rate_limit"` // Requests per minute
	UsageCount  int64      `gorm:"default:0" json:"usage_count"`  // Total API calls made
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (k *APIKeyPair) BeforeCreate(tx *gorm.DB) error {
	if k.ID == uuid.Nil {
		k.ID = uuid.New()
	}
	return nil
}

// APIKeyUsageLog represents usage tracking for API keys
type APIKeyUsageLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	APIKeyID   uuid.UUID `gorm:"type:uuid;not null;index" json:"api_key_id"`
	Endpoint   string    `gorm:"type:varchar(255);not null" json:"endpoint"`
	Method     string    `gorm:"type:varchar(10);not null" json:"method"`
	StatusCode int       `gorm:"not null" json:"status_code"`
	IPAddress  string    `gorm:"type:varchar(50)" json:"ip_address"`
	UserAgent  string    `gorm:"type:text" json:"user_agent,omitempty"`
	ResponseMs int64     `json:"response_ms"` // Response time in milliseconds
	CreatedAt  time.Time `json:"created_at"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (l *APIKeyUsageLog) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

// CreateAPIKeyPairRequest represents a request to create a new API key pair
type CreateAPIKeyPairRequest struct {
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description,omitempty"`
	Permissions []string   `json:"permissions,omitempty"`
	RateLimit   int        `json:"rate_limit,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// UpdateAPIKeyPairRequest represents a request to update an API key pair
type UpdateAPIKeyPairRequest struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	RateLimit   int      `json:"rate_limit,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
}

// APIKeyPairResponse represents the API response for key pair (hides private key after creation)
type APIKeyPairResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	PublicKey   string     `json:"public_key"`
	PrivateKey  string     `json:"private_key,omitempty"` // Only included on creation
	Permissions []string   `json:"permissions,omitempty"`
	IsActive    bool       `json:"is_active"`
	RateLimit   int        `json:"rate_limit"`
	UsageCount  int64      `json:"usage_count"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
}
