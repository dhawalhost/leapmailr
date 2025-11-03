package service

import (
	"fmt"
	"time"

	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TemplateService handles template operations
type TemplateService struct {
	db *gorm.DB
}

// NewTemplateService creates a new template service
func NewTemplateService() *TemplateService {
	return &TemplateService{
		db: database.GetDB(),
	}
}

// CreateTemplate creates a new email template
func (s *TemplateService) CreateTemplate(req models.CreateTemplateRequest, userID uuid.UUID) (*models.Template, error) {
	// Handle empty variables - set to empty JSON array instead of empty string
	variables := req.Variables
	if variables == "" {
		variables = "[]"
	}

	template := models.Template{
		UserID:              &userID,
		Name:                req.Name,
		Description:         req.Description,
		Subject:             req.Subject,
		HTMLContent:         req.HTMLContent,
		TextContent:         req.TextContent,
		Variables:           variables,
		FromEmail:           req.FromEmail,
		FromName:            req.FromName,
		ReplyToEmail:        req.ReplyToEmail,
		AutoReplyEnabled:    req.AutoReplyEnabled,
		AutoReplyTemplateID: req.AutoReplyTemplateID,
		Version:             1,
		IsActive:            true,
	}

	if err := s.db.Create(&template).Error; err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return &template, nil
}

// GetTemplate retrieves a template by ID
func (s *TemplateService) GetTemplate(templateID, userID uuid.UUID) (*models.Template, error) {
	var template models.Template
	err := s.db.Where("id = ? AND (user_id = ? OR organization_id IN (SELECT organization_id FROM user_organizations WHERE user_id = ?))",
		templateID, userID, userID).First(&template).Error
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	return &template, nil
}

// UpdateTemplate updates an existing template
func (s *TemplateService) UpdateTemplate(templateID uuid.UUID, req models.UpdateTemplateRequest, userID uuid.UUID) (*models.Template, error) {
	var template models.Template
	if err := s.db.Where("id = ? AND (user_id = ? OR organization_id IN (SELECT organization_id FROM user_organizations WHERE user_id = ?))",
		templateID, userID, userID).First(&template).Error; err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Update fields if provided
	if req.Name != "" {
		template.Name = req.Name
	}
	if req.Description != "" {
		template.Description = req.Description
	}
	if req.Subject != "" {
		template.Subject = req.Subject
	}
	if req.HTMLContent != "" {
		template.HTMLContent = req.HTMLContent
	}
	if req.TextContent != "" {
		template.TextContent = req.TextContent
	}
	if req.Variables != "" {
		template.Variables = req.Variables
	}
	if req.FromEmail != "" {
		template.FromEmail = req.FromEmail
	}
	if req.FromName != "" {
		template.FromName = req.FromName
	}
	if req.ReplyToEmail != "" {
		template.ReplyToEmail = req.ReplyToEmail
	}
	if req.AutoReplyEnabled != nil {
		template.AutoReplyEnabled = *req.AutoReplyEnabled
	}
	if req.AutoReplyTemplateID != nil {
		template.AutoReplyTemplateID = req.AutoReplyTemplateID
	}
	if req.IsActive != nil {
		template.IsActive = *req.IsActive
	}

	// Increment version on content changes
	if req.HTMLContent != "" || req.TextContent != "" || req.Subject != "" {
		template.Version++
	}

	template.UpdatedAt = time.Now()

	if err := s.db.Save(&template).Error; err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	return &template, nil
}

// DeleteTemplate deletes a template
func (s *TemplateService) DeleteTemplate(templateID, userID uuid.UUID) error {
	result := s.db.Where("id = ? AND (user_id = ? OR organization_id IN (SELECT organization_id FROM user_organizations WHERE user_id = ?))",
		templateID, userID, userID).Delete(&models.Template{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete template: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("template not found")
	}

	return nil
}

// ListTemplates retrieves all templates for a user
func (s *TemplateService) ListTemplates(userID uuid.UUID, filters models.TemplateFilters) ([]models.Template, error) {
	var templates []models.Template

	query := s.db.Where("user_id = ? OR organization_id IN (SELECT organization_id FROM user_organizations WHERE user_id = ?)",
		userID, userID)

	// Apply filters
	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}
	if filters.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filters.Name+"%")
	}

	// Apply ordering
	orderBy := "created_at DESC"
	if filters.OrderBy != "" {
		orderBy = filters.OrderBy
	}
	query = query.Order(orderBy)

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	if err := query.Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve templates: %w", err)
	}

	return templates, nil
}

// TestTemplate tests a template with sample data
func (s *TemplateService) TestTemplate(templateID, userID uuid.UUID, testData map[string]interface{}) (*models.TemplateTestResult, error) {
	template, err := s.GetTemplate(templateID, userID)
	if err != nil {
		return nil, err
	}

	// Use the email service to process the template
	emailSvc := NewEmailService()
	htmlContent, textContent, err := emailSvc.processTemplate(*template, testData, template.Subject)
	if err != nil {
		return &models.TemplateTestResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &models.TemplateTestResult{
		Success:     true,
		HTMLContent: htmlContent,
		TextContent: textContent,
		Subject:     template.Subject,
	}, nil
}

// CloneTemplate creates a copy of an existing template
func (s *TemplateService) CloneTemplate(templateID, userID uuid.UUID, name string) (*models.Template, error) {
	original, err := s.GetTemplate(templateID, userID)
	if err != nil {
		return nil, err
	}

	clone := models.Template{
		UserID:      &userID,
		Name:        name,
		Description: original.Description + " (Copy)",
		Subject:     original.Subject,
		HTMLContent: original.HTMLContent,
		TextContent: original.TextContent,
		Variables:   original.Variables,
		Version:     1,
		IsActive:    true,
	}

	if err := s.db.Create(&clone).Error; err != nil {
		return nil, fmt.Errorf("failed to clone template: %w", err)
	}

	return &clone, nil
}

// GetTemplateVersions retrieves version history of a template
func (s *TemplateService) GetTemplateVersions(templateID, userID uuid.UUID) ([]models.TemplateVersion, error) {
	var versions []models.TemplateVersion

	// This is a simplified version - in production, you'd have a separate versions table
	// For now, we just return the current template as version 1
	template, err := s.GetTemplate(templateID, userID)
	if err != nil {
		return nil, err
	}

	versions = append(versions, models.TemplateVersion{
		ID:          template.ID,
		Version:     template.Version,
		Name:        template.Name,
		Subject:     template.Subject,
		HTMLContent: template.HTMLContent,
		TextContent: template.TextContent,
		CreatedAt:   template.UpdatedAt,
		IsActive:    template.IsActive,
	})

	return versions, nil
}

// GetDefaultTemplates retrieves all default/pre-built templates
func (s *TemplateService) GetDefaultTemplates(category string) ([]models.Template, error) {
	var templates []models.Template

	query := database.DB.Where("is_default = ? AND is_public = ? AND is_active = ?", true, true, true)

	// Filter by category if provided
	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Order("category, name").Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch default templates: %w", err)
	}

	return templates, nil
}
