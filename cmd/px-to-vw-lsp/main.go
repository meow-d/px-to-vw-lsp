package main

import (
	"flag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func parseFlags() (logLevel, logFile string) {
	level := flag.String("log-level", "warn", "log level (debug, info, warn, error)")
	file := flag.String("log-file", "/tmp/px-to-vw-lsp.log", "log file path")
	flag.Parse()
	return *level, *file
}

func initLogger(logLevel, logFile string) *zap.Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.OutputPaths = []string{"stdout", logFile}

	level := zapcore.WarnLevel
	switch logLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}
	cfg.Level = zap.NewAtomicLevelAt(level)

	logger, _ := cfg.Build()
	return logger
}

func main() {
	logLevel, logFile := parseFlags()
	logger := initLogger(logLevel, logFile)
	defer logger.Sync()

	StartServer(logger)
}
