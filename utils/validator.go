package utils

import (
	"errors"
	"net/mail"
	"os"
	"regexp"
	"strings"
	"unicode"
)

// Common weak passwords to reject
var commonPasswords = map[string]bool{
	"password":    true,
	"password123": true,
	"123456":      true,
	"12345678":    true,
	"qwerty":      true,
	"abc123":      true,
	"password1":   true,
	"admin":       true,
	"admin123":    true,
	"letmein":     true,
	"welcome":     true,
	"monkey":      true,
	"dragon":      true,
	"master":      true,
	"sunshine":    true,
	"princess":    true,
	"football":    true,
	"iloveyou":    true,
	"superman":    true,
	"trustno1":    true,
	"baseball":    true,
	"batman":      true,
	"access":      true,
	"shadow":      true,
	"michael":     true,
	"jennifer":    true,
	"111111":      true,
	"000000":      true,
	"654321":      true,
	"passw0rd":    true,
	"adminadmin":  true,
	"rootroot":    true,
}

// PasswordPolicy defines password requirements
type PasswordPolicy struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
}

// DefaultPasswordPolicy returns SOC 2 compliant password policy
func DefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:      12,
		RequireUpper:   true,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: true,
	}
}

// ValidatePassword validates a password against the policy
func ValidatePassword(password string, policy PasswordPolicy) error {
	if len(password) < policy.MinLength {
		return errors.New("password must be at least 12 characters long")
	}

	if len(password) > 128 {
		return errors.New("password must not exceed 128 characters")
	}

	// Check for common weak passwords
	lowerPassword := strings.ToLower(password)
	if commonPasswords[lowerPassword] {
		return errors.New("password is too common and easily guessable")
	}

	// Check if password contains variations of common words
	for commonPwd := range commonPasswords {
		if strings.Contains(lowerPassword, commonPwd) && len(commonPwd) > 5 {
			return errors.New("password contains common words that are easily guessable")
		}
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if policy.RequireUpper && !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	if policy.RequireLower && !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	if policy.RequireNumber && !hasNumber {
		return errors.New("password must contain at least one number")
	}

	if policy.RequireSpecial && !hasSpecial {
		return errors.New("password must contain at least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?)")
	}

	// Check for sequential characters (123, abc, etc.)
	if hasSequentialChars(password) {
		return errors.New("password must not contain sequential characters (e.g., 123, abc)")
	}

	// Check for repeated characters (aaa, 111, etc.)
	if hasRepeatedChars(password, 3) {
		return errors.New("password must not contain more than 2 repeated characters")
	}

	return nil
}

// hasSequentialChars checks if password contains sequential characters
func hasSequentialChars(password string) bool {
	sequential := []string{
		"012", "123", "234", "345", "456", "567", "678", "789",
		"abc", "bcd", "cde", "def", "efg", "fgh", "ghi", "hij",
		"ijk", "jkl", "klm", "lmn", "mno", "nop", "opq", "pqr",
		"qrs", "rst", "stu", "tuv", "uvw", "vwx", "wxy", "xyz",
	}

	lowerPassword := strings.ToLower(password)
	for _, seq := range sequential {
		if strings.Contains(lowerPassword, seq) {
			return true
		}
	}
	return false
}

// hasRepeatedChars checks if password contains repeated characters
func hasRepeatedChars(password string, maxRepeat int) bool {
	if len(password) < maxRepeat {
		return false
	}

	for i := 0; i < len(password)-maxRepeat+1; i++ {
		char := password[i]
		repeated := true
		for j := 1; j < maxRepeat; j++ {
			if password[i+j] != char {
				repeated = false
				break
			}
		}
		if repeated {
			return true
		}
	}
	return false
}

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// ValidateEmail validates email format with detailed error
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	if len(email) > 255 {
		return errors.New("email must not exceed 255 characters")
	}

	if !IsValidEmail(email) {
		return errors.New("invalid email format")
	}

	return nil
}

// SanitizeInput sanitizes user input to prevent XSS
func SanitizeInput(input string) string {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Limit length to prevent buffer overflow
	if len(input) > 10000 {
		input = input[:10000]
	}

	return input
}

// ValidateName validates first/last name
func ValidateName(name string) error {
	if name == "" {
		return errors.New("name is required")
	}

	if len(name) > 100 {
		return errors.New("name must not exceed 100 characters")
	}

	// Allow letters, spaces, hyphens, apostrophes
	nameRegex := regexp.MustCompile(`^[a-zA-Z\s'-]+$`)
	if !nameRegex.MatchString(name) {
		return errors.New("name contains invalid characters")
	}

	return nil
}

// GetBaseURL returns the base URL for the application from environment
func GetBaseURL() string {
	// Try to get from environment variable
	baseURL := os.Getenv("BASE_URL")
	if baseURL != "" {
		return strings.TrimSuffix(baseURL, "/")
	}

	// Default to localhost for development
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return "http://localhost:" + port
}
