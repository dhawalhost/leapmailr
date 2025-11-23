package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailTrackingService handles email open and click tracking
type EmailTrackingService struct {
	db *gorm.DB
}

// NewEmailTrackingService creates a new email tracking service
func NewEmailTrackingService() *EmailTrackingService {
	return &EmailTrackingService{
		db: database.GetDB(),
	}
}

// CreateTracking creates a tracking record for an email
func (s *EmailTrackingService) CreateTracking(emailLogID uuid.UUID) (*models.EmailTracking, error) {
	// Generate unique tracking pixel ID
	trackingPixelID, err := generateTrackingID()
	if err != nil {
		return nil, err
	}

	tracking := &models.EmailTracking{
		EmailLogID:      emailLogID,
		TrackingPixelID: trackingPixelID,
		OpenCount:       0,
		UniqueOpens:     0,
	}

	if err := s.db.Create(tracking).Error; err != nil {
		return nil, err
	}

	return tracking, nil
}

// InjectTrackingPixel injects a tracking pixel into HTML email content
func (s *EmailTrackingService) InjectTrackingPixel(htmlContent string, trackingPixelID string, baseURL string) string {
	// Create tracking pixel URL
	trackingPixelURL := fmt.Sprintf("%s/api/v1/track/open/%s", baseURL, trackingPixelID)

	// Tracking pixel HTML (1x1 transparent GIF)
	trackingPixel := fmt.Sprintf(`<img src="%s" alt="" width="1" height="1" style="display:block;border:0;outline:none;text-decoration:none;" />`, trackingPixelURL)

	// Try to inject before closing body tag
	if strings.Contains(htmlContent, "</body>") {
		return strings.Replace(htmlContent, "</body>", trackingPixel+"</body>", 1)
	}

	// If no body tag, append to end
	return htmlContent + trackingPixel
}

// InjectLinkTracking replaces all links in HTML with tracked links
// SECURITY: Stores all tracked URLs in database to prevent open redirect vulnerabilities
func (s *EmailTrackingService) InjectLinkTracking(htmlContent string, trackingPixelID string, baseURL string) string {
	// Regular expression to match href attributes
	linkRegex := regexp.MustCompile(`href=["']([^"']+)["']`)

	replacedContent := linkRegex.ReplaceAllStringFunc(htmlContent, func(match string) string {
		// Extract the URL
		urlMatch := regexp.MustCompile(`href=["']([^"']+)["']`).FindStringSubmatch(match)
		if len(urlMatch) < 2 {
			return match
		}

		originalURL := urlMatch[1]

		// Skip tracking pixel and anchor links
		if strings.HasPrefix(originalURL, "#") ||
			strings.Contains(originalURL, "/track/open/") ||
			strings.Contains(originalURL, "/track/click/") {
			return match
		}

		// Generate link ID
		linkID := generateLinkID(originalURL)

		// Store the tracked link in database (SECURITY: Pre-approved URLs only)
		trackedLink := models.TrackedLink{
			TrackingPixelID: trackingPixelID,
			LinkID:          linkID,
			OriginalURL:     originalURL,
		}
		s.db.Create(&trackedLink)

		// Create tracking URL WITHOUT the original URL parameter
		// URL will be retrieved from database using tracking IDs
		trackingURL := fmt.Sprintf("%s/api/v1/track/click/%s/%s",
			baseURL,
			trackingPixelID,
			linkID)

		return fmt.Sprintf(`href="%s"`, trackingURL)
	})

	return replacedContent
}

