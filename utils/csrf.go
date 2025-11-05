package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CSRFToken represents a CSRF token with metadata
type CSRFToken struct {
	Token     string
	UserID    uuid.UUID
	ExpiresAt time.Time
}

// CSRFService manages CSRF tokens (GAP-SEC-014)
type CSRFService struct {
	tokens map[string]*CSRFToken
	mu     sync.RWMutex
	ttl    time.Duration
}

var (
	csrfServiceInstance *CSRFService
	csrfOnce            sync.Once
)

// ErrInvalidCSRFToken indicates an invalid or expired CSRF token
var ErrInvalidCSRFToken = errors.New("invalid or expired CSRF token")

// GetCSRFService returns the singleton CSRF service instance
func GetCSRFService() *CSRFService {
	csrfOnce.Do(func() {
		csrfServiceInstance = &CSRFService{
			tokens: make(map[string]*CSRFToken),
			ttl:    24 * time.Hour, // CSRF tokens valid for 24 hours
		}
		// Start cleanup goroutine
		go csrfServiceInstance.cleanup()
	})
	return csrfServiceInstance
}

// GenerateToken creates a new CSRF token for a user
func (s *CSRFService) GenerateToken(userID uuid.UUID) (string, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Store token with metadata
	s.mu.Lock()
	defer s.mu.Unlock()

	csrfToken := &CSRFToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(s.ttl),
	}

	s.tokens[token] = csrfToken

	return token, nil
}

// ValidateToken validates a CSRF token for a user
func (s *CSRFService) ValidateToken(token string, userID uuid.UUID) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	csrfToken, exists := s.tokens[token]
	if !exists {
		return ErrInvalidCSRFToken
	}

	// Check expiration
	if time.Now().After(csrfToken.ExpiresAt) {
		return ErrInvalidCSRFToken
	}

	// Check user ID match
	if csrfToken.UserID != userID {
		return ErrInvalidCSRFToken
	}

	return nil
}

// DeleteToken removes a CSRF token (e.g., on logout)
func (s *CSRFService) DeleteToken(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, token)
}

// cleanup periodically removes expired tokens
func (s *CSRFService) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for token, csrfToken := range s.tokens {
			if now.After(csrfToken.ExpiresAt) {
				delete(s.tokens, token)
			}
		}
		s.mu.Unlock()
	}
}

// RefreshToken extends the expiration of an existing token
func (s *CSRFService) RefreshToken(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	csrfToken, exists := s.tokens[token]
	if !exists {
		return ErrInvalidCSRFToken
	}

	// Extend expiration
	csrfToken.ExpiresAt = time.Now().Add(s.ttl)
	return nil
}
