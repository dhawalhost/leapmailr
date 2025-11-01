package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/google/uuid"
)

// CreateContact creates a new contact or updates existing one
func CreateContact(req models.CreateContactRequest, userID uuid.UUID) (*models.ContactResponse, error) {
	if req.Email == "" {
		return nil, errors.New("email is required")
	}

	// Normalize email
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Check if contact already exists
	var existingContact models.Contact
	err := database.DB.Where("user_id = ? AND email = ?", userID, email).First(&existingContact).Error

	if err == nil {
		// Contact exists, update submission count and last submitted time
		existingContact.SubmissionCount++
		existingContact.LastSubmittedAt = time.Now()

		// Update other fields if provided
		if req.Name != "" {
			existingContact.Name = req.Name
		}
		if req.Phone != "" {
			existingContact.Phone = req.Phone
		}
		if req.Company != "" {
			existingContact.Company = req.Company
		}

		// Merge metadata
		if len(req.Metadata) > 0 {
			var existingMeta map[string]string
			if existingContact.Metadata != "" {
				json.Unmarshal([]byte(existingContact.Metadata), &existingMeta)
			} else {
				existingMeta = make(map[string]string)
			}
			for k, v := range req.Metadata {
				existingMeta[k] = v
			}
			metaBytes, _ := json.Marshal(existingMeta)
			existingContact.Metadata = string(metaBytes)
		}

		// Merge tags
		if len(req.Tags) > 0 {
			var existingTags []string
			if existingContact.Tags != "" {
				json.Unmarshal([]byte(existingContact.Tags), &existingTags)
			}
			tagMap := make(map[string]bool)
			for _, tag := range existingTags {
				tagMap[tag] = true
			}
			for _, tag := range req.Tags {
				tagMap[tag] = true
			}
			allTags := make([]string, 0, len(tagMap))
			for tag := range tagMap {
				allTags = append(allTags, tag)
			}
			tagsBytes, _ := json.Marshal(allTags)
			existingContact.Tags = string(tagsBytes)
		}

		if err := database.DB.Save(&existingContact).Error; err != nil {
			return nil, fmt.Errorf("failed to update contact: %w", err)
		}

		return toContactResponse(&existingContact), nil
	}

	// Create new contact
	source := req.Source
	if source == "" {
		source = "manual"
	}

	var metadataJSON, tagsJSON string
	if len(req.Metadata) > 0 {
		metaBytes, _ := json.Marshal(req.Metadata)
		metadataJSON = string(metaBytes)
	}
	if len(req.Tags) > 0 {
		tagsBytes, _ := json.Marshal(req.Tags)
		tagsJSON = string(tagsBytes)
	}

	contact := models.Contact{
		UserID:          userID,
		Email:           email,
		Name:            req.Name,
		Phone:           req.Phone,
		Company:         req.Company,
		Source:          source,
		Metadata:        metadataJSON,
		Tags:            tagsJSON,
		IsSubscribed:    true,
		SubmissionCount: 1,
		LastSubmittedAt: time.Now(),
	}

	if err := database.DB.Create(&contact).Error; err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	return toContactResponse(&contact), nil
}

// GetContact retrieves a single contact by ID
func GetContact(contactID uuid.UUID, userID uuid.UUID) (*models.ContactResponse, error) {
	var contact models.Contact
	err := database.DB.Where("id = ? AND user_id = ?", contactID, userID).First(&contact).Error
	if err != nil {
		return nil, err
	}
	return toContactResponse(&contact), nil
}

