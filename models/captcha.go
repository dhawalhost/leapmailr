package models

import (
	"time"

	"github.com/google/uuid"
)

// CaptchaProvider represents the type of CAPTCHA service
type CaptchaProvider string

const (
	ReCaptchaV2 CaptchaProvider = "recaptcha_v2"
	HCaptcha    CaptchaProvider = "hcaptcha"
)

// CaptchaConfig stores the CAPTCHA settings for a user or organization
type CaptchaConfig struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID         *uuid.UUID      `json:"user_id,omitempty" gorm:"type:uuid"`
	OrganizationID *uuid.UUID      `json:"organization_id,omitempty" gorm:"type:uuid"`
	Provider       CaptchaProvider `json:"provider" gorm:"not null"`
	SiteKey        string          `json:"site_key" gorm:"not null"`
	SecretKey      string          `json:"-" gorm:"not null"` // Encrypted secret key
	Domains        []string        `json:"domains" gorm:"type:text[]"`
	IsActive       bool            `json:"is_active" gorm:"default:true"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`

	// Relationships
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
}

// CreateCaptchaConfigRequest represents a request to create a CAPTCHA configuration
type CreateCaptchaConfigRequest struct {
	Provider  CaptchaProvider `json:"provider" binding:"required,oneof=recaptcha_v2 hcaptcha"`
	SiteKey   string          `json:"site_key" binding:"required"`
	SecretKey string          `json:"secret_key" binding:"required"`
	Domains   []string        `json:"domains,omitempty"`
	IsActive  *bool           `json:"is_active,omitempty"`
}

// UpdateCaptchaConfigRequest represents a request to update a CAPTCHA configuration
type UpdateCaptchaConfigRequest struct {
	SiteKey   string   `json:"site_key,omitempty"`
	SecretKey string   `json:"secret_key,omitempty"`
	Domains   []string `json:"domains,omitempty"`
	IsActive  *bool    `json:"is_active,omitempty"`
}

// CaptchaConfigResponse represents a CAPTCHA configuration response (without sensitive data)
type CaptchaConfigResponse struct {
	ID        uuid.UUID       `json:"id"`
	Provider  CaptchaProvider `json:"provider"`
	SiteKey   string          `json:"site_key"`
	Domains   []string        `json:"domains"`
	IsActive  bool            `json:"is_active"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
