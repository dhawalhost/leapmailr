package handlers

import (
	"net/http"
	"strconv"

	"github.com/dhawalhost/leapmailr/models"
	"github.com/dhawalhost/leapmailr/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GenerateAPIKeyPairHandler creates a new API key pair
func GenerateAPIKeyPairHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	var req models.CreateAPIKeyPairRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	keyPair, err := service.GenerateAPIKeyPair(req, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "API key pair created successfully. Save the private key securely - it won't be shown again!",
		"data":    keyPair,
	})
}

// ListAPIKeyPairsHandler lists all API key pairs for the user
func ListAPIKeyPairsHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	keyPairs, err := service.ListAPIKeyPairs(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"keys":  keyPairs,
		"total": len(keyPairs),
	})
}

// GetAPIKeyPairHandler retrieves a single API key pair
func GetAPIKeyPairHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	keyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key ID"})
		return
	}

	keyPair, err := service.GetAPIKeyPair(keyID, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	c.JSON(http.StatusOK, keyPair)
}

// UpdateAPIKeyPairHandler updates an API key pair
func UpdateAPIKeyPairHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	keyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key ID"})
		return
	}

	var req models.UpdateAPIKeyPairRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.UpdateAPIKeyPair(keyID, user.ID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key updated successfully"})
}

// RevokeAPIKeyPairHandler revokes an API key pair
func RevokeAPIKeyPairHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	keyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key ID"})
		return
	}

	if err := service.RevokeAPIKeyPair(keyID, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key revoked successfully"})
}

// DeleteAPIKeyPairHandler permanently deletes an API key pair
func DeleteAPIKeyPairHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	keyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key ID"})
		return
	}

	if err := service.DeleteAPIKeyPair(keyID, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key deleted successfully"})
}

// RotateAPIKeyPairHandler rotates the private key of an API key pair
func RotateAPIKeyPairHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	keyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key ID"})
		return
	}

	keyPair, err := service.RotateAPIKeyPair(keyID, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Private key rotated successfully. Save the new private key securely!",
		"data":    keyPair,
	})
}

// GetAPIKeyUsageHandler retrieves usage statistics for an API key
func GetAPIKeyUsageHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	keyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key ID"})
		return
	}

	// Get limit from query params (default 100)
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	logs, err := service.GetAPIKeyUsageStats(keyID, user.ID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": len(logs),
	})
}
