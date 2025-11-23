package handlers

import (
	"net/http"

	"github.com/dhawalhost/leapmailr/models"
	"github.com/dhawalhost/leapmailr/service"
	"github.com/dhawalhost/leapmailr/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AddSuppressionHandler adds an email to the suppression list
func AddSuppressionHandler(c *gin.Context) {
	var req models.CreateSuppressionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, _ := c.Get("user")
	userData := user.(*models.User)

	suppressionService := service.NewSuppressionService()
	response, err := suppressionService.AddSuppression(req, userData.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// AddBulkSuppressionsHandler adds multiple emails to the suppression list
func AddBulkSuppressionsHandler(c *gin.Context) {
	var req models.BulkSuppressionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, _ := c.Get("user")
	userData := user.(*models.User)

	suppressionService := service.NewSuppressionService()
	added, err := suppressionService.AddBulkSuppressions(req, userData.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Emails added to suppression list",
		"added":   added,
		"total":   len(req.Emails),
	})
}

// ListSuppressionsHandler lists all suppression entries
func ListSuppressionsHandler(c *gin.Context) {
	user, _ := c.Get("user")
	userData := user.(*models.User)

	// Parse query parameters
	filters := models.SuppressionFilters{
		Reason: c.Query("reason"),
		Source: c.Query("source"),
	}

	// Sanitize search query
	if search := c.Query("search"); search != "" {
		sanitized, err := utils.SanitizeSearchQuery(search)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		filters.Search = sanitized
	}

	// Validate pagination params
	limit, offset, err := utils.ValidatePaginationParams(c.Query("limit"), c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	filters.Limit = limit
	filters.Offset = offset

	suppressionService := service.NewSuppressionService()
	suppressions, total, err := suppressionService.ListSuppressions(userData.ID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"suppressions": suppressions,
		"total":        total,
		"limit":        filters.Limit,
		"offset":       filters.Offset,
	})
}

// DeleteSuppressionHandler removes an email from the suppression list
func DeleteSuppressionHandler(c *gin.Context) {
	suppressionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid suppression ID"})
		return
	}

	user, _ := c.Get("user")
	userData := user.(*models.User)

	suppressionService := service.NewSuppressionService()
	if err := suppressionService.DeleteSuppression(suppressionID, userData.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email removed from suppression list"})
}

// CheckSuppressionHandler checks if an email is suppressed
func CheckSuppressionHandler(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email parameter is required"})
		return
	}

	user, _ := c.Get("user")
	userData := user.(*models.User)

	suppressionService := service.NewSuppressionService()
	isSuppressed, suppression, err := suppressionService.IsEmailSuppressed(email, userData.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if isSuppressed {
		c.JSON(http.StatusOK, gin.H{
			"suppressed": true,
			"suppression": models.SuppressionResponse{
				ID:        suppression.ID,
				Email:     suppression.Email,
				Reason:    suppression.Reason,
				Source:    suppression.Source,
				Metadata:  suppression.Metadata,
				CreatedAt: suppression.CreatedAt,
			},
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"suppressed": false,
		})
	}
}
