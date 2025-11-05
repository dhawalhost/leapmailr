package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dhawalhost/leapmailr/config"
	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/dhawalhost/leapmailr/utils"
)

// Migration script to encrypt existing sensitive data in the database
// This should be run ONCE after deploying the encryption feature
//
// Usage:
//   go run scripts/encrypt-existing-data.go
//
// Prerequisites:
//   1. ENCRYPTION_KEY must be set in .env file
//   2. Database must be accessible
//   3. Create a backup before running: pg_dump -U postgres leapmailr > backup.sql

func main() {
	fmt.Println("=== Sensitive Data Encryption Migration ===")
	fmt.Println("This script will encrypt existing sensitive data in the database.")
	fmt.Println()

	// Confirm before proceeding
	fmt.Print("Have you created a database backup? (yes/no): ")
	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) != "yes" {
		fmt.Println("❌ Please create a backup first using: pg_dump -U postgres leapmailr > backup.sql")
		os.Exit(1)
	}

	// Load configuration
	fmt.Println("Loading configuration...")
	cfg := config.LoadConfig()

	// Check if encryption key is set
	if cfg.EncryptionKey == "" {
		fmt.Println("❌ ERROR: ENCRYPTION_KEY is not set in .env file")
		fmt.Println("Generate one with: openssl rand -base64 32")
		os.Exit(1)
	}

	// Initialize database
	fmt.Println("Connecting to database...")
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	db := database.GetDB()

	// Initialize encryption service
	encryption, err := utils.NewEncryptionService()
	if err != nil {
		log.Fatalf("❌ Failed to initialize encryption service: %v", err)
	}

	fmt.Println("✅ Encryption service initialized")
	fmt.Println()

	// Encrypt User.PrivateKey
	fmt.Println("1️⃣  Encrypting User.PrivateKey fields...")
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatalf("❌ Failed to fetch users: %v", err)
	}

	userEncrypted := 0
	userSkipped := 0
	for _, user := range users {
		if user.PrivateKey == "" {
			userSkipped++
			continue
		}

		// Check if already encrypted (base64 strings are typically longer and contain specific chars)
		if isLikelyEncrypted(user.PrivateKey) {
			userSkipped++
			fmt.Printf("   ⏭️  Skipping user %s (already encrypted)\n", user.Email)
			continue
		}

		// Encrypt the private key
		encryptedKey, err := encryption.Encrypt(user.PrivateKey)
		if err != nil {
			fmt.Printf("   ❌ Failed to encrypt private key for user %s: %v\n", user.Email, err)
			continue
		}

		// Update in database
		if err := db.Model(&user).Update("private_key", encryptedKey).Error; err != nil {
			fmt.Printf("   ❌ Failed to update user %s: %v\n", user.Email, err)
			continue
		}

		userEncrypted++
		fmt.Printf("   ✅ Encrypted private key for user %s\n", user.Email)
	}

	fmt.Printf("   📊 Users: %d encrypted, %d skipped\n", userEncrypted, userSkipped)
	fmt.Println()

	// Encrypt EmailService.Configuration
	fmt.Println("2️⃣  Encrypting EmailService.Configuration fields...")
	var services []models.EmailService
	if err := db.Find(&services).Error; err != nil {
		log.Fatalf("❌ Failed to fetch email services: %v", err)
	}

	serviceEncrypted := 0
	serviceSkipped := 0
	for _, service := range services {
		if service.Configuration == "" {
			serviceSkipped++
			continue
		}

		// Check if already encrypted
		if isLikelyEncrypted(service.Configuration) {
			serviceSkipped++
			fmt.Printf("   ⏭️  Skipping service '%s' (already encrypted)\n", service.Name)
			continue
		}

		// Encrypt the configuration
		encryptedConfig, err := encryption.Encrypt(service.Configuration)
		if err != nil {
			fmt.Printf("   ❌ Failed to encrypt configuration for service '%s': %v\n", service.Name, err)
			continue
		}

		// Update in database
		if err := db.Model(&service).Update("configuration", encryptedConfig).Error; err != nil {
			fmt.Printf("   ❌ Failed to update service '%s': %v\n", service.Name, err)
			continue
		}

		serviceEncrypted++
		fmt.Printf("   ✅ Encrypted configuration for service '%s' (Provider: %s)\n", service.Name, service.Provider)
	}

	fmt.Printf("   📊 Email Services: %d encrypted, %d skipped\n", serviceEncrypted, serviceSkipped)
	fmt.Println()

	// Summary
	fmt.Println("=== Migration Complete ===")
	fmt.Printf("Total encrypted: %d users + %d email services\n", userEncrypted, serviceEncrypted)
	fmt.Printf("Total skipped: %d users + %d email services\n", userSkipped, serviceSkipped)
	fmt.Println()

	// Verification
	fmt.Println("3️⃣  Verifying encryption...")
	var verifyUser models.User
	if err := db.Where("private_key != ''").First(&verifyUser).Error; err == nil {
		decrypted, err := encryption.Decrypt(verifyUser.PrivateKey)
		if err != nil {
			fmt.Printf("   ⚠️  Warning: Could not decrypt private key for user %s\n", verifyUser.Email)
		} else {
			fmt.Printf("   ✅ Verified: User.PrivateKey encryption/decryption works (user: %s)\n", verifyUser.Email)
			fmt.Printf("      Encrypted length: %d, Decrypted length: %d\n", len(verifyUser.PrivateKey), len(decrypted))
		}
	}

	var verifyService models.EmailService
	if err := db.Where("configuration != ''").First(&verifyService).Error; err == nil {
		decrypted, err := encryption.Decrypt(verifyService.Configuration)
		if err != nil {
			fmt.Printf("   ⚠️  Warning: Could not decrypt configuration for service '%s'\n", verifyService.Name)
		} else {
			fmt.Printf("   ✅ Verified: EmailService.Configuration encryption/decryption works (service: %s)\n", verifyService.Name)
			fmt.Printf("      Encrypted length: %d, Decrypted length: %d\n", len(verifyService.Configuration), len(decrypted))
		}
	}

	fmt.Println()
	fmt.Println("✅ Migration completed successfully!")
	fmt.Println()
	fmt.Println("⚠️  IMPORTANT:")
	fmt.Println("   1. Test your application thoroughly")
	fmt.Println("   2. Verify users can login and create email services")
	fmt.Println("   3. Verify email sending still works")
	fmt.Println("   4. Keep the database backup until you're confident")
	fmt.Println("   5. NEVER change or lose the ENCRYPTION_KEY - data will be unrecoverable!")
}

// isLikelyEncrypted checks if a string appears to be already encrypted
// Our encryption produces base64 strings that are typically longer and contain specific patterns
func isLikelyEncrypted(s string) bool {
	// Empty strings are not encrypted
	if s == "" {
		return false
	}

	// Check length - encrypted strings are typically much longer due to base64 encoding
	// Original 64-char key becomes ~120+ chars when encrypted with AES-256-GCM + base64
	// JSON configs also expand significantly
	if len(s) < 40 {
		return false // Too short to be encrypted
	}

	// Check for base64 characteristics
	// Base64 uses A-Z, a-z, 0-9, +, /, = (padding)
	base64Chars := 0
	jsonChars := 0
	for _, c := range s {
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=' {
			base64Chars++
		}
		if c == '{' || c == '}' || c == '"' || c == ':' {
			jsonChars++
		}
	}

	// If mostly base64 characters and very few JSON chars, likely encrypted
	base64Ratio := float64(base64Chars) / float64(len(s))
	jsonRatio := float64(jsonChars) / float64(len(s))

	// Encrypted data should be >95% base64 chars and <5% JSON chars
	return base64Ratio > 0.95 && jsonRatio < 0.05
}
