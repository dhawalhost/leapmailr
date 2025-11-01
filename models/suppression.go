package models

import (
	"time"

	"github.com/google/uuid"
)

// SuppressionReason represents the reason for suppression
type SuppressionReason string

const (
	SuppressionBounce      SuppressionReason = "bounce"
	SuppressionComplaint   SuppressionReason = "complaint"
	SuppressionUnsubscribe SuppressionReason = "unsubscribe"
	SuppressionManual      SuppressionReason = "manual"
)

// SuppressionSource represents where the suppression came from
type SuppressionSource string

const (
	SourceWebhook SuppressionSource = "webhook"
	SourceManual  SuppressionSource = "manual"
	SourceAPI     SuppressionSource = "api"
)

// Suppression represents an email address that should not receive emails
type Suppression struct {
	ID             uuid.UUID         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID         *uuid.UUID        `json:"user_id,omitempty" gorm:"type:uuid;index"`
	OrganizationID *uuid.UUID        `json:"organization_id,omitempty" gorm:"type:uuid;index"`
	Email          string            `json:"email" gorm:"not null;index"`
	Reason         SuppressionReason `json:"reason" gorm:"not null"`
	Source         SuppressionSource `json:"source" gorm:"not null;default:'manual'"`
	Metadata       string            `json:"metadata,omitempty" gorm:"type:jsonb"` // Additional info (bounce type, complaint details, etc.)
	CreatedAt      time.Time         `json:"created_at" gorm:"index"`
	UpdatedAt      time.Time         `json:"updated_at"`

	// Relationships
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
}

// CreateSuppressionRequest represents a request to add an email to suppression list
type CreateSuppressionRequest struct {
	Email    string                 `json:"email" binding:"required,email"`
	Reason   SuppressionReason      `json:"reason" binding:"required,oneof=bounce complaint unsubscribe manual"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SuppressionResponse represents a suppression list entry response
type SuppressionResponse struct {
	ID        uuid.UUID         `json:"id"`
	Email     string            `json:"email"`
	Reason    SuppressionReason `json:"reason"`
	Source    SuppressionSource `json:"source"`
	Metadata  string            `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

// SuppressionFilters represents filters for listing suppressions
type SuppressionFilters struct {
	Reason string `json:"reason,omitempty"`
	Source string `json:"source,omitempty"`
	Search string `json:"search,omitempty"` // Search by email
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

// BulkSuppressionRequest represents a request to add multiple emails to suppression list
type BulkSuppressionRequest struct {
	Emails   []string               `json:"emails" binding:"required,min=1"`
	Reason   SuppressionReason      `json:"reason" binding:"required,oneof=bounce complaint unsubscribe manual"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
