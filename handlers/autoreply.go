package handlers

import (
	"net/http"

	"github.com/dhawalhost/leapmailr/models"
	"github.com/dhawalhost/leapmailr/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateAutoReplyHandler creates a new auto-reply configuration
func CreateAutoReplyHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	var req models.AutoReplyConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.UserID = user.ID
	if err := service.CreateAutoReply(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// GetAutoReplyHandler retrieves a single auto-reply configuration
func GetAutoReplyHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	config, err := service.GetAutoReply(id, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "auto-reply not found"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// ListAutoRepliesHandler lists all auto-reply configurations for the user
func ListAutoRepliesHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	configs, err := service.ListAutoReplies(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"autoreplies": configs,
		"total":       len(configs),
	})
}

// UpdateAutoReplyHandler updates an auto-reply configuration
func UpdateAutoReplyHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prevent updating user_id
	delete(updates, "user_id")
	delete(updates, "id")

	if err := service.UpdateAutoReply(id, user.ID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "auto-reply updated successfully"})
}

// DeleteAutoReplyHandler deletes an auto-reply configuration
func DeleteAutoReplyHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	if err := service.DeleteAutoReply(id, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "auto-reply deleted successfully"})
}

// GetAutoReplyLogsHandler retrieves logs for auto-replies
func GetAutoReplyLogsHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	// Check if filtering by specific auto-reply config
	configIDStr := c.Query("config_id")
	limit := 100 // Default limit

	var logs []models.AutoReplyLog
	var err error

	if configIDStr != "" {
		configID, parseErr := uuid.Parse(configIDStr)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config_id"})
			return
		}
		logs, err = service.GetAutoReplyLogsByConfig(configID, user.ID, limit)
	} else {
		logs, err = service.GetAutoReplyLogs(user.ID, limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": len(logs),
	})
}

// TestAutoReplyHandler sends a test auto-reply
func TestAutoReplyHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	var req struct {
		TestEmail string            `json:"test_email" binding:"required,email"`
		Variables map[string]string `json:"variables,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the auto-reply configuration
	config, err := service.GetAutoReply(id, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "auto-reply not found"})
		return
	}

	// Send test auto-reply
	if err := service.SendAutoReply(config, req.TestEmail, req.Variables); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "test auto-reply sent successfully"})
}
