package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	Outlogger = "outlogger"
	ApiLogger = "apilogger"
)

func InitLogger() *zap.Logger {
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "application.log",
		MaxSize:    10, // megabytes
		MaxBackups: 10,
		MaxAge:     30, // days
	})

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		w,
		zap.InfoLevel,
	)

	// backgroundCore, err := nrzap.WrapBackgroundCore(core, AddNewRelic())
	// if err != nil && err != nrzap.ErrNilApp {
	// 	panic(err)
	// }

	logger := zap.New(core)
	// logger := zap.New(backgroundCore)
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