// RecordOpen records an email open event
func (s *EmailTrackingService) RecordOpen(trackingPixelID string, ipAddress string, userAgent string) error {
	// Find tracking record
	var tracking models.EmailTracking
	if err := s.db.Where("tracking_pixel_id = ?", trackingPixelID).First(&tracking).Error; err != nil {
		return err
	}

	now := time.Now()

	// Check if this is a unique open (by IP + user agent)
	var existingOpen models.EmailOpenEvent
	isUnique := s.db.Where("tracking_id = ? AND ip_address = ? AND user_agent = ?",
		tracking.ID, ipAddress, userAgent).First(&existingOpen).Error == gorm.ErrRecordNotFound

	// Parse user agent for device and email client info
	device := parseDevice(userAgent)
	emailClient := parseEmailClient(userAgent)

	// Create open event
	openEvent := models.EmailOpenEvent{
		TrackingID:  tracking.ID,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Device:      device,
		EmailClient: emailClient,
		OpenedAt:    now,
	}

	if err := s.db.Create(&openEvent).Error; err != nil {
		return err
	}

	// Update tracking record
	updates := map[string]interface{}{
		"open_count":     tracking.OpenCount + 1,
		"last_opened_at": now,
	}

	if tracking.FirstOpenedAt == nil {
		updates["first_opened_at"] = now
	}

	if isUnique {
		updates["unique_opens"] = tracking.UniqueOpens + 1
	}

	if err := s.db.Model(&tracking).Updates(updates).Error; err != nil {
		return err
	}

	// Update EmailLog status if not already opened
	var emailLog models.EmailLog
	if err := s.db.Where("id = ?", tracking.EmailLogID).First(&emailLog).Error; err == nil {
		if emailLog.OpenedAt == nil {
			s.db.Model(&emailLog).Updates(map[string]interface{}{
				"status":    "opened",
				"opened_at": now,
			})
		}
	}

	return nil
}

// RecordClick records a link click event
func (s *EmailTrackingService) RecordClick(trackingPixelID string, linkID string, linkURL string, ipAddress string, userAgent string) error {
	// Find tracking record
	var tracking models.EmailTracking
	if err := s.db.Where("tracking_pixel_id = ?", trackingPixelID).First(&tracking).Error; err != nil {
		return err
	}

	now := time.Now()

	// Parse user agent
	device := parseDevice(userAgent)

	// Create click event
	clickEvent := models.EmailClickEvent{
		TrackingID: tracking.ID,
		LinkID:     linkID,
		LinkURL:    linkURL,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Device:     device,
		ClickedAt:  now,
	}

	if err := s.db.Create(&clickEvent).Error; err != nil {
		return err
	}

	// Update EmailLog status if not already clicked
	var emailLog models.EmailLog
	if err := s.db.Where("id = ?", tracking.EmailLogID).First(&emailLog).Error; err == nil {
		updates := map[string]interface{}{}

		// If never opened, mark as opened too
		if emailLog.OpenedAt == nil {
			updates["opened_at"] = now
		}

		if emailLog.ClickedAt == nil {
			updates["clicked_at"] = now
			updates["status"] = "clicked"
		}

		if len(updates) > 0 {
			s.db.Model(&emailLog).Updates(updates)
		}
	}

	return nil
}

// GetTrackedLinkURL retrieves the original URL for a tracked link from database
// SECURITY: This method ensures URLs are pre-approved (stored at send time), not user-controlled
func (s *EmailTrackingService) GetTrackedLinkURL(trackingPixelID string, linkID string) (string, error) {
	var trackedLink models.TrackedLink
	err := s.db.Where("tracking_pixel_id = ? AND link_id = ?", trackingPixelID, linkID).
		First(&trackedLink).Error

	if err != nil {
		return "", fmt.Errorf("tracked link not found: %w", err)
	}

	return trackedLink.OriginalURL, nil
}

