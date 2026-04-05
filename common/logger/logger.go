package logger

import (
	"github.com/syamsv/go-template/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger
var Sugar *zap.SugaredLogger

func Init() {
	var cfg zap.Config

	switch config.ENVIRONMENT {
	case "production", "staging":
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.EncoderConfig.StacktraceKey = "stacktrace"
	default:
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	var err error
	Log, err = cfg.Build(zap.AddCallerSkip(0))
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	Sugar = Log.Sugar()
	zap.ReplaceGlobals(Log)
}

func Sync() {
	_ = Log.Sync()
}

func With(fields ...zap.Field) *zap.Logger {
	return Log.With(fields...)
}

func Named(name string) *zap.Logger {
	return Log.Named(name)
}
