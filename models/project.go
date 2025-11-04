package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Project represents a project that organizes email services and templates
type Project struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	IsDefault   bool           `json:"is_default" gorm:"default:false;index"`
	Color       string         `json:"color" gorm:"default:'#3b82f6'"` // Hex color for UI
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	User          User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	EmailServices []EmailService `json:"email_services,omitempty" gorm:"foreignKey:ProjectID"`
	Templates     []Template     `json:"templates,omitempty" gorm:"foreignKey:ProjectID"`
}

// BeforeCreate will set a UUID if not provided
func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name for the Project model
func (Project) TableName() string {
	return "projects"
}