// GetAnalytics retrieves analytics for a specific email
func (s *EmailTrackingService) GetAnalytics(emailLogID uuid.UUID) (*models.EmailTrackingAnalytics, error) {
	var tracking models.EmailTracking
	if err := s.db.Where("email_log_id = ?", emailLogID).
		Preload("OpenEvents").
		Preload("ClickEvents").
		First(&tracking).Error; err != nil {
		return nil, err
	}

	analytics := &models.EmailTrackingAnalytics{
		EmailLogID:      emailLogID,
		TotalOpens:      tracking.OpenCount,
		UniqueOpens:     tracking.UniqueOpens,
		FirstOpenedAt:   tracking.FirstOpenedAt,
		LastOpenedAt:    tracking.LastOpenedAt,
		DeviceBreakdown: make(map[string]int),
		ClientBreakdown: make(map[string]int),
	}

	// Calculate click stats
	totalClicks := len(tracking.ClickEvents)
	uniqueClicks := make(map[string]bool)
	linkStats := make(map[string]*models.LinkStats)

	var firstClickedAt, lastClickedAt *time.Time

	for _, click := range tracking.ClickEvents {
		// Track first and last click
		if firstClickedAt == nil || click.ClickedAt.Before(*firstClickedAt) {
			firstClickedAt = &click.ClickedAt
		}
		if lastClickedAt == nil || click.ClickedAt.After(*lastClickedAt) {
			lastClickedAt = &click.ClickedAt
		}

		// Track unique clicks by IP
		uniqueKey := fmt.Sprintf("%s|%s", click.IPAddress, click.UserAgent)
		uniqueClicks[uniqueKey] = true

		// Aggregate link stats
		if _, exists := linkStats[click.LinkURL]; !exists {
			linkStats[click.LinkURL] = &models.LinkStats{
				LinkURL:      click.LinkURL,
				LinkText:     click.LinkText,
				TotalClicks:  0,
				UniqueClicks: 0,
			}
		}
		linkStats[click.LinkURL].TotalClicks++
	}

	analytics.TotalClicks = totalClicks
	analytics.UniqueClicks = len(uniqueClicks)
	analytics.FirstClickedAt = firstClickedAt
	analytics.LastClickedAt = lastClickedAt

	// Calculate rates
	if analytics.TotalOpens > 0 {
		analytics.OpenRate = 100.0 // Email sent, so 100% if opened
		analytics.ClickRate = float64(analytics.UniqueClicks) / float64(analytics.UniqueOpens) * 100.0
		analytics.ClickToOpenRate = float64(analytics.UniqueClicks) / float64(analytics.UniqueOpens) * 100.0
	}

	// Device and client breakdown
	for _, open := range tracking.OpenEvents {
		if open.Device != "" {
			analytics.DeviceBreakdown[open.Device]++
		}
		if open.EmailClient != "" {
			analytics.ClientBreakdown[open.EmailClient]++
		}
	}

	// Convert link stats map to slice
	for _, stats := range linkStats {
		analytics.TopLinks = append(analytics.TopLinks, *stats)
	}

	return analytics, nil
}

// Helper functions

// generateTrackingID generates a unique tracking ID
func generateTrackingID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// generateLinkID generates a deterministic ID for a link
func generateLinkID(url string) string {
	// Simple hash of URL
	hash := 0
	for _, c := range url {
		hash = 31*hash + int(c)
	}
	return fmt.Sprintf("link_%d", hash)
}

// parseDevice determines device type from user agent
func parseDevice(userAgent string) string {
	ua := strings.ToLower(userAgent)

	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		return "Mobile"
	}

	if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		return "Tablet"
	}

	return "Desktop"
}

// parseEmailClient determines email client from user agent
func parseEmailClient(userAgent string) string {
	ua := strings.ToLower(userAgent)

	clients := map[string]string{
		"gmail":       "Gmail",
		"outlook":     "Outlook",
		"apple mail":  "Apple Mail",
		"thunderbird": "Thunderbird",
		"yahoo":       "Yahoo Mail",
		"aol":         "AOL Mail",
		"protonmail":  "ProtonMail",
	}

	for key, value := range clients {
		if strings.Contains(ua, key) {
			return value
		}
	}

	return "Unknown"
}

// GetTrackingPixel returns a 1x1 transparent GIF
func GetTrackingPixel() []byte {
	// 1x1 transparent GIF (43 bytes)
	gif := []byte{
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00,
		0x01, 0x00, 0x80, 0x00, 0x00, 0xFF, 0xFF, 0xFF,
		0x00, 0x00, 0x00, 0x21, 0xF9, 0x04, 0x01, 0x00,
		0x00, 0x00, 0x00, 0x2C, 0x00, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44,
		0x01, 0x00, 0x3B,
	}
	return gif
}
