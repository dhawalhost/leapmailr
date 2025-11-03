package models

import (
	"time"

	"github.com/google/uuid"
)

// EmailService represents an email service provider configuration
type EmailService struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID         *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty" gorm:"type:uuid"`
	Name           string     `json:"name" gorm:"not null"`
	Provider       string     `json:"provider" gorm:"not null"`   // smtp, sendgrid, mailgun, ses, etc.
	Configuration  string     `json:"-" gorm:"type:jsonb"`        // Encrypted JSON config (SMTP credentials)
	FromEmail      string     `json:"from_email" gorm:"not null"` // Sender email address (shown to recipients)
	FromName       string     `json:"from_name"`                  // Sender name (shown to recipients)
	ReplyToEmail   string     `json:"reply_to_email,omitempty"`   // Reply-to email address
	IsDefault      bool       `json:"is_default" gorm:"default:false"`
	Status         string     `json:"status" gorm:"default:'active'"` // active, inactive, error
	LastError      string     `json:"last_error,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relationships
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	EmailLogs    []EmailLog    `json:"email_logs,omitempty" gorm:"foreignKey:ServiceID"`
}

// Template represents an email template
type Template struct {
	ID                  uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID              *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid"`
	OrganizationID      *uuid.UUID `json:"organization_id,omitempty" gorm:"type:uuid"`
	Name                string     `json:"name" gorm:"not null"`
	Description         string     `json:"description"`
	Subject             string     `json:"subject"`
	HTMLContent         string     `json:"html_content" gorm:"type:text"`
	TextContent         string     `json:"text_content" gorm:"type:text"`
	Variables           string     `json:"variables" gorm:"type:jsonb"`            // JSON array of variable names
	Category            string     `json:"category" gorm:"default:'custom'"`       // contact_form, newsletter, transactional, notification, custom, auto_reply
	IsDefault           bool       `json:"is_default" gorm:"default:false"`        // System default template
	IsPublic            bool       `json:"is_public" gorm:"default:false"`         // Available to all users
	ClonedFrom          *uuid.UUID `json:"cloned_from,omitempty" gorm:"type:uuid"` // ID of original template if cloned
	Version             int        `json:"version" gorm:"default:1"`
	IsActive            bool       `json:"is_active" gorm:"default:true"`
	UsageCount          int64      `json:"usage_count" gorm:"default:0"`                      // Track how many times template is used
	PreviewImage        string     `json:"preview_image,omitempty"`                           // URL to template preview image
	FromEmail           string     `json:"from_email,omitempty"`                              // Override service from_email if set
	FromName            string     `json:"from_name,omitempty"`                               // Override service from_name if set
	ReplyToEmail        string     `json:"reply_to_email,omitempty"`                          // Override service reply_to_email if set
	AutoReplyEnabled    bool       `json:"auto_reply_enabled" gorm:"default:false"`           // Enable auto-reply for this template
	AutoReplyTemplateID *uuid.UUID `json:"auto_reply_template_id,omitempty" gorm:"type:uuid"` // Template to use for auto-reply
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`

	// Relationships
	User              *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Organization      *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	EmailLogs         []EmailLog    `json:"email_logs,omitempty" gorm:"foreignKey:TemplateID"`
	AutoReplyTemplate *Template     `json:"auto_reply_template,omitempty" gorm:"foreignKey:AutoReplyTemplateID"`
}

// EmailLog represents an email sending log
type EmailLog struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID         *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty" gorm:"type:uuid"`
	TemplateID     *uuid.UUID `json:"template_id,omitempty" gorm:"type:uuid"`
	ServiceID      *uuid.UUID `json:"service_id,omitempty" gorm:"type:uuid"`
	MessageID      string     `json:"message_id"` // Provider message ID
	FromEmail      string     `json:"from_email" gorm:"not null"`
	FromName       string     `json:"from_name"`
	ToEmail        string     `json:"to_email" gorm:"not null"`
	ToName         string     `json:"to_name"`
	Subject        string     `json:"subject"`
	Status         string     `json:"status" gorm:"default:'queued'"` // queued, sent, delivered, bounced, failed, opened, clicked
	ErrorMessage   string     `json:"error_message,omitempty"`
	Metadata       string     `json:"metadata,omitempty" gorm:"type:jsonb"` // Additional data
	SentAt         *time.Time `json:"sent_at"`
	DeliveredAt    *time.Time `json:"delivered_at"`
	OpenedAt       *time.Time `json:"opened_at"`
	ClickedAt      *time.Time `json:"clicked_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relationships
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	Template     *Template     `json:"template,omitempty" gorm:"foreignKey:TemplateID;constraint:OnDelete:SET NULL"`
	Service      *EmailService `json:"service,omitempty" gorm:"foreignKey:ServiceID;constraint:OnDelete:SET NULL"`
}

// EmailRequest represents a request to send an email
type EmailRequest struct {
	ServiceID      *uuid.UUID             `json:"service_id,omitempty"`
	TemplateID     uuid.UUID              `json:"template_id" binding:"required"`
	ToEmail        string                 `json:"to_email" binding:"required,email"`
	ToName         string                 `json:"to_name,omitempty"`
	FromEmail      string                 `json:"from_email,omitempty"`
	FromName       string                 `json:"from_name,omitempty"`
	Subject        string                 `json:"subject,omitempty"`
	TemplateParams map[string]interface{} `json:"template_params,omitempty"`
	Attachments    []EmailAttachment      `json:"attachments,omitempty"`
	ScheduleAt     *time.Time             `json:"schedule_at,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	// Auto-reply configuration (template-based)
	AutoReplyEnabled    bool       `json:"auto_reply_enabled,omitempty"`
	AutoReplyTemplateID *uuid.UUID `json:"auto_reply_template_id,omitempty"`
}

