package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetEmailServiceConfigHandler returns the configuration of an email service (for debugging)
func GetEmailServiceConfigHandler(c *gin.Context) {
	user, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	serviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	var service models.EmailService
	if err := database.GetDB().Where("id = ? AND user_id = ?", serviceID, user.ID).First(&service).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Email service not found"})
		return
	}

	// Parse configuration to show (without passwords)
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(service.Configuration), &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse configuration"})
		return
	}

	// Mask sensitive fields
	if _, ok := config["password"]; ok {
		config["password"] = "***masked***"
	}
	if _, ok := config["api_key"]; ok {
		config["api_key"] = "***masked***"
	}
	if _, ok := config["secret_key"]; ok {
		config["secret_key"] = "***masked***"
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"id":            service.ID,
			"name":          service.Name,
			"provider":      service.Provider,
			"configuration": config,
			"is_default":    service.IsDefault,
			"status":        service.Status,
			"last_error":    service.LastError,
		},
	})
}
