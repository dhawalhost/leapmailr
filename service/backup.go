package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// BackupService handles database backups and disaster recovery (GAP-AV-002)
type BackupService struct {
	logger        *zap.Logger
	backupDir     string
	dbHost        string
	dbPort        string
	dbName        string
	dbUser        string
	retentionDays int
}

// BackupConfig holds backup configuration
type BackupConfig struct {
	BackupDir     string
	DBHost        string
	DBPort        string
	DBName        string
	DBUser        string
	RetentionDays int
}

// NewBackupService creates a new backup service
func NewBackupService(config BackupConfig) *BackupService {
	return &BackupService{
		logger:        zap.L().Named("backup"),
		backupDir:     config.BackupDir,
		dbHost:        config.DBHost,
		dbPort:        config.DBPort,
		dbName:        config.DBName,
		dbUser:        config.DBUser,
		retentionDays: config.RetentionDays,
	}
}

// CreateBackup creates a full database backup
func (s *BackupService) CreateBackup() (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("leapmailr_backup_%s.sql", timestamp)
	backupPath := filepath.Join(s.backupDir, filename)

	// Ensure backup directory exists
	if err := os.MkdirAll(s.backupDir, 0755); err != nil {
		s.logger.Error("Failed to create backup directory",
			zap.String("dir", s.backupDir),
			zap.Error(err),
		)
		return "", err
	}

	s.logger.Info("Starting database backup",
		zap.String("database", s.dbName),
		zap.String("backup_file", backupPath),
	)

	// Create pg_dump command
	cmd := exec.Command(
		"pg_dump",
		"-h", s.dbHost,
		"-p", s.dbPort,
		"-U", s.dbUser,
		"-d", s.dbName,
		"-F", "c", // Custom format (compressed)
		"-f", backupPath,
		"--no-owner",
		"--no-acl",
	)

	// Set PGPASSWORD from environment
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", os.Getenv("DB_PASSWORD")))

	// Execute backup
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("Backup failed",
			zap.String("output", string(output)),
			zap.Error(err),
		)
		return "", fmt.Errorf("backup failed: %w - %s", err, string(output))
	}

	// Verify backup file was created
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		s.logger.Error("Backup file not found", zap.Error(err))
		return "", err
	}

	s.logger.Info("Backup completed successfully",
		zap.String("backup_file", backupPath),
		zap.Int64("size_bytes", fileInfo.Size()),
		zap.String("size_mb", fmt.Sprintf("%.2f", float64(fileInfo.Size())/(1024*1024))),
	)

	return backupPath, nil
}

// VerifyBackup verifies a backup file can be restored
func (s *BackupService) VerifyBackup(backupPath string) error {
	s.logger.Info("Verifying backup", zap.String("backup_file", backupPath))

	// Check if file exists
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		return fmt.Errorf("backup file not found: %w", err)
	}

	// Check file size (should be > 1KB)
	if fileInfo.Size() < 1024 {
		return fmt.Errorf("backup file too small: %d bytes", fileInfo.Size())
	}

	// Verify pg_restore can read the backup
	cmd := exec.Command(
		"pg_restore",
		"--list",
		backupPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("Backup verification failed",
			zap.String("output", string(output)),
			zap.Error(err),
		)
		return fmt.Errorf("backup verification failed: %w", err)
	}

	s.logger.Info("Backup verified successfully",
		zap.String("backup_file", backupPath),
		zap.Int("items", len(output)),
	)

	return nil
}

// RestoreBackup restores database from a backup file
func (s *BackupService) RestoreBackup(backupPath string) error {
	s.logger.Warn("Starting database restore - THIS WILL OVERWRITE EXISTING DATA",
		zap.String("backup_file", backupPath),
		zap.String("database", s.dbName),
	)

	// Verify backup first
	if err := s.VerifyBackup(backupPath); err != nil {
		return fmt.Errorf("backup verification failed: %w", err)
	}

	// Create pg_restore command
	cmd := exec.Command(
		"pg_restore",
		"-h", s.dbHost,
		"-p", s.dbPort,
		"-U", s.dbUser,
		"-d", s.dbName,
		"--clean",
		"--if-exists",
		"--no-owner",
		"--no-acl",
		backupPath,
	)

	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", os.Getenv("DB_PASSWORD")))

	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("Restore failed",
			zap.String("output", string(output)),
			zap.Error(err),
		)
		return fmt.Errorf("restore failed: %w - %s", err, string(output))
	}

	s.logger.Info("Database restored successfully",
		zap.String("backup_file", backupPath),
	)

	return nil
}

// CleanupOldBackups removes backups older than retention period
func (s *BackupService) CleanupOldBackups() error {
	s.logger.Info("Cleaning up old backups",
		zap.Int("retention_days", s.retentionDays),
	)

	cutoffTime := time.Now().AddDate(0, 0, -s.retentionDays)
	deleted := 0

	files, err := filepath.Glob(filepath.Join(s.backupDir, "leapmailr_backup_*.sql"))
	if err != nil {
		return err
	}

	for _, file := range files {
		fileInfo, err := os.Stat(file)
		if err != nil {
			continue
		}

		if fileInfo.ModTime().Before(cutoffTime) {
			s.logger.Info("Removing old backup",
				zap.String("file", file),
				zap.Time("created", fileInfo.ModTime()),
			)
			if err := os.Remove(file); err != nil {
				s.logger.Error("Failed to remove old backup",
					zap.String("file", file),
					zap.Error(err),
				)
			} else {
				deleted++
			}
		}
	}

	s.logger.Info("Cleanup completed",
		zap.Int("deleted_count", deleted),
	)

	return nil
}

// ListBackups lists all available backups
func (s *BackupService) ListBackups() ([]BackupInfo, error) {
	files, err := filepath.Glob(filepath.Join(s.backupDir, "leapmailr_backup_*.sql"))
	if err != nil {
		return nil, err
	}

	backups := make([]BackupInfo, 0, len(files))
	for _, file := range files {
		fileInfo, err := os.Stat(file)
		if err != nil {
			continue
		}

		backups = append(backups, BackupInfo{
			Path:      file,
			Filename:  filepath.Base(file),
			Size:      fileInfo.Size(),
			CreatedAt: fileInfo.ModTime(),
		})
	}

	return backups, nil
}

// BackupInfo contains information about a backup file
type BackupInfo struct {
	Path      string
	Filename  string
	Size      int64
	CreatedAt time.Time
}

// GetBackupStats returns statistics about backups
func (s *BackupService) GetBackupStats() (BackupStats, error) {
	backups, err := s.ListBackups()
	if err != nil {
		return BackupStats{}, err
	}

	var totalSize int64
	var oldest, newest time.Time

	if len(backups) > 0 {
		oldest = backups[0].CreatedAt
		newest = backups[0].CreatedAt

		for _, backup := range backups {
			totalSize += backup.Size
			if backup.CreatedAt.Before(oldest) {
				oldest = backup.CreatedAt
			}
			if backup.CreatedAt.After(newest) {
				newest = backup.CreatedAt
			}
		}
	}

	return BackupStats{
		TotalBackups: len(backups),
		TotalSize:    totalSize,
		OldestBackup: oldest,
		NewestBackup: newest,
	}, nil
}

// BackupStats contains backup statistics
type BackupStats struct {
	TotalBackups int
	TotalSize    int64
	OldestBackup time.Time
	NewestBackup time.Time
}
