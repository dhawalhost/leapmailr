package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	Outlogger = "outlogger"
	ApiLogger = "apilogger"
)

// LogConfig holds logging configuration (GAP-SEC-009)
type LogConfig struct {
	Level         string // debug, info, warn, error
	Format        string // json, console
	OutputPaths   []string
	EnableFile    bool
	FilePath      string
	MaxSize       int // MB
	MaxBackups    int
	MaxAge        int // days
	Compress      bool
	EnableConsole bool
}

// DefaultLogConfig returns default logging configuration
func DefaultLogConfig() LogConfig {
	return LogConfig{
		Level:         "info",
		Format:        "json",
		OutputPaths:   []string{"stdout"},
		EnableFile:    true,
		FilePath:      "logs/application.log",
		MaxSize:       100, // 100MB
		MaxBackups:    10,
		MaxAge:        30,
		Compress:      true,
		EnableConsole: true,
	}
}

// InitLogger initializes the enhanced centralized logger (GAP-SEC-009)
func InitLogger() *zap.Logger {
	config := DefaultLogConfig()
	envLevel := os.Getenv("LOG_LEVEL")
	if envLevel != "" {
		config.Level = envLevel
	}

	// Parse log level
	level := zapcore.InfoLevel
	switch config.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	// Enhanced encoder config with structured fields
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    "function",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if config.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Setup output writers
	var cores []zapcore.Core

	// File output with rotation (GAP-SEC-009: Log retention)
	if config.EnableFile {
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.FilePath,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		})
		fileCore := zapcore.NewCore(encoder, fileWriter, level)
		cores = append(cores, fileCore)
	}

	// Console output for container environments (stdout for ELK/Loki)
	if config.EnableConsole {
		consoleWriter := zapcore.AddSync(os.Stdout)
		consoleCore := zapcore.NewCore(encoder, consoleWriter, level)
		cores = append(cores, consoleCore)
	}

	// Combine cores
	core := zapcore.NewTee(cores...)

	// Create logger with caller info and stacktrace on error
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	zap.ReplaceGlobals(logger)
	return logger
}

// Logger is a named logger
type Logger struct {
	*zap.Logger
}

func GetNamedLogger(name string) *zap.Logger {
	return zap.L().Named(name)
}

func GetApiLogger() *zap.Logger {
	return zap.L().Named(ApiLogger)
}

func GetOutLogger() *zap.Logger {
	return zap.L().Named(Outlogger)
}
