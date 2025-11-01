package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Contact represents a contact collected from form submissions
type Contact struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UserID          uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Email           string    `gorm:"type:varchar(255);not null;index" json:"email"`
	Name            string    `gorm:"type:varchar(255)" json:"name,omitempty"`
	Phone           string    `gorm:"type:varchar(50)" json:"phone,omitempty"`
	Company         string    `gorm:"type:varchar(255)" json:"company,omitempty"`
	Source          string    `gorm:"type:varchar(100)" json:"source"` // form, api, import
	TemplateID      uuid.UUID `gorm:"type:uuid;index" json:"template_id,omitempty"`
	ServiceID       uuid.UUID `gorm:"type:uuid;index" json:"service_id,omitempty"`
	Metadata        string    `gorm:"type:jsonb" json:"metadata,omitempty"` // Additional custom fields
	Tags            string    `gorm:"type:jsonb" json:"tags,omitempty"`     // JSON array of tags
	IsSubscribed    bool      `gorm:"default:true" json:"is_subscribed"`
	SubmissionCount int       `gorm:"default:1" json:"submission_count"` // Number of times this contact submitted
	LastSubmittedAt time.Time `json:"last_submitted_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (c *Contact) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// ContactList represents a custom contact list/segment
type ContactList struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	ContactIDs  string    `gorm:"type:jsonb" json:"contact_ids"` // JSON array of contact IDs
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (cl *ContactList) BeforeCreate(tx *gorm.DB) error {
	if cl.ID == uuid.Nil {
		cl.ID = uuid.New()
	}
	return nil
}

// CreateContactRequest represents a request to create a contact
type CreateContactRequest struct {
	Email    string            `json:"email" binding:"required,email"`
	Name     string            `json:"name,omitempty"`
	Phone    string            `json:"phone,omitempty"`
	Company  string            `json:"company,omitempty"`
	Source   string            `json:"source,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Tags     []string          `json:"tags,omitempty"`
}

// UpdateContactRequest represents a request to update a contact
type UpdateContactRequest struct {
	Name         string            `json:"name,omitempty"`
	Phone        string            `json:"phone,omitempty"`
	Company      string            `json:"company,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Tags         []string          `json:"tags,omitempty"`
	IsSubscribed *bool             `json:"is_subscribed,omitempty"`
}

// ImportContactsRequest represents a bulk import request
type ImportContactsRequest struct {
	Contacts []CreateContactRequest `json:"contacts" binding:"required,min=1"`
	Source   string                 `json:"source,omitempty"`
}

// ContactResponse represents the API response for a contact
type ContactResponse struct {
	ID              uuid.UUID         `json:"id"`
	Email           string            `json:"email"`
	Name            string            `json:"name,omitempty"`
	Phone           string            `json:"phone,omitempty"`
	Company         string            `json:"company,omitempty"`
	Source          string            `json:"source"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	Tags            []string          `json:"tags,omitempty"`
	IsSubscribed    bool              `json:"is_subscribed"`
	SubmissionCount int               `json:"submission_count"`
	LastSubmittedAt time.Time         `json:"last_submitted_at"`
	CreatedAt       time.Time         `json:"created_at"`
}
