package logger

import (
	"log"
	"log/slog"
	"os"
)

var logLevel = slog.LevelInfo
var logger *slog.Logger

func Init() {
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: &logLevel}))
}

func Terminate() {
	logger = nil
}

func SetDebug() {
	logLevel = slog.LevelDebug
}

func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}

func Fatal(msg string) {
	log.Fatal(msg)
}
