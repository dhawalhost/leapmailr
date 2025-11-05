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
	PrivateKey    string    `json:"-"` // Encrypted
	IsActive      bool      `json:"is_active" gorm:"default:true"`
	EmailVerified bool      `json:"email_verified" gorm:"default:false"`

	// Security - Account Lockout (GAP-SEC-003)
	FailedLoginAttempts int        `json:"-" gorm:"default:0"`
	LockedUntil         *time.Time `json:"-"`
	LastLoginAttempt    *time.Time `json:"-"`
	LastLoginSuccess    *time.Time `json:"last_login_success,omitempty"`

	// Multi-Factor Authentication (GAP-SEC-001)
	MFAEnabled     bool       `json:"mfa_enabled" gorm:"default:false"`
	MFASecret      string     `json:"-"` // Encrypted TOTP secret
	MFABackupCodes string     `json:"-"` // Encrypted JSON array of backup codes
	MFAVerifiedAt  *time.Time `json:"mfa_verified_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

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

// AuditLog represents security audit log (GAP-SEC-008)
type AuditLog struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID     *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid;index"`
	Action     string     `json:"action" gorm:"not null;index"` // login, logout, password_change, data_access, api_key_created, etc.
	Resource   string     `json:"resource,omitempty"`           // user, email_service, template, etc.
	ResourceID *uuid.UUID `json:"resource_id,omitempty" gorm:"type:uuid"`
	IPAddress  string     `json:"ip_address"`
	UserAgent  string     `json:"user_agent"`
	Status     string     `json:"status" gorm:"not null"`              // success, failure
	Details    string     `json:"details,omitempty" gorm:"type:jsonb"` // Additional context as JSON
	Timestamp  time.Time  `json:"timestamp" gorm:"not null;index"`
	CreatedAt  time.Time  `json:"created_at"`

	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// PasswordHistory stores hashed passwords for password reuse prevention (GAP-SEC-002)
type PasswordHistory struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	PasswordHash string    `json:"-" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`

	User User `json:"user" gorm:"foreignKey:UserID"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=12,max=128"`
}

// RegisterRequest represents registration data
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email,max=255"`
	Password  string `json:"password" binding:"required,min=12,max=128"`
	FirstName string `json:"first_name" binding:"required,max=100,alphanumunicode"`
	LastName  string `json:"last_name" binding:"required,max=100,alphanumunicode"`
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
	Name        string     `json:"name" binding:"required,max=100,alphanumunicode"`
	Permissions []string   `json:"permissions" binding:"max=20,dive,oneof=read write admin email template contact"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// UpdateUserRequest represents user update request
type UpdateUserRequest struct {
	FirstName string `json:"first_name" binding:"omitempty,max=100,alphanumunicode"`
	LastName  string `json:"last_name" binding:"omitempty,max=100,alphanumunicode"`
	Email     string `json:"email" binding:"omitempty,email,max=255"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=12,max=128"`
	NewPassword     string `json:"new_password" binding:"required,min=12,max=128"`
}

// InviteUserRequest represents user invitation request
type InviteUserRequest struct {
	Email          string    `json:"email" binding:"required,email,max=255"`
	Role           string    `json:"role" binding:"required,oneof=admin member viewer"`
	OrganizationID uuid.UUID `json:"organization_id" binding:"required,uuid"`
}

// MFA Models (GAP-SEC-001)

// MFASetupRequest represents MFA setup initiation request
type MFASetupRequest struct {
	Password string `json:"password" binding:"required,min=12,max=128"`
}

// MFASetupResponse represents MFA setup response with QR code data
type MFASetupResponse struct {
	Secret        string   `json:"secret"`           // For manual entry
	QRCodeDataURL string   `json:"qr_code_data_url"` // Base64-encoded QR code image
	BackupCodes   []string `json:"backup_codes"`     // One-time use backup codes
}

// MFAVerifySetupRequest represents MFA setup verification request
type MFAVerifySetupRequest struct {
	Code string `json:"code" binding:"required,len=6,numeric"`
}

// MFALoginRequest represents login with MFA code
type MFALoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=12,max=128"`
	Code     string `json:"code" binding:"required,len=6,numeric"`
}

// MFABackupCodeRequest represents login with backup code
type MFABackupCodeRequest struct {
	Email      string `json:"email" binding:"required,email,max=255"`
	Password   string `json:"password" binding:"required,min=12,max=128"`
	BackupCode string `json:"backup_code" binding:"required,len=8,alphanum"`
}

// MFADisableRequest represents MFA disable request
type MFADisableRequest struct {
	Password string `json:"password" binding:"required,min=12,max=128"`
	Code     string `json:"code" binding:"required,len=6,numeric"`
}

// MFARegenerateBackupCodesRequest represents backup code regeneration request
type MFARegenerateBackupCodesRequest struct {
	Password string `json:"password" binding:"required,min=12,max=128"`
	Code     string `json:"code" binding:"required,len=6,numeric"`
}

// MFARegenerateBackupCodesResponse represents new backup codes
type MFARegenerateBackupCodesResponse struct {
	BackupCodes []string `json:"backup_codes"`
}