// BulkEmailRequest represents a request to send bulk emails
type BulkEmailRequest struct {
	ServiceID     *uuid.UUID             `json:"service_id,omitempty"`
	TemplateID    uuid.UUID              `json:"template_id" binding:"required"`
	Recipients    []EmailRecipient       `json:"recipients" binding:"required,min=1"`
	FromEmail     string                 `json:"from_email,omitempty"`
	FromName      string                 `json:"from_name,omitempty"`
	Subject       string                 `json:"subject,omitempty"`
	DefaultParams map[string]interface{} `json:"default_params,omitempty"`
	ScheduleAt    *time.Time             `json:"schedule_at,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
}

// EmailRecipient represents a recipient in bulk email
type EmailRecipient struct {
	Email          string                 `json:"email" binding:"required,email"`
	Name           string                 `json:"name,omitempty"`
	TemplateParams map[string]interface{} `json:"template_params,omitempty"`
}

// EmailAttachment represents an email attachment
type EmailAttachment struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"content_type"`
	Content     []byte `json:"content" binding:"required"`
	Size        int64  `json:"size"`
}

// WebhookEvent represents a webhook event
type WebhookEvent struct {
	ID         uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	EmailLogID uuid.UUID              `json:"email_log_id" gorm:"type:uuid;not null"`
	Event      string                 `json:"event" gorm:"not null"` // sent, delivered, bounced, opened, clicked, etc.
	Data       map[string]interface{} `json:"data" gorm:"type:jsonb"`
	CreatedAt  time.Time              `json:"created_at"`

	EmailLog EmailLog `json:"email_log" gorm:"foreignKey:EmailLogID"`
}

// Template management models

// CreateTemplateRequest represents a request to create a template
type CreateTemplateRequest struct {
	Name                string     `json:"name" binding:"required"`
	Description         string     `json:"description,omitempty"`
	Subject             string     `json:"subject" binding:"required"`
	HTMLContent         string     `json:"html_content,omitempty"`
	TextContent         string     `json:"text_content,omitempty"`
	Variables           string     `json:"variables,omitempty"` // JSON array of variable names
	FromEmail           string     `json:"from_email,omitempty" binding:"omitempty,email"`
	FromName            string     `json:"from_name,omitempty"`
	ReplyToEmail        string     `json:"reply_to_email,omitempty" binding:"omitempty,email"`
	AutoReplyEnabled    bool       `json:"auto_reply_enabled,omitempty"`
	AutoReplyTemplateID *uuid.UUID `json:"auto_reply_template_id,omitempty"`
}

