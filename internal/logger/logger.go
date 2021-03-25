package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	instance   *zap.Logger
	logCfg     zap.Config
	encoderCfg = zap.NewProductionEncoderConfig()
)

func init() {
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
}

func NewDevLogger() (*zap.Logger, error) {
	logCfg := zap.NewProductionConfig()
	logCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logCfg.DisableStacktrace = false
	logCfg.EncoderConfig = encoderCfg
	return logCfg.Build()
}

func NewProdLogger() (*zap.Logger, error) {
	logCfg := zap.NewProductionConfig()
	logCfg.DisableStacktrace = true
	logCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	logCfg.EncoderConfig = encoderCfg
	return logCfg.Build()
}

func Prod() *zap.Logger {

	l, _ := NewProdLogger()
	instance = l

	return instance
}

func Debug() *zap.Logger {

	l, _ := NewDevLogger()
	instance = l

	return instance
}
