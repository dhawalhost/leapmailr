package handlers

import (
	"net/http"

	"github.com/dhawalhost/leapmailr/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Project error messages
const (
	errUserNotAuthenticated = "User not authenticated"
	errInvalidProjectID     = "Invalid project ID"
)

// CreateProjectRequest represents the request body for creating a project
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color"`
	IsDefault   bool   `json:"is_default"`
}

// UpdateProjectRequest represents the request body for updating a project
type UpdateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	IsDefault   *bool  `json:"is_default"`
}

// CreateProject handles POST /api/v1/projects
func CreateProject(c *gin.Context) {
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errUserNotAuthenticated})
		return
	}

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default color if not provided
	if req.Color == "" {
		req.Color = "#3b82f6"
	}

	project, err := service.CreateProject(user.ID, req.Name, req.Description, req.Color, req.IsDefault)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// GetProjects handles GET /api/v1/projects
func GetProjects(c *gin.Context) {
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errUserNotAuthenticated})
		return
	}

	projects, err := service.GetUserProjects(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// GetProject handles GET /api/v1/projects/:id
func GetProject(c *gin.Context) {
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errUserNotAuthenticated})
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errInvalidProjectID})
		return
	}

	project, err := service.GetProjectByID(projectID, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

// UpdateProject handles PUT /api/v1/projects/:id
func UpdateProject(c *gin.Context) {
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errUserNotAuthenticated})
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errInvalidProjectID})
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := service.UpdateProject(projectID, user.ID, req.Name, req.Description, req.Color, req.IsDefault)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject handles DELETE /api/v1/projects/:id
func DeleteProject(c *gin.Context) {
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errUserNotAuthenticated})
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errInvalidProjectID})
		return
	}

	if err := service.DeleteProject(projectID, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

// SetDefaultProject handles POST /api/v1/projects/:id/default
func SetDefaultProject(c *gin.Context) {
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errUserNotAuthenticated})
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errInvalidProjectID})
		return
	}

	if err := service.SetDefaultProject(projectID, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Default project set successfully"})
}

// GetDefaultProject handles GET /api/v1/projects/default
func GetDefaultProject(c *gin.Context) {
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errUserNotAuthenticated})
		return
	}

	project, err := service.GetDefaultProject(user.ID)
	if err != nil {
		// If no default project, try to ensure one exists
		project, err = service.EnsureDefaultProject(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, project)
}
