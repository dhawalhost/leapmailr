package handlers

import (
	"fmt"
	"net/http"

	"github.com/dhawalhost/leapmailr/models"
	"github.com/dhawalhost/leapmailr/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateEmailServiceHandler creates a new email service
func CreateEmailServiceHandler(c *gin.Context) {
	emailServiceManager := service.NewEmailServiceManager()

	var req models.CreateEmailServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from context (set by auth middleware)
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userData := user.(*models.User)
	userUUID := userData.ID

	response, err := emailServiceManager.CreateEmailService(req, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// ListEmailServicesHandler lists all email services for the authenticated user
func ListEmailServicesHandler(c *gin.Context) {
	emailServiceManager := service.NewEmailServiceManager()

	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userData := user.(*models.User)
	userUUID := userData.ID

	// Parse filters from query parameters
	var filters models.EmailServiceFilters
	filters.Provider = c.Query("provider")
	filters.Status = c.Query("status")

	// Parse limit and offset
	if limit := c.Query("limit"); limit != "" {
		var l int
		if _, err := fmt.Sscanf(limit, "%d", &l); err == nil {
			filters.Limit = l
		}
	}
	if offset := c.Query("offset"); offset != "" {
		var o int
		if _, err := fmt.Sscanf(offset, "%d", &o); err == nil {
			filters.Offset = o
		}
	}

	services, err := emailServiceManager.ListEmailServices(userUUID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"services": services})
}

// GetEmailServiceHandler retrieves a single email service by ID
func GetEmailServiceHandler(c *gin.Context) {
	emailServiceManager := service.NewEmailServiceManager()

	serviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userData := user.(*models.User)
	userUUID := userData.ID

	service, err := emailServiceManager.GetEmailService(serviceID, userUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, service)
}

// UpdateEmailServiceHandler updates an existing email service
func UpdateEmailServiceHandler(c *gin.Context) {
	emailServiceManager := service.NewEmailServiceManager()

	serviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	var req models.UpdateEmailServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userData := user.(*models.User)
	userUUID := userData.ID

	response, err := emailServiceManager.UpdateEmailService(serviceID, req, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteEmailServiceHandler deletes an email service
func DeleteEmailServiceHandler(c *gin.Context) {
	emailServiceManager := service.NewEmailServiceManager()

	serviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userData := user.(*models.User)
	userUUID := userData.ID

	if err := emailServiceManager.DeleteEmailService(serviceID, userUUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email service deleted successfully"})
}

// TestEmailServiceHandler tests an email service configuration
func TestEmailServiceHandler(c *gin.Context) {
	emailServiceManager := service.NewEmailServiceManager()

	serviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	var req models.TestEmailServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userData := user.(*models.User)
	userUUID := userData.ID

	response, err := emailServiceManager.TestEmailService(serviceID, userUUID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SetDefaultServiceHandler sets an email service as the default
func SetDefaultServiceHandler(c *gin.Context) {
	emailServiceManager := service.NewEmailServiceManager()

	serviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userData := user.(*models.User)
	userUUID := userData.ID

	if err := emailServiceManager.SetDefault(serviceID, userUUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Default email service set successfully"})
}

// GetSMTPProvidersHandler returns all available SMTP providers
func GetSMTPProvidersHandler(c *gin.Context) {
	providers := models.GetSMTPProviders()
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   providers,
		"count":  len(providers),
	})
}

// GetSMTPProviderHandler returns a specific SMTP provider by ID
func GetSMTPProviderHandler(c *gin.Context) {
	providerID := c.Param("id")
	provider := models.GetSMTPProviderByID(providerID)

	if provider == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "Provider not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   provider,
	})
}

// GetSMTPProviderCategoriesHandler returns provider categories
func GetSMTPProviderCategoriesHandler(c *gin.Context) {
	categories := models.GetSMTPProviderCategories()
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   categories,
	})
}
