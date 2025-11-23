package service

import (
	"errors"
	"fmt"

	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Project query constants
const (
	queryUserIDIsDefault = "user_id = ? AND is_default = ?"
)

// CreateProject creates a new project for a user
func CreateProject(userID uuid.UUID, name, description, color string, isDefault bool) (*models.Project, error) {
	db := database.GetDB()

	// If this is set as default, unset other defaults for this user
	if isDefault {
		if err := db.Model(&models.Project{}).
			Where(queryUserIDIsDefault, userID, true).
			Update("is_default", false).Error; err != nil {
			return nil, fmt.Errorf("failed to unset existing default: %w", err)
		}
	}

	project := &models.Project{
		UserID:      userID,
		Name:        name,
		Description: description,
		Color:       color,
		IsDefault:   isDefault,
	}

	if err := db.Create(project).Error; err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return project, nil
}

// GetUserProjects retrieves all projects for a user
func GetUserProjects(userID uuid.UUID) ([]models.Project, error) {
	db := database.GetDB()
	var projects []models.Project

	if err := db.Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}

	return projects, nil
}

// GetProjectByID retrieves a project by ID and verifies user ownership
func GetProjectByID(projectID, userID uuid.UUID) (*models.Project, error) {
	db := database.GetDB()
	var project models.Project

	if err := db.Where("id = ? AND user_id = ?", projectID, userID).
		First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found")
		}
		return nil, fmt.Errorf("failed to fetch project: %w", err)
	}

	return &project, nil
}

// GetDefaultProject retrieves the default project for a user
func GetDefaultProject(userID uuid.UUID) (*models.Project, error) {
	db := database.GetDB()
	var project models.Project

	if err := db.Where(queryUserIDIsDefault, userID, true).
		First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no default project found")
		}
		return nil, fmt.Errorf("failed to fetch default project: %w", err)
	}

	return &project, nil
}

// UpdateProject updates a project
func UpdateProject(projectID, userID uuid.UUID, name, description, color string, isDefault *bool) (*models.Project, error) {
	db := database.GetDB()

	// Verify ownership
	project, err := GetProjectByID(projectID, userID)
	if err != nil {
		return nil, err
	}

	// If setting as default, unset other defaults
	if isDefault != nil && *isDefault {
		if err := db.Model(&models.Project{}).
			Where("user_id = ? AND id != ? AND is_default = ?", userID, projectID, true).
			Update("is_default", false).Error; err != nil {
			return nil, fmt.Errorf("failed to unset existing default: %w", err)
		}
	}

	// Update fields
	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	if description != "" {
		updates["description"] = description
	}
	if color != "" {
		updates["color"] = color
	}
	if isDefault != nil {
		updates["is_default"] = *isDefault
	}

	if err := db.Model(project).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return project, nil
}

// DeleteProject deletes a project (soft delete)
func DeleteProject(projectID, userID uuid.UUID) error {
	db := database.GetDB()

	// Verify ownership
	project, err := GetProjectByID(projectID, userID)
	if err != nil {
		return err
	}

	// Prevent deletion of default project if it's the only one
	if project.IsDefault {
		var count int64
		if err := db.Model(&models.Project{}).
			Where("user_id = ?", userID).
			Count(&count).Error; err != nil {
			return fmt.Errorf("failed to count projects: %w", err)
		}

		if count <= 1 {
			return errors.New("cannot delete the only project")
		}

		// If deleting default, set another project as default
		var nextProject models.Project
		if err := db.Where("user_id = ? AND id != ?", userID, projectID).
			First(&nextProject).Error; err != nil {
			return fmt.Errorf("failed to find replacement default project: %w", err)
		}

		if err := db.Model(&nextProject).Update("is_default", true).Error; err != nil {
			return fmt.Errorf("failed to set new default project: %w", err)
		}
	}

	// Soft delete
	if err := db.Delete(project).Error; err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// SetDefaultProject sets a project as the default for a user
func SetDefaultProject(projectID, userID uuid.UUID) error {
	db := database.GetDB()

	// Verify ownership
	if _, err := GetProjectByID(projectID, userID); err != nil {
		return err
	}

	// Unset all defaults for this user
	if err := db.Model(&models.Project{}).
		Where(queryUserIDIsDefault, userID, true).
		Update("is_default", false).Error; err != nil {
		return fmt.Errorf("failed to unset existing defaults: %w", err)
	}

	// Set new default
	if err := db.Model(&models.Project{}).
		Where("id = ? AND user_id = ?", projectID, userID).
		Update("is_default", true).Error; err != nil {
		return fmt.Errorf("failed to set default project: %w", err)
	}

	return nil
}

// EnsureDefaultProject ensures a user has a default project, creating one if needed
func EnsureDefaultProject(userID uuid.UUID) (*models.Project, error) {
	// Check if user has a default project
	project, err := GetDefaultProject(userID)
	if err == nil {
		return project, nil
	}

	// No default project exists, create one
	return CreateProject(userID, "Default Project", "Your default project", "#3b82f6", true)
}
