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
	"gorm.io/gorm"
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

	// Confirm backup
	if !confirmBackup() {
		os.Exit(1)
	}

	// Setup encryption service
	encryption, db := setupEncryption()

	// Encrypt data
	userStats := encryptUserPrivateKeys(db, encryption)
	serviceStats := encryptEmailServiceConfigs(db, encryption)

	// Print summary
	printSummary(userStats, serviceStats)

	// Verify encryption works
	verifyEncryption(db, encryption)

	printFinalInstructions()
}

// confirmBackup prompts user to confirm database backup
func confirmBackup() bool {
	fmt.Print("Have you created a database backup? (yes/no): ")
	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) != "yes" {
		fmt.Println("❌ Please create a backup first using: pg_dump -U postgres leapmailr > backup.sql")
		return false
	}
	return true
}

// setupEncryption initializes configuration, database, and encryption service
func setupEncryption() (*utils.EncryptionService, *gorm.DB) {
	fmt.Println("Loading configuration...")
	cfg := config.LoadConfig()

	if cfg.EncryptionKey == "" {
		fmt.Println("❌ ERROR: ENCRYPTION_KEY is not set in .env file")
		fmt.Println("Generate one with: openssl rand -base64 32")
		os.Exit(1)
	}

	fmt.Println("Connecting to database...")
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	db := database.GetDB()

	encryption, err := utils.NewEncryptionService()
	if err != nil {
		log.Fatalf("❌ Failed to initialize encryption service: %v", err)
	}

	fmt.Println("✅ Encryption service initialized")
	fmt.Println()

	return encryption, db
}

// encryptionStats holds encryption operation statistics
type encryptionStats struct {
	encrypted int
	skipped   int
}

// encryptUserPrivateKeys encrypts all user private keys
func encryptUserPrivateKeys(db *gorm.DB, encryption *utils.EncryptionService) encryptionStats {
	fmt.Println("1️⃣  Encrypting User.PrivateKey fields...")

	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatalf("❌ Failed to fetch users: %v", err)
	}

	stats := encryptionStats{}
	for _, user := range users {
		if user.PrivateKey == "" {
			stats.skipped++
			continue
		}

		if isLikelyEncrypted(user.PrivateKey) {
			stats.skipped++
			fmt.Printf("   ⏭️  Skipping user %s (already encrypted)\n", user.Email)
			continue
		}

		if encryptAndUpdateUser(db, encryption, &user) {
			stats.encrypted++
			fmt.Printf("   ✅ Encrypted private key for user %s\n", user.Email)
		}
	}

	fmt.Printf("   📊 Users: %d encrypted, %d skipped\n", stats.encrypted, stats.skipped)
	fmt.Println()
	return stats
}

// encryptAndUpdateUser encrypts a single user's private key
func encryptAndUpdateUser(db *gorm.DB, encryption *utils.EncryptionService, user *models.User) bool {
	encryptedKey, err := encryption.Encrypt(user.PrivateKey)
	if err != nil {
		fmt.Printf("   ❌ Failed to encrypt private key for user %s: %v\n", user.Email, err)
		return false
	}

	if err := db.Model(user).Update("private_key", encryptedKey).Error; err != nil {
		fmt.Printf("   ❌ Failed to update user %s: %v\n", user.Email, err)
		return false
	}

	return true
}

// encryptEmailServiceConfigs encrypts all email service configurations
func encryptEmailServiceConfigs(db *gorm.DB, encryption *utils.EncryptionService) encryptionStats {
	fmt.Println("2️⃣  Encrypting EmailService.Configuration fields...")

	var services []models.EmailService
	if err := db.Find(&services).Error; err != nil {
		log.Fatalf("❌ Failed to fetch email services: %v", err)
	}

	stats := encryptionStats{}
	for _, service := range services {
		if service.Configuration == "" {
			stats.skipped++
			continue
		}

		if isLikelyEncrypted(service.Configuration) {
			stats.skipped++
			fmt.Printf("   ⏭️  Skipping service '%s' (already encrypted)\n", service.Name)
			continue
		}

		if encryptAndUpdateService(db, encryption, &service) {
			stats.encrypted++
			fmt.Printf("   ✅ Encrypted configuration for service '%s' (Provider: %s)\n", service.Name, service.Provider)
		}
	}

	fmt.Printf("   📊 Email Services: %d encrypted, %d skipped\n", stats.encrypted, stats.skipped)
	fmt.Println()
	return stats
}

// encryptAndUpdateService encrypts a single email service configuration
func encryptAndUpdateService(db *gorm.DB, encryption *utils.EncryptionService, service *models.EmailService) bool {
	encryptedConfig, err := encryption.Encrypt(service.Configuration)
	if err != nil {
		fmt.Printf("   ❌ Failed to encrypt configuration for service '%s': %v\n", service.Name, err)
		return false
	}

	if err := db.Model(service).Update("configuration", encryptedConfig).Error; err != nil {
		fmt.Printf("   ❌ Failed to update service '%s': %v\n", service.Name, err)
		return false
	}

	return true
}

// printSummary prints migration summary
func printSummary(userStats, serviceStats encryptionStats) {
	fmt.Println("=== Migration Complete ===")
	fmt.Printf("Total encrypted: %d users + %d email services\n", userStats.encrypted, serviceStats.encrypted)
	fmt.Printf("Total skipped: %d users + %d email services\n", userStats.skipped, serviceStats.skipped)
	fmt.Println()
}

// verifyEncryption verifies that encryption/decryption works
func verifyEncryption(db *gorm.DB, encryption *utils.EncryptionService) {
	fmt.Println("3️⃣  Verifying encryption...")

	verifyUserEncryption(db, encryption)
	verifyServiceEncryption(db, encryption)

	fmt.Println()
	fmt.Println("✅ Migration completed successfully!")
	fmt.Println()
}

// verifyUserEncryption verifies user private key encryption
func verifyUserEncryption(db *gorm.DB, encryption *utils.EncryptionService) {
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
}

// verifyServiceEncryption verifies email service configuration encryption
func verifyServiceEncryption(db *gorm.DB, encryption *utils.EncryptionService) {
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
}

// printFinalInstructions prints important post-migration instructions
func printFinalInstructions() {
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
