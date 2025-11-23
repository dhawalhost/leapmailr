package models

import (
	"time"

	"github.com/google/uuid"
)

// EmailTracking represents tracking data for an email
type EmailTracking struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	EmailLogID      uuid.UUID  `json:"email_log_id" gorm:"type:uuid;not null;index"`
	TrackingPixelID string     `json:"tracking_pixel_id" gorm:"type:varchar(255);uniqueIndex:idx_tracking_pixel_id;not null"` // Unique identifier for tracking pixel
	OpenCount       int        `json:"open_count" gorm:"default:0"`
	LastOpenedAt    *time.Time `json:"last_opened_at"`
	FirstOpenedAt   *time.Time `json:"first_opened_at"`
	UniqueOpens     int        `json:"unique_opens" gorm:"default:0"` // Track unique IP/user agent combinations
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relationships
	EmailLog    EmailLog          `json:"email_log" gorm:"foreignKey:EmailLogID"`
	OpenEvents  []EmailOpenEvent  `json:"open_events,omitempty" gorm:"foreignKey:TrackingID"`
	ClickEvents []EmailClickEvent `json:"click_events,omitempty" gorm:"foreignKey:TrackingID"`
}

// EmailOpenEvent represents a single email open event
type EmailOpenEvent struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TrackingID  uuid.UUID `json:"tracking_id" gorm:"type:uuid;not null;index"`
	IPAddress   string    `json:"ip_address" gorm:"not null"`
	UserAgent   string    `json:"user_agent" gorm:"type:text"`
	Location    string    `json:"location,omitempty"`     // City, Country
	Device      string    `json:"device,omitempty"`       // Desktop, Mobile, Tablet
	EmailClient string    `json:"email_client,omitempty"` // Gmail, Outlook, Apple Mail, etc.
	OpenedAt    time.Time `json:"opened_at" gorm:"not null;index"`

	// Relationships
	Tracking EmailTracking `json:"tracking,omitempty" gorm:"foreignKey:TrackingID"`
}

// EmailClickEvent represents a link click event
type EmailClickEvent struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TrackingID uuid.UUID `json:"tracking_id" gorm:"type:uuid;not null;index"`
	LinkID     string    `json:"link_id" gorm:"not null;index"` // Unique identifier for the link
	LinkURL    string    `json:"link_url" gorm:"type:text;not null"`
	LinkText   string    `json:"link_text,omitempty"`
	IPAddress  string    `json:"ip_address" gorm:"not null"`
	UserAgent  string    `json:"user_agent" gorm:"type:text"`
	Location   string    `json:"location,omitempty"`
	Device     string    `json:"device,omitempty"`
	ClickedAt  time.Time `json:"clicked_at" gorm:"not null;index"`

	// Relationships
	Tracking EmailTracking `json:"tracking,omitempty" gorm:"foreignKey:TrackingID"`
}

// TrackedLink stores pre-approved URLs for secure redirect tracking
// This prevents open redirect vulnerabilities by storing URLs at email send time
type TrackedLink struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TrackingPixelID string    `json:"tracking_pixel_id" gorm:"type:varchar(255);not null;index"`
	LinkID          string    `json:"link_id" gorm:"type:varchar(255);not null;index"`
	OriginalURL     string    `json:"original_url" gorm:"type:text;not null"`
	CreatedAt       time.Time `json:"created_at"`

	// Composite index for fast lookups
	// Index on (tracking_pixel_id, link_id) for O(1) URL lookup
}

// EmailTrackingAnalytics represents aggregated analytics for an email
type EmailTrackingAnalytics struct {
	EmailLogID      uuid.UUID      `json:"email_log_id"`
	TotalOpens      int            `json:"total_opens"`
	UniqueOpens     int            `json:"unique_opens"`
	TotalClicks     int            `json:"total_clicks"`
	UniqueClicks    int            `json:"unique_clicks"`
	FirstOpenedAt   *time.Time     `json:"first_opened_at"`
	LastOpenedAt    *time.Time     `json:"last_opened_at"`
	FirstClickedAt  *time.Time     `json:"first_clicked_at"`
	LastClickedAt   *time.Time     `json:"last_clicked_at"`
	OpenRate        float64        `json:"open_rate"`          // Percentage
	ClickRate       float64        `json:"click_rate"`         // Percentage
	ClickToOpenRate float64        `json:"click_to_open_rate"` // Clicks / Opens
	TopLinks        []LinkStats    `json:"top_links,omitempty"`
	DeviceBreakdown map[string]int `json:"device_breakdown,omitempty"`
	ClientBreakdown map[string]int `json:"email_client_breakdown,omitempty"`
}

// LinkStats represents statistics for a specific link
type LinkStats struct {
	LinkURL      string `json:"link_url"`
	LinkText     string `json:"link_text,omitempty"`
	TotalClicks  int    `json:"total_clicks"`
	UniqueClicks int    `json:"unique_clicks"`
}

// CampaignTrackingAnalytics represents aggregated analytics for a campaign/bulk send
type CampaignTrackingAnalytics struct {
	CampaignID      string             `json:"campaign_id"`
	TotalSent       int                `json:"total_sent"`
	TotalDelivered  int                `json:"total_delivered"`
	TotalOpened     int                `json:"total_opened"`
	TotalClicked    int                `json:"total_clicked"`
	TotalBounced    int                `json:"total_bounced"`
	TotalFailed     int                `json:"total_failed"`
	OpenRate        float64            `json:"open_rate"`
	ClickRate       float64            `json:"click_rate"`
	BounceRate      float64            `json:"bounce_rate"`
	DeliveryRate    float64            `json:"delivery_rate"`
	ClickToOpenRate float64            `json:"click_to_open_rate"`
	TopPerformers   []EmailPerformance `json:"top_performers,omitempty"`
}

// EmailPerformance represents performance metrics for a single email in a campaign
type EmailPerformance struct {
	EmailLogID     uuid.UUID  `json:"email_log_id"`
	ToEmail        string     `json:"to_email"`
	OpenCount      int        `json:"open_count"`
	ClickCount     int        `json:"click_count"`
	FirstOpenedAt  *time.Time `json:"first_opened_at"`
	FirstClickedAt *time.Time `json:"first_clicked_at"`
}