// UpdateTemplateRequest represents a request to update a template
type UpdateTemplateRequest struct {
	Name                string     `json:"name,omitempty"`
	Description         string     `json:"description,omitempty"`
	Subject             string     `json:"subject,omitempty"`
	HTMLContent         string     `json:"html_content,omitempty"`
	TextContent         string     `json:"text_content,omitempty"`
	Variables           string     `json:"variables,omitempty"`
	IsActive            *bool      `json:"is_active,omitempty"`
	FromEmail           string     `json:"from_email,omitempty" binding:"omitempty,email"`
	FromName            string     `json:"from_name,omitempty"`
	ReplyToEmail        string     `json:"reply_to_email,omitempty" binding:"omitempty,email"`
	AutoReplyEnabled    *bool      `json:"auto_reply_enabled,omitempty"`
	AutoReplyTemplateID *uuid.UUID `json:"auto_reply_template_id,omitempty"`
}

// TemplateFilters represents filters for listing templates
type TemplateFilters struct {
	IsActive *bool  `json:"is_active,omitempty"`
	Name     string `json:"name,omitempty"`
	OrderBy  string `json:"order_by,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

// TemplateTestResult represents the result of testing a template
type TemplateTestResult struct {
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
	Subject     string `json:"subject,omitempty"`
	HTMLContent string `json:"html_content,omitempty"`
	TextContent string `json:"text_content,omitempty"`
}

// TemplateVersion represents a template version (for version history)
type TemplateVersion struct {
	ID          uuid.UUID `json:"id"`
	Version     int       `json:"version"`
	Name        string    `json:"name"`
	Subject     string    `json:"subject"`
	HTMLContent string    `json:"html_content"`
	TextContent string    `json:"text_content"`
	CreatedAt   time.Time `json:"created_at"`
	IsActive    bool      `json:"is_active"`
}

// Email Service management models

// CreateEmailServiceRequest represents a request to create an email service
type CreateEmailServiceRequest struct {
	Name          string                 `json:"name" binding:"required"`
	Provider      string                 `json:"provider" binding:"required,oneof=smtp sendgrid mailgun ses postmark resend"`
	Configuration map[string]interface{} `json:"configuration" binding:"required"`
	FromEmail     string                 `json:"from_email" binding:"required,email"`
	FromName      string                 `json:"from_name,omitempty"`
	ReplyToEmail  string                 `json:"reply_to_email,omitempty" binding:"omitempty,email"`
	IsDefault     bool                   `json:"is_default,omitempty"`
}

// UpdateEmailServiceRequest represents a request to update an email service
type UpdateEmailServiceRequest struct {
	Name          string                 `json:"name,omitempty"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
	FromEmail     string                 `json:"from_email,omitempty" binding:"omitempty,email"`
	FromName      string                 `json:"from_name,omitempty"`
	ReplyToEmail  string                 `json:"reply_to_email,omitempty" binding:"omitempty,email"`
	IsDefault     *bool                  `json:"is_default,omitempty"`
	Status        string                 `json:"status,omitempty" binding:"omitempty,oneof=active inactive"`
}

// EmailServiceResponse represents an email service response (without sensitive data)
type EmailServiceResponse struct {
	ID            uuid.UUID         `json:"id"`
	Name          string            `json:"name"`
	Provider      string            `json:"provider"`
	ProviderLogo  string            `json:"provider_logo,omitempty"`  // Provider logo icon identifier
	ProviderColor string            `json:"provider_color,omitempty"` // Provider brand color
	FromEmail     string            `json:"from_email"`
	FromName      string            `json:"from_name"`
	ReplyToEmail  string            `json:"reply_to_email,omitempty"`
	IsDefault     bool              `json:"is_default"`
	Status        string            `json:"status"`
	LastError     string            `json:"last_error,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	ConfigPreview map[string]string `json:"config_preview,omitempty"` // Safe preview of config (e.g., masked SMTP server)
}

// EmailServiceFilters represents filters for listing email services
type EmailServiceFilters struct {
	Provider string `json:"provider,omitempty"`
	Status   string `json:"status,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

// TestEmailServiceRequest represents a request to test an email service
type TestEmailServiceRequest struct {
	ToEmail string `json:"to_email" binding:"required,email"`
}

// TestEmailServiceResponse represents a response from testing an email service
type TestEmailServiceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
