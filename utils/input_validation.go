package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Compile regex patterns once at package level for performance
var tagValidationRegex = regexp.MustCompile(`^[a-zA-Z0-9 _-]+$`)

// ValidatePaginationParams validates and sanitizes limit/offset parameters
// Returns sanitized values with defaults and bounds checking
func ValidatePaginationParams(limitStr, offsetStr string) (limit, offset int, err error) {
	// Set defaults
	limit = 50
	offset = 0

	// Validate limit
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil {
			return 0, 0, errors.New("invalid limit: must be a number")
		}
		if l < 1 {
			return 0, 0, errors.New("invalid limit: must be at least 1")
		}
		if l > 1000 {
			return 0, 0, errors.New("invalid limit: maximum is 1000")
		}
		limit = l
	}

	// Validate offset
	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil {
			return 0, 0, errors.New("invalid offset: must be a number")
		}
		if o < 0 {
			return 0, 0, errors.New("invalid offset: must be non-negative")
		}
		offset = o
	}

	return limit, offset, nil
}

// ValidateEnum validates a string value against a list of allowed values
func ValidateEnum(value string, allowedValues []string, fieldName string) error {
	if value == "" {
		return nil // Empty is OK for optional fields
	}

	// Convert to lowercase for case-insensitive comparison
	valueLower := strings.ToLower(value)

	for _, allowed := range allowedValues {
		if valueLower == strings.ToLower(allowed) {
			return nil
		}
	}

	return fmt.Errorf("invalid %s: must be one of %v", fieldName, allowedValues)
}

// SanitizeSearchQuery sanitizes search input for SQL LIKE queries
// Prevents SQL injection and removes dangerous characters
func SanitizeSearchQuery(query string) (string, error) {
	// Max length check
	if len(query) > 255 {
		return "", errors.New("search query too long (max 255 characters)")
	}

	if query == "" {
		return "", nil
	}

	// Remove control characters and null bytes
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || r == 0 {
			return -1
		}
		return r
	}, query)

	// Escape SQL wildcards for PostgreSQL
	cleaned = strings.ReplaceAll(cleaned, "\\", "\\\\")
	cleaned = strings.ReplaceAll(cleaned, "%", "\\%")
	cleaned = strings.ReplaceAll(cleaned, "_", "\\_")

	// Trim whitespace
	cleaned = strings.TrimSpace(cleaned)

	return cleaned, nil
}

// ValidateAlphanumeric validates string contains only alphanumeric + allowed chars
func ValidateAlphanumeric(value string, maxLength int, allowedChars string, fieldName string) error {
	if len(value) == 0 {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	if len(value) > maxLength {
		return fmt.Errorf("%s exceeds maximum length of %d", fieldName, maxLength)
	}

	// Build allowed character set
	allowed := allowedChars + "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	for _, char := range value {
		if !strings.ContainsRune(allowed, char) {
			return fmt.Errorf("%s contains invalid characters", fieldName)
		}
	}

	return nil
}

// ValidateTrackingID validates tracking pixel/link IDs
// Should be alphanumeric with hyphens and underscores, reasonable length
func ValidateTrackingID(id string) error {
	if len(id) == 0 {
		return errors.New("tracking ID cannot be empty")
	}

	if len(id) > 255 {
		return errors.New("tracking ID too long (max 255 characters)")
	}

	// Check for valid characters (alphanumeric, hyphens, underscores)
	matched, err := regexp.MatchString("^[a-zA-Z0-9_-]+$", id)
	if err != nil {
		return fmt.Errorf("regex error: %v", err)
	}

	if !matched {
		return errors.New("tracking ID contains invalid characters (allowed: a-z, A-Z, 0-9, _, -)")
	}

	return nil
}

// ValidateProvider validates SMTP provider name against whitelist
func ValidateProvider(provider string) error {
	if provider == "" {
		return nil // Empty is OK for optional filtering
	}

	allowedProviders := []string{
		"gmail", "outlook", "yahoo", "zoho", "godaddy", "namecheap",
		"sendgrid", "mailgun", "postmark", "ses", "sparkpost", "brevo",
		"mailjet", "elastic", "icloud", "fastmail", "protonmail", "custom",
	}

	return ValidateEnum(provider, allowedProviders, "provider")
}

// ValidateStatus validates email/service status against allowed values
func ValidateStatus(status string) error {
	if status == "" {
		return nil // Empty is OK for optional filtering
	}

	allowedStatuses := []string{
		"active", "inactive", "pending", "disabled", "error",
		"sent", "delivered", "bounced", "failed", "queued",
		"opened", "clicked", "unsubscribed", "complained",
	}

	return ValidateEnum(status, allowedStatuses, "status")
}

// ValidateTagsList validates comma-separated tags
// Returns cleaned array of tags
func ValidateTagsList(tagsStr string) ([]string, error) {
	if tagsStr == "" {
		return []string{}, nil
	}

	tags := strings.Split(tagsStr, ",")

	if len(tags) > 50 {
		return nil, errors.New("too many tags (max 50)")
	}

	validated := make([]string, 0, len(tags))
	seen := make(map[string]bool) // Deduplicate

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)

		if len(tag) == 0 {
			continue // Skip empty tags
		}

		if len(tag) > 50 {
			return nil, fmt.Errorf("tag '%s' exceeds maximum length of 50", tag)
		}

		// Allow alphanumeric, spaces, hyphens, underscores
		if !tagValidationRegex.MatchString(tag) {
			return nil, fmt.Errorf("tag '%s' contains invalid characters (allowed: a-z, A-Z, 0-9, space, _, -)", tag)
		}

		// Deduplicate (case-insensitive)
		tagLower := strings.ToLower(tag)
		if !seen[tagLower] {
			seen[tagLower] = true
			validated = append(validated, tag)
		}
	}

	return validated, nil
}

// ValidateBooleanParam validates boolean query parameters
func ValidateBooleanParam(value string, fieldName string) (bool, error) {
	if value == "" {
		return false, nil // Empty means not set
	}

	valueLower := strings.ToLower(value)

	switch valueLower {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid %s: must be true/false, 1/0, or yes/no", fieldName)
	}
}

// ValidateStringLength validates string length constraints
func ValidateStringLength(value string, minLength, maxLength int, fieldName string) error {
	length := len(value)

	if minLength > 0 && length < minLength {
		return fmt.Errorf("%s must be at least %d characters", fieldName, minLength)
	}

	if maxLength > 0 && length > maxLength {
		return fmt.Errorf("%s must be at most %d characters", fieldName, maxLength)
	}

	return nil
}

// SanitizeFilename sanitizes filenames to prevent path traversal
func SanitizeFilename(filename string) (string, error) {
	if filename == "" {
		return "", errors.New("filename cannot be empty")
	}

	// Remove path separators
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	filename = strings.ReplaceAll(filename, "..", "")

	// Remove control characters
	filename = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, filename)

	// Max length
	if len(filename) > 255 {
		return "", errors.New("filename too long (max 255 characters)")
	}

	if filename == "" {
		return "", errors.New("filename contains only invalid characters")
	}

	return filename, nil
}
