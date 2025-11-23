package handlers

import (
	"net/http"

	"github.com/dhawalhost/leapmailr/models"
	"github.com/dhawalhost/leapmailr/service"
	"github.com/dhawalhost/leapmailr/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateTemplateHandler creates a new email template
func CreateTemplateHandler(c *gin.Context) {
	templateService := service.NewTemplateService()
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	var req models.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	template, err := templateService.CreateTemplate(req, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create template",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   template,
	})
}

// GetTemplateHandler retrieves a specific template
func GetTemplateHandler(c *gin.Context) {
	templateService := service.NewTemplateService()
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid template ID",
		})
		return
	}

	template, err := templateService.GetTemplate(templateID, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Template not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   template,
	})
}

// UpdateTemplateHandler updates an existing template
func UpdateTemplateHandler(c *gin.Context) {
	templateService := service.NewTemplateService()
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid template ID",
		})
		return
	}

	var req models.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	template, err := templateService.UpdateTemplate(templateID, req, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update template",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   template,
	})
}

// DeleteTemplateHandler deletes a template
func DeleteTemplateHandler(c *gin.Context) {
	templateService := service.NewTemplateService()
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid template ID",
		})
		return
	}

	if err := templateService.DeleteTemplate(templateID, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete template",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Template deleted successfully",
	})
}

// ListTemplatesHandler retrieves all templates for the user
func ListTemplatesHandler(c *gin.Context) {
	templateService := service.NewTemplateService()
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	// Parse query parameters for filters
	var filters models.TemplateFilters

	// Parse project_id if provided
	if projectIDStr := c.Query("project_id"); projectIDStr != "" {
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			return
		}
		filters.ProjectID = &projectID
	}

	// Parse is_active if provided
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		isActive, err := utils.ValidateBooleanParam(isActiveStr, "is_active")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		filters.IsActive = &isActive
	}

	// Sanitize name search
	if name := c.Query("name"); name != "" {
		sanitized, err := utils.SanitizeSearchQuery(name)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		filters.Name = sanitized
	}

	filters.OrderBy = c.Query("order_by")

	// Validate pagination params
	limit, offset, err := utils.ValidatePaginationParams(c.Query("limit"), c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	filters.Limit = limit
	filters.Offset = offset

	templates, err := templateService.ListTemplates(user.ID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve templates",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   templates,
		"count":  len(templates),
	})
}

// TestTemplateHandler tests a template with sample data
func TestTemplateHandler(c *gin.Context) {
	templateService := service.NewTemplateService()
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid template ID",
		})
		return
	}

	var testData map[string]interface{}
	if err := c.ShouldBindJSON(&testData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid test data",
			"details": err.Error(),
		})
		return
	}

	result, err := templateService.TestTemplate(templateID, user.ID, testData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to test template",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result,
	})
}

// CloneTemplateHandler creates a copy of an existing template
func CloneTemplateHandler(c *gin.Context) {
	templateService := service.NewTemplateService()
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid template ID",
		})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Name is required for cloning",
			"details": err.Error(),
		})
		return
	}

	template, err := templateService.CloneTemplate(templateID, user.ID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to clone template",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   template,
	})
}

// GetTemplateVersionsHandler retrieves version history of a template
func GetTemplateVersionsHandler(c *gin.Context) {
	templateService := service.NewTemplateService()
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid template ID",
		})
		return
	}

	versions, err := templateService.GetTemplateVersions(templateID, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve template versions",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   versions,
	})
}

// GetDefaultTemplatesHandler retrieves all default/pre-built templates
func GetDefaultTemplatesHandler(c *gin.Context) {
	templateService := service.NewTemplateService()

	category := c.Query("category") // optional filter by category

	templates, err := templateService.GetDefaultTemplates(category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve default templates",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   templates,
		"count":  len(templates),
	})
}

// GetTemplateCategoriesHandler returns available template categories
func GetTemplateCategoriesHandler(c *gin.Context) {
	categories := []map[string]interface{}{
		{
			"id":          "contact_form",
			"name":        "Contact Forms",
			"description": "Contact form templates for collecting user inquiries",
			"icon":        "📧",
		},
		{
			"id":          "transactional",
			"name":        "Transactional",
			"description": "Order confirmations, password resets, account notifications",
			"icon":        "🔔",
		},
		{
			"id":          "newsletter",
			"name":        "Newsletters",
			"description": "Newsletter and content distribution templates",
			"icon":        "📰",
		},
		{
			"id":          "notification",
			"name":        "Notifications",
			"description": "User notifications, alerts, and team invitations",
			"icon":        "🔔",
		},
		{
			"id":          "custom",
			"name":        "Custom",
			"description": "Your custom-built templates",
			"icon":        "✨",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   categories,
	})
}
