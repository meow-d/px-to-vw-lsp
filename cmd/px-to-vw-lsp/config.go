package main

import (
	"os"
)

// config
type Config struct {
	ViewportWidth float64 `json:"viewportWidth"`
	UnitPrecision int     `json:"unitPrecision"`
}

func loadDefaultConfig() Config {
	return Config{
		ViewportWidth: 1920,
		UnitPrecision: 3,
	}
}

func loadConfig(root string) Config {
	defaultConfig := Config{
		ViewportWidth: 1920,
		UnitPrecision: 3,
	}

	file, err := os.ReadFile(root + "/.cssrem")
	if err != nil {
		log.Sugar().Warnf("Failed to open config file: %v", err)
		return defaultConfig
	}

	// SchemaJson is auto generated with https://github.com/omissis/go-jsonschema
	// from cssrem schema: https://raw.githubusercontent.com/cipchk/vscode-cssrem/master/schema.json
	cssremConfig := SchemaJson{}
	err = cssremConfig.UnmarshalJSON(file)
	if err != nil {
		log.Sugar().Warnf("Failed to parse config file: %v", err)
		return defaultConfig
	}
	log.Sugar().Infof("Loaded config: %+v", cssremConfig)

	return Config{
		ViewportWidth: cssremConfig.VwDesign,
		UnitPrecision: int(cssremConfig.FixedDigits),
	}
}
