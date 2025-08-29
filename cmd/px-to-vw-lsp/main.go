package main

import (
  "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// TODO
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.OutputPaths = []string{ "stdout", "/tmp/px-to-vw-lsp.log" }
  logger, _ := zap.NewDevelopment()

  StartServer(logger)
}
