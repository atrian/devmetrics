package logger

import (
	"log"

	"go.uber.org/zap"
)

type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger() *ZapLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	return &ZapLogger{
		logger: logger,
	}
}

func (z ZapLogger) Fatal(message string, err error) {
	z.logger.Fatal(message, zap.Error(err))
}

func (z ZapLogger) Panic(message string, err error) {
	z.logger.Panic(message, zap.Error(err))
}

func (z ZapLogger) Error(message string, err error) {
	z.logger.Error(message, zap.Error(err))
}

func (z ZapLogger) Warning(message string) {
	z.logger.Warn(message)
}

func (z ZapLogger) Info(message string) {
	z.logger.Info(message)
}

func (z ZapLogger) Debug(message string) {
	z.logger.Debug(message)
}

func (z ZapLogger) Sync() {
	z.logger.Sync()
}
