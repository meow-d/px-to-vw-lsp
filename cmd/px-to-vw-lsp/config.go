package main

import (
	"os"
	"path/filepath"
)

// config
type Config struct {
	ViewportWidth float64 `json:"viewportWidth"`
	UnitPrecision int     `json:"unitPrecision"`
}

func loadDefaultConfig() Config {
	return Config{
		ViewportWidth: 1440,
		UnitPrecision: 3,
	}
}

func loadConfig(root string) Config {
	defaultConfig := loadDefaultConfig()

	cssremPath := filepath.Join(root, ".cssrem")
	file, err := os.ReadFile(cssremPath)
	if err != nil {
		log.Sugar().Warnf("Failed to open config file %s: %v", cssremPath, err)
		return defaultConfig
	}

	cssremConfig, err := parseCssremConfig(file)
	if err != nil {
		log.Sugar().Warnf("Failed to parse config file %s: %v", cssremPath, err)
		return defaultConfig
	}

	config := convertToConfig(*cssremConfig)
	log.Sugar().Infof("Loaded config from %s: viewport=%.0f, precision=%d",
		cssremPath, config.ViewportWidth, config.UnitPrecision)

	return config
}

func parseCssremConfig(data []byte) (*SchemaJson, error) {
	cssremConfig := SchemaJson{}
	if err := cssremConfig.UnmarshalJSON(data); err != nil {
		return nil, err
	}
	return &cssremConfig, nil
}

func convertToConfig(schema SchemaJson) Config {
	return Config{
		ViewportWidth: schema.VwDesign,
		UnitPrecision: int(schema.FixedDigits),
	}
}
