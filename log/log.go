package log

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"docker-operator/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitZapLog() *zap.Logger {
	logWriteMode := config.DefaultConfig.GetString("LOG_WRITE_MODE")
	logDirectory := config.DefaultConfig.GetString("LOG_PATH")
	cfg := zap.Config{
		Encoding:         "json",
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "time",
			EncodeTime: zapcore.ISO8601TimeEncoder,
		},
	}
	switch logWriteMode {
	case "console":
		cfg.OutputPaths = []string{"stdout"}
	case "file":
		filename := fmt.Sprintf("%swatchdog-%s.log", logDirectory, time.Now().Format("2006-01-02"))
		if err := os.MkdirAll(filepath.Dir(filename), 0770); err != nil {
			panic(err)
		}
		_, err := os.Create(filename)
		if err != nil {
			panic(err)
		}
		cfg.OutputPaths = []string{filename}
	}
	switch config.DefaultConfig.GetString("LOG_LEVEL") {
	case "DEBUG":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "WARN":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "INFO":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "ERROR":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger
}
