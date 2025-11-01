package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AutoReplyConfig represents an automatic reply configuration for email services
type AutoReplyConfig struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UserID           uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	EmailServiceID   uuid.UUID `gorm:"type:uuid;index" json:"email_service_id"` // Optional: specific service
	Name             string    `gorm:"type:varchar(255);not null" json:"name"`
	Subject          string    `gorm:"type:varchar(500);not null" json:"subject"`
	Body             string    `gorm:"type:text;not null" json:"body"`
	FromEmail        string    `gorm:"type:varchar(255)" json:"from_email"` // Optional: custom from email
	FromName         string    `gorm:"type:varchar(255)" json:"from_name"`  // Optional: custom from name
	ReplyTo          string    `gorm:"type:varchar(255)" json:"reply_to"`   // Optional: reply-to address
	IsActive         bool      `gorm:"default:true" json:"is_active"`
	TriggerOnForm    bool      `gorm:"default:true" json:"trigger_on_form"`   // Trigger on /send-form
	TriggerOnAPI     bool      `gorm:"default:false" json:"trigger_on_api"`   // Trigger on /send-email
	IncludeVariables bool      `gorm:"default:true" json:"include_variables"` // Replace {{variables}} in body
	DelaySeconds     int       `gorm:"default:0" json:"delay_seconds"`        // Delay before sending (0 = immediate)
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (a *AutoReplyConfig) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// AutoReplyLog represents a log entry for sent auto-replies
type AutoReplyLog struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	AutoReplyID    uuid.UUID `gorm:"type:uuid;not null;index" json:"autoreply_id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	RecipientEmail string    `gorm:"type:varchar(255);not null;index" json:"recipient_email"`
	Subject        string    `gorm:"type:varchar(500)" json:"subject"`
	Status         string    `gorm:"type:varchar(50);default:'sent'" json:"status"` // sent, failed
	ErrorMessage   string    `gorm:"type:text" json:"error_message,omitempty"`
	SentAt         time.Time `json:"sent_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (a *AutoReplyLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
