package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user account in the system
type User struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email         string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash  string    `json:"-" gorm:"not null"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	PlanType      string    `json:"plan_type" gorm:"default:'free'"` // free, professional, business, enterprise
	APIKey        string    `json:"api_key" gorm:"uniqueIndex"`
	PrivateKey    string    `json:"-"`
	IsActive      bool      `json:"is_active" gorm:"default:true"`
	EmailVerified bool      `json:"email_verified" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships
	Organizations []Organization `json:"organizations,omitempty" gorm:"many2many:user_organizations;"`
	EmailServices []EmailService `json:"email_services,omitempty"`
	Templates     []Template     `json:"templates,omitempty"`
	EmailLogs     []EmailLog     `json:"email_logs,omitempty"`
}

// Organization represents a multi-tenant organization
type Organization struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"not null"`
	OwnerID   uuid.UUID `json:"owner_id" gorm:"type:uuid;not null"`
	PlanType  string    `json:"plan_type" gorm:"default:'free'"`
	Settings  string    `json:"settings" gorm:"type:jsonb"` // JSON settings
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Owner         User           `json:"owner" gorm:"foreignKey:OwnerID"`
	Users         []User         `json:"users,omitempty" gorm:"many2many:user_organizations;"`
	EmailServices []EmailService `json:"email_services,omitempty"`
	Templates     []Template     `json:"templates,omitempty"`
	EmailLogs     []EmailLog     `json:"email_logs,omitempty"`
}

// UserOrganization represents the many-to-many relationship between users and organizations
type UserOrganization struct {
	UserID         uuid.UUID `json:"user_id" gorm:"type:uuid;primaryKey"`
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;primaryKey"`
	Role           string    `json:"role" gorm:"default:'member'"` // owner, admin, member, viewer
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	User         User         `json:"user" gorm:"foreignKey:UserID"`
	Organization Organization `json:"organization" gorm:"foreignKey:OrganizationID"`
}

// AuthToken represents JWT tokens for authentication
type AuthToken struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	TokenType string    `json:"token_type"` // access, refresh
	Token     string    `json:"-" gorm:"uniqueIndex"`
	ExpiresAt time.Time `json:"expires_at"`
	IsRevoked bool      `json:"is_revoked" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`

	User User `json:"user" gorm:"foreignKey:UserID"`
}

// APIKey represents API keys for external access
type APIKey struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	Name        string     `json:"name" gorm:"not null"`
	Key         string     `json:"key" gorm:"uniqueIndex;not null"`
	Permissions string     `json:"permissions" gorm:"type:jsonb"` // JSON array of permissions
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	User User `json:"user" gorm:"foreignKey:UserID"`
}

// UserSession represents active user sessions
type UserSession struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	SessionID string    `json:"session_id" gorm:"uniqueIndex;not null"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`

	User User `json:"user" gorm:"foreignKey:UserID"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest represents registration data
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User         User   `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// CreateAPIKeyRequest represents API key creation request
type CreateAPIKeyRequest struct {
	Name        string     `json:"name" binding:"required"`
	Permissions []string   `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// UpdateUserRequest represents user update request
type UpdateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email" binding:"omitempty,email"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// InviteUserRequest represents user invitation request
type InviteUserRequest struct {
	Email          string    `json:"email" binding:"required,email"`
	Role           string    `json:"role" binding:"required,oneof=admin member viewer"`
	OrganizationID uuid.UUID `json:"organization_id" binding:"required"`
}