// ListContacts retrieves all contacts for a user with optional filters
func ListContacts(userID uuid.UUID, search string, tags []string, subscribed *bool, limit, offset int) ([]models.ContactResponse, int64, error) {
	query := database.DB.Where("user_id = ?", userID)

	// Apply search filter
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("email ILIKE ? OR name ILIKE ? OR company ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	// Apply subscription filter
	if subscribed != nil {
		query = query.Where("is_subscribed = ?", *subscribed)
	}

	// Apply tags filter
	if len(tags) > 0 {
		for _, tag := range tags {
			query = query.Where("tags::jsonb @> ?", fmt.Sprintf(`["%s"]`, tag))
		}
	}

	// Get total count
	var total int64
	if err := query.Model(&models.Contact{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	query = query.Order("last_submitted_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var contacts []models.Contact
	if err := query.Find(&contacts).Error; err != nil {
		return nil, 0, err
	}

	responses := make([]models.ContactResponse, len(contacts))
	for i, c := range contacts {
		responses[i] = *toContactResponse(&c)
	}

	return responses, total, nil
}

// UpdateContact updates an existing contact
func UpdateContact(contactID uuid.UUID, userID uuid.UUID, req models.UpdateContactRequest) error {
	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Company != "" {
		updates["company"] = req.Company
	}
	if len(req.Metadata) > 0 {
		metaBytes, _ := json.Marshal(req.Metadata)
		updates["metadata"] = string(metaBytes)
	}
	if len(req.Tags) > 0 {
		tagsBytes, _ := json.Marshal(req.Tags)
		updates["tags"] = string(tagsBytes)
	}
	if req.IsSubscribed != nil {
		updates["is_subscribed"] = *req.IsSubscribed
	}

	result := database.DB.Model(&models.Contact{}).
		Where("id = ? AND user_id = ?", contactID, userID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("contact not found")
	}

	return nil
}

// DeleteContact deletes a contact
func DeleteContact(contactID uuid.UUID, userID uuid.UUID) error {
	result := database.DB.Where("id = ? AND user_id = ?", contactID, userID).Delete(&models.Contact{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("contact not found")
	}
	return nil
}

// ImportContacts bulk imports contacts
func ImportContacts(req models.ImportContactsRequest, userID uuid.UUID) (int, int, error) {
	imported := 0
	updated := 0

	for _, contactReq := range req.Contacts {
		if req.Source != "" {
			contactReq.Source = req.Source
		}

		// Check if contact exists
		email := strings.ToLower(strings.TrimSpace(contactReq.Email))
		var existingContact models.Contact
		err := database.DB.Where("user_id = ? AND email = ?", userID, email).First(&existingContact).Error

		if err == nil {
			// Update existing
			updated++
		} else {
			// Create new
			imported++
		}

		_, err = CreateContact(contactReq, userID)
		if err != nil {
			// Continue with other contacts even if one fails
			continue
		}
	}

	return imported, updated, nil
}

// GetContactStats retrieves statistics about contacts
func GetContactStats(userID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total contacts
	var total int64
	database.DB.Model(&models.Contact{}).Where("user_id = ?", userID).Count(&total)
	stats["total"] = total

	// Subscribed contacts
	var subscribed int64
	database.DB.Model(&models.Contact{}).Where("user_id = ? AND is_subscribed = ?", userID, true).Count(&subscribed)
	stats["subscribed"] = subscribed

	// Unsubscribed contacts
	stats["unsubscribed"] = total - subscribed

	// New this month
	startOfMonth := time.Now().AddDate(0, 0, -30)
	var thisMonth int64
	database.DB.Model(&models.Contact{}).Where("user_id = ? AND created_at >= ?", userID, startOfMonth).Count(&thisMonth)
	stats["new_this_month"] = thisMonth

	return stats, nil
}

// ExportContacts exports contacts to CSV format
func ExportContacts(userID uuid.UUID) ([][]string, error) {
	var contacts []models.Contact
	if err := database.DB.Where("user_id = ?", userID).Order("email").Find(&contacts).Error; err != nil {
		return nil, err
	}

	// CSV header
	csv := [][]string{
		{"Email", "Name", "Phone", "Company", "Source", "Subscribed", "Submission Count", "Last Submitted", "Created"},
	}

	for _, contact := range contacts {
		row := []string{
			contact.Email,
			contact.Name,
			contact.Phone,
			contact.Company,
			contact.Source,
			fmt.Sprintf("%t", contact.IsSubscribed),
			fmt.Sprintf("%d", contact.SubmissionCount),
			contact.LastSubmittedAt.Format("2006-01-02 15:04:05"),
			contact.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		csv = append(csv, row)
	}

	return csv, nil
}

// Helper functions

func toContactResponse(c *models.Contact) *models.ContactResponse {
	var metadata map[string]string
	var tags []string

	if c.Metadata != "" {
		json.Unmarshal([]byte(c.Metadata), &metadata)
	}
	if c.Tags != "" {
		json.Unmarshal([]byte(c.Tags), &tags)
	}

	return &models.ContactResponse{
		ID:              c.ID,
		Email:           c.Email,
		Name:            c.Name,
		Phone:           c.Phone,
		Company:         c.Company,
		Source:          c.Source,
		Metadata:        metadata,
		Tags:            tags,
		IsSubscribed:    c.IsSubscribed,
		SubmissionCount: c.SubmissionCount,
		LastSubmittedAt: c.LastSubmittedAt,
		CreatedAt:       c.CreatedAt,
	}
}
