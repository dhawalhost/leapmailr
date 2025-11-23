package handlers

import (
	"encoding/csv"
	"net/http"

	"github.com/dhawalhost/leapmailr/models"
	"github.com/dhawalhost/leapmailr/service"
	"github.com/dhawalhost/leapmailr/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateContactHandler creates a new contact
func CreateContactHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	var req models.CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contact, err := service.CreateContact(req, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, contact)
}

// GetContactHandler retrieves a single contact
func GetContactHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	contactID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contact ID"})
		return
	}

	contact, err := service.GetContact(contactID, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "contact not found"})
		return
	}

	c.JSON(http.StatusOK, contact)
}

// ListContactsHandler lists all contacts for the user
func ListContactsHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	// Sanitize search query
	search := c.Query("search")
	if search != "" {
		sanitized, err := utils.SanitizeSearchQuery(search)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		search = sanitized
	}

	// Validate and parse tags
	tags, err := utils.ValidateTagsList(c.Query("tags"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse subscribed boolean
	var subscribed *bool
	if subscribedStr := c.Query("subscribed"); subscribedStr != "" {
		val, err := utils.ValidateBooleanParam(subscribedStr, "subscribed")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		subscribed = &val
	}

	// Validate pagination params
	limit, offset, err := utils.ValidatePaginationParams(c.Query("limit"), c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contacts, total, err := service.ListContacts(user.ID, search, tags, subscribed, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contacts": contacts,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// UpdateContactHandler updates a contact
func UpdateContactHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	contactID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contact ID"})
		return
	}

	var req models.UpdateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.UpdateContact(contactID, user.ID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "contact updated successfully"})
}

// DeleteContactHandler deletes a contact
func DeleteContactHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	contactID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contact ID"})
		return
	}

	if err := service.DeleteContact(contactID, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "contact deleted successfully"})
}

// ImportContactsHandler bulk imports contacts
func ImportContactsHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	var req models.ImportContactsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imported, updated, err := service.ImportContacts(req, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "contacts imported successfully",
		"imported": imported,
		"updated":  updated,
		"total":    imported + updated,
	})
}

// GetContactStatsHandler retrieves contact statistics
func GetContactStatsHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	stats, err := service.GetContactStats(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ExportContactsHandler exports contacts as CSV
func ExportContactsHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	csvData, err := service.ExportContacts(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set headers for CSV download
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=contacts.csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	for _, row := range csvData {
		if err := writer.Write(row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write CSV"})
			return
		}
	}
}
