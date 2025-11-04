package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type CaptchaService struct {
	db *gorm.DB
}

func NewCaptchaService() *CaptchaService {
	return &CaptchaService{
		db: database.GetDB(),
	}
}

func (s *CaptchaService) CreateCaptchaConfig(req models.CreateCaptchaConfigRequest, userID uuid.UUID) (*models.CaptchaConfigResponse, error) {
	// TODO: Encrypt secret key
	newConfig := models.CaptchaConfig{
		UserID:    &userID,
		Provider:  req.Provider,
		SiteKey:   req.SiteKey,
		SecretKey: req.SecretKey, // This should be encrypted
		Domains:   pq.StringArray(req.Domains),
	}
	if req.IsActive != nil {
		newConfig.IsActive = *req.IsActive
	}

	if err := s.db.Create(&newConfig).Error; err != nil {
		return nil, err
	}

	return &models.CaptchaConfigResponse{
		ID:        newConfig.ID,
		Provider:  newConfig.Provider,
		SiteKey:   newConfig.SiteKey,
		Domains:   []string(newConfig.Domains),
		IsActive:  newConfig.IsActive,
		CreatedAt: newConfig.CreatedAt,
		UpdatedAt: newConfig.UpdatedAt,
	}, nil
}

func (s *CaptchaService) ListCaptchaConfigs(userID uuid.UUID) ([]models.CaptchaConfigResponse, error) {
	var configs []models.CaptchaConfig
	if err := s.db.Where("user_id = ?", userID).Find(&configs).Error; err != nil {
		return nil, err
	}

	var response []models.CaptchaConfigResponse
	for _, config := range configs {
		response = append(response, models.CaptchaConfigResponse{
			ID:        config.ID,
			Provider:  config.Provider,
			SiteKey:   config.SiteKey,
			Domains:   []string(config.Domains),
			IsActive:  config.IsActive,
			CreatedAt: config.CreatedAt,
			UpdatedAt: config.UpdatedAt,
		})
	}
	return response, nil
}

func (s *CaptchaService) GetCaptchaConfig(configID, userID uuid.UUID) (*models.CaptchaConfigResponse, error) {
	var config models.CaptchaConfig
	if err := s.db.Where("id = ? AND user_id = ?", configID, userID).First(&config).Error; err != nil {
		return nil, err
	}

	return &models.CaptchaConfigResponse{
		ID:        config.ID,
		Provider:  config.Provider,
		SiteKey:   config.SiteKey,
		Domains:   []string(config.Domains),
		IsActive:  config.IsActive,
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
	}, nil
}

func (s *CaptchaService) UpdateCaptchaConfig(configID, userID uuid.UUID, req models.UpdateCaptchaConfigRequest) (*models.CaptchaConfigResponse, error) {
	var config models.CaptchaConfig
	if err := s.db.Where("id = ? AND user_id = ?", configID, userID).First(&config).Error; err != nil {
		return nil, err
	}

	if req.SiteKey != "" {
		config.SiteKey = req.SiteKey
	}
	if req.SecretKey != "" {
		// TODO: Encrypt secret key
		config.SecretKey = req.SecretKey
	}
	if req.Domains != nil {
		config.Domains = pq.StringArray(req.Domains)
	}
	if req.IsActive != nil {
		config.IsActive = *req.IsActive
	}

	if err := s.db.Save(&config).Error; err != nil {
		return nil, err
	}

	return &models.CaptchaConfigResponse{
		ID:        config.ID,
		Provider:  config.Provider,
		SiteKey:   config.SiteKey,
		Domains:   []string(config.Domains),
		IsActive:  config.IsActive,
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
	}, nil
}

func (s *CaptchaService) DeleteCaptchaConfig(configID, userID uuid.UUID) error {
	return s.db.Where("id = ? AND user_id = ?", configID, userID).Delete(&models.CaptchaConfig{}).Error
}

// Captcha verification endpoints
const (
	reCaptchaVerifyURL = "https://www.google.com/recaptcha/api/siteverify"
	hCaptchaVerifyURL  = "https://api.hcaptcha.com/siteverify"
)

// reCaptchaResponse is the response from Google's reCAPTCHA verification
type reCaptchaResponse struct {
	Success     bool      `json:"success"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}

// hCaptchaResponse is the response from hCaptcha's verification
type hCaptchaResponse struct {
	Success     bool      `json:"success"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	Credit      bool      `json:"credit"`
	ErrorCodes  []string  `json:"error-codes"`
}

// VerifyCaptcha checks the validity of a CAPTCHA token against the provider
func (s *CaptchaService) VerifyCaptcha(token string, configID uuid.UUID) (bool, error) {
	var config models.CaptchaConfig
	if err := s.db.Where("id = ? AND is_active = ?", configID, true).First(&config).Error; err != nil {
		return false, fmt.Errorf("active captcha config not found: %w", err)
	}

	// TODO: Decrypt secret key
	secretKey := config.SecretKey

	switch config.Provider {
	case models.ReCaptchaV2:
		return s.verifyReCaptcha(token, secretKey)
	case models.HCaptcha:
		return s.verifyHCaptcha(token, secretKey)
	default:
		return false, fmt.Errorf("unsupported captcha provider: %s", config.Provider)
	}
}

func (s *CaptchaService) verifyReCaptcha(token, secretKey string) (bool, error) {
	resp, err := http.PostForm(reCaptchaVerifyURL,
		url.Values{"secret": {secretKey}, "response": {token}})
	if err != nil {
		return false, fmt.Errorf("failed to post to reCAPTCHA verify endpoint: %w", err)
	}
	defer resp.Body.Close()

	var result reCaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode reCAPTCHA response: %w", err)
	}

	return result.Success, nil
}

func (s *CaptchaService) verifyHCaptcha(token, secretKey string) (bool, error) {
	resp, err := http.PostForm(hCaptchaVerifyURL,
		url.Values{"secret": {secretKey}, "response": {token}})
	if err != nil {
		return false, fmt.Errorf("failed to post to hCaptcha verify endpoint: %w", err)
	}
	defer resp.Body.Close()

	var result hCaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode hCaptcha response: %w", err)
	}

	return result.Success, nil
}
