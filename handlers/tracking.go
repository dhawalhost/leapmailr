package handlers

import (
	"net/http"

	"github.com/dhawalhost/leapmailr/service"
	"github.com/dhawalhost/leapmailr/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Tracking error messages
const (
	errAuthRequired = "Authentication required"
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
		_ = trackingService.RecordOpen(trackingPixelID, ipAddress, userAgent)
	}()

	// Return 1x1 transparent GIF
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Data(http.StatusOK, "image/gif", service.GetTrackingPixel())
}

// TrackClickHandler handles link click tracking and redirect
// SECURITY: No longer accepts URLs from user input - retrieves from database
func TrackClickHandler(c *gin.Context) {
	trackingPixelID := c.Param("pixel_id")
	linkID := c.Param("link_id")

	// Validate tracking IDs
	if err := utils.ValidateTrackingID(trackingPixelID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tracking pixel ID"})
		return
	}
	if err := utils.ValidateTrackingID(linkID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid link ID"})
		return
	}

	// SECURITY: Retrieve pre-approved URL from database (stored at send time)
	// This prevents open redirect attacks - URL is NOT from user input
	trackingService := service.NewEmailTrackingService()
	originalURL, err := trackingService.GetTrackedLinkURL(trackingPixelID, linkID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Tracked link not found",
		})
		return
	}

	// Additional validation as defense-in-depth
	// (URLs should already be validated at send time, but validate again)
	if err := utils.ValidateRedirectURL(originalURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid redirect URL",
			"details": "The stored URL failed validation",
		})
		return
	}

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Record click event (async)
	go func() {
		trackingService := service.NewEmailTrackingService()
		_ = trackingService.RecordClick(trackingPixelID, linkID, originalURL, ipAddress, userAgent)
	}()

	// Redirect to pre-approved URL from database (NOT from user input)
	c.Redirect(http.StatusFound, originalURL)
}

// GetEmailAnalyticsHandler returns analytics for a specific email
func GetEmailAnalyticsHandler(c *gin.Context) {
	_, err := GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errAuthRequired})
		return
	}

	emailIDStr := c.Param("email_id")
	emailID, err := uuid.Parse(emailIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email ID"})
		return
	}

	// Note: Authorization check should verify email belongs to the authenticated user
	// This should be implemented alongside the user authentication middleware

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
		c.JSON(http.StatusUnauthorized, gin.H{"error": errAuthRequired})
		return
	}

	campaignID := c.Param("campaign_id")
	if campaignID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Campaign ID required"})
		return
	}

	// Note: Campaign analytics aggregation requires a campaigns table and relationship mapping
	// Future enhancement: Aggregate analytics across all emails in a campaign
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": errAuthRequired})
		return
	}

	emailIDStr := c.Param("email_id")
	emailID, err := uuid.Parse(emailIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email ID"})
		return
	}

	// Note: Detailed events retrieval needs to query email_open_events and email_click_events tables
	// Future enhancement: Return full event history with timestamps, IPs, user agents, etc.
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"email_id":     emailID,
			"open_events":  []interface{}{},
			"click_events": []interface{}{},
		},
	})
}
