// internal/logger/logger.go
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger

func InitLogger() error {
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg), // JSON encoding
		zapcore.AddSync(os.Stdout),         // Output to stdout
		zapcore.InfoLevel,                  // Minimum log level
	)

	logger := zap.New(core, zap.AddCaller())
	Log = logger.Sugar()
	return nil
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
