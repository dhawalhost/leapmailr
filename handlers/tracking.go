package handlers

import (
	"encoding/base64"
	"net/http"

	"github.com/dhawalhost/leapmailr/service"
	"github.com/dhawalhost/leapmailr/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TrackOpenHandler handles email open tracking via pixel
func TrackOpenHandler(c *gin.Context) {
	trackingPixelID := c.Param("pixel_id")

	// Validate tracking ID
	if err := utils.ValidateTrackingID(trackingPixelID); err != nil {
		// Return transparent pixel anyway (don't reveal tracking failure)
		c.Data(http.StatusOK, "image/gif", service.GetTrackingPixel())
		return
	}

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Record open event (async to not slow down pixel load)
	go func() {
		trackingService := service.NewEmailTrackingService()
		trackingService.RecordOpen(trackingPixelID, ipAddress, userAgent)
	}()

	// Return 1x1 transparent GIF
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Data(http.StatusOK, "image/gif", service.GetTrackingPixel())
}

// TrackClickHandler handles link click tracking and redirect
func TrackClickHandler(c *gin.Context) {
	trackingPixelID := c.Param("pixel_id")
	linkID := c.Param("link_id")
	encodedURL := c.Query("url")

	// Validate tracking IDs
	if err := utils.ValidateTrackingID(trackingPixelID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tracking pixel ID"})
		return
	}
	if err := utils.ValidateTrackingID(linkID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid link ID"})
		return
	}

	// Decode the original URL
	urlBytes, err := base64.URLEncoding.DecodeString(encodedURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tracking URL"})
		return
	}

	originalURL := string(urlBytes)

	// Validate the URL before redirecting (SECURITY: Prevent open redirect attacks)
	if err := utils.ValidateRedirectURL(originalURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid redirect URL",
			"details": err.Error(),
		})
		return
	}

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Record click event (async)
	go func() {
		trackingService := service.NewEmailTrackingService()
		trackingService.RecordClick(trackingPixelID, linkID, originalURL, ipAddress, userAgent)
	}()

	// Redirect to validated URL
	c.Redirect(http.StatusFound, originalURL)
}

// GetEmailAnalyticsHandler returns analytics for a specific email
func GetEmailAnalyticsHandler(c *gin.Context) {
	_, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	emailIDStr := c.Param("email_id")
	emailID, err := uuid.Parse(emailIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email ID"})
		return
	}

	// Verify email belongs to user
	// TODO: Add authorization check

	trackingService := service.NewEmailTrackingService()
	analytics, err := trackingService.GetAnalytics(emailID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Analytics not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   analytics,
	})
}

// GetCampaignAnalyticsHandler returns aggregated analytics for a campaign
func GetCampaignAnalyticsHandler(c *gin.Context) {
	_, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	campaignID := c.Param("campaign_id")
	if campaignID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Campaign ID required"})
		return
	}

	// TODO: Implement campaign analytics aggregation
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Campaign analytics coming soon",
		"data": gin.H{
			"campaign_id": campaignID,
		},
	})
}

// GetEmailTrackingEventsHandler returns detailed tracking events for an email
func GetEmailTrackingEventsHandler(c *gin.Context) {
	_, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	emailIDStr := c.Param("email_id")
	emailID, err := uuid.Parse(emailIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email ID"})
		return
	}

	// TODO: Implement detailed events retrieval
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"email_id":     emailID,
			"open_events":  []interface{}{},
			"click_events": []interface{}{},
		},
	})
}
