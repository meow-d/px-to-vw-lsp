package main

import (
	"context"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"testing"
)

func createTestLogger(t *testing.T) *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stderr"}
	config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	logger, err := config.Build()
	if err != nil {
		t.Fatalf("Failed to create test logger: %v", err)
	}
	return logger
}

func TestParseCssremConfig(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    SchemaJson
		expectError bool
	}{
		{
			name: "Valid config with all fields",
			input: `{
				"$schema": "https://raw.githubusercontent.com/cipchk/vscode-cssrem/master/schema.json",
				"fixedDigits": 3,
				"vwDesign": 1920
			}`,
			expected: SchemaJson{
				VwDesign:    1920,
				FixedDigits: 3,
			},
			expectError: false,
		},
		{
			name: "Minimal config",
			input: `{
				"vwDesign": 1440
			}`,
			expected: SchemaJson{
				VwDesign:    1440,
				FixedDigits: 6, // SchemaJson default is 6, not 3
			},
			expectError: false,
		},
		{
			name: "Config with only precision",
			input: `{
				"fixedDigits": 2
			}`,
			expected: SchemaJson{
				VwDesign:    750, // SchemaJson default is 750, not 1440
				FixedDigits: 2,
			},
			expectError: false,
		},
		{
			name:  "Empty config",
			input: `{}`,
			expected: SchemaJson{
				VwDesign:    750, // SchemaJson default
				FixedDigits: 6,   // SchemaJson default
			},
			expectError: false,
		},
		{
			name: "Invalid JSON",
			input: `{
				"vwDesign": 1920
				"fixedDigits": 3
			}`,
			expectError: true,
		},
		{
			name:        "Invalid JSON structure",
			input:       `"vwDesign": 1920`,
			expectError: true,
		},
		{
			name: "Negative viewport width",
			input: `{
				"vwDesign": -1920
			}`,
			expected: SchemaJson{
				VwDesign:    -1920,
				FixedDigits: 6, // SchemaJson default is 6
			},
			expectError: false,
		},
		{
			name: "Zero precision",
			input: `{
				"vwDesign": 1920,
				"fixedDigits": 0
			}`,
			expected: SchemaJson{
				VwDesign:    1920,
				FixedDigits: 0,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCssremConfig([]byte(tt.input))

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.VwDesign != tt.expected.VwDesign {
				t.Errorf("VwDesign: got %f, want %f", result.VwDesign, tt.expected.VwDesign)
			}
			if result.FixedDigits != tt.expected.FixedDigits {
				t.Errorf("FixedDigits: got %g, want %g", result.FixedDigits, tt.expected.FixedDigits)
			}
		})
	}
}

func TestConvertToConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    SchemaJson
		expected Config
	}{
		{
			name: "Standard conversion",
			input: SchemaJson{
				VwDesign:    1920,
				FixedDigits: 3,
			},
			expected: Config{
				ViewportWidth: 1920,
				UnitPrecision: 3,
			},
		},
		{
			name: "Zero precision",
			input: SchemaJson{
				VwDesign:    1440,
				FixedDigits: 0,
			},
			expected: Config{
				ViewportWidth: 1440,
				UnitPrecision: 0,
			},
		},
		{
			name: "Large precision",
			input: SchemaJson{
				VwDesign:    2560,
				FixedDigits: 5,
			},
			expected: Config{
				ViewportWidth: 2560,
				UnitPrecision: 5,
			},
		},
		{
			name: "Decimal viewport width",
			input: SchemaJson{
				VwDesign:    1440.5,
				FixedDigits: 2,
			},
			expected: Config{
				ViewportWidth: 1440.5,
				UnitPrecision: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToConfig(tt.input)

			if result.ViewportWidth != tt.expected.ViewportWidth {
				t.Errorf("ViewportWidth: got %f, want %f", result.ViewportWidth, tt.expected.ViewportWidth)
			}
			if result.UnitPrecision != tt.expected.UnitPrecision {
				t.Errorf("UnitPrecision: got %d, want %d", result.UnitPrecision, tt.expected.UnitPrecision)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Test loading config from file
	t.Run("Load from valid config file", func(t *testing.T) {
		tempDir := t.TempDir()
		logger := createTestLogger(t)
		configFile := filepath.Join(tempDir, ".cssrem")
		configContent := `{
			"vwDesign": 1920,
			"fixedDigits": 2
		}`

		err := os.WriteFile(configFile, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test config file: %v", err)
		}

		config := loadConfig(tempDir, logger)

		if config.ViewportWidth != 1920 {
			t.Errorf("ViewportWidth: got %f, want 1920", config.ViewportWidth)
		}
		if config.UnitPrecision != 2 {
			t.Errorf("UnitPrecision: got %d, want 2", config.UnitPrecision)
		}
	})

	// Test loading config when file doesn't exist
	t.Run("Load when config file doesn't exist", func(t *testing.T) {
		tempDir := t.TempDir()
		logger := createTestLogger(t)
		config := loadConfig(tempDir, logger)

		// Should return default config
		defaultConfig := loadDefaultConfig()
		if config.ViewportWidth != defaultConfig.ViewportWidth {
			t.Errorf("ViewportWidth: got %f, want %f", config.ViewportWidth, defaultConfig.ViewportWidth)
		}
		if config.UnitPrecision != defaultConfig.UnitPrecision {
			t.Errorf("UnitPrecision: got %d, want %d", config.UnitPrecision, defaultConfig.UnitPrecision)
		}
	})

	// Test loading config when file has invalid JSON
	t.Run("Load with invalid JSON", func(t *testing.T) {
		tempDir := t.TempDir()
		logger := createTestLogger(t)
		configFile := filepath.Join(tempDir, ".cssrem")
		invalidContent := `{invalid json}`

		err := os.WriteFile(configFile, []byte(invalidContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test config file: %v", err)
		}

		config := loadConfig(tempDir, logger)

		// Should return default config on error
		defaultConfig := loadDefaultConfig()
		if config.ViewportWidth != defaultConfig.ViewportWidth {
			t.Errorf("ViewportWidth: got %f, want %f", config.ViewportWidth, defaultConfig.ViewportWidth)
		}
		if config.UnitPrecision != defaultConfig.UnitPrecision {
			t.Errorf("UnitPrecision: got %d, want %d", config.UnitPrecision, defaultConfig.UnitPrecision)
		}
	})
}

func TestLoadDefaultConfig(t *testing.T) {
	config := loadDefaultConfig()

	// Check default values
	if config.ViewportWidth != 1440 {
		t.Errorf("Default ViewportWidth: got %f, want 1440", config.ViewportWidth)
	}
	if config.UnitPrecision != 3 {
		t.Errorf("Default UnitPrecision: got %d, want 3", config.UnitPrecision)
	}
}

func TestGlobalConfig(t *testing.T) {
	// Test GlobalConfig creation
	t.Run("Create global config", func(t *testing.T) {
		ctx := context.Background()

		// Use a non-existent config path to avoid interfering with real config
		oldUserConfigDir := os.Getenv("XDG_CONFIG_HOME")
		defer func() {
			if oldUserConfigDir != "" {
				os.Setenv("XDG_CONFIG_HOME", oldUserConfigDir)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}
		}()

		// Set config dir to temp directory
		tempDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tempDir)

		logger := createTestLogger(t)

		globalConfig, err := NewGlobalConfig(ctx, logger)
		if err != nil {
			t.Fatalf("Failed to create global config: %v", err)
		}
		defer globalConfig.Close()

		config := globalConfig.Get()
		if config == nil {
			t.Error("Global config Get() returned nil")
		}
		if config.ViewportWidth != 1440 {
			t.Errorf("Default ViewportWidth: got %f, want 1440", config.ViewportWidth)
		}
		if config.UnitPrecision != 3 {
			t.Errorf("Default UnitPrecision: got %d, want 3", config.UnitPrecision)
		}
	})

	// Test GlobalConfig with file loading
	t.Run("Load global config from file", func(t *testing.T) {
		ctx := context.Background()
		tempDir := t.TempDir()

		// Create config directory and file
		configDir := filepath.Join(tempDir, ".config", "px-to-vw-lsp")
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		configFile := filepath.Join(configDir, "config.json")
		configContent := `{
			"vwDesign": 2560,
			"fixedDigits": 1
		}`

		err = os.WriteFile(configFile, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write global config file: %v", err)
		}

		// Temporarily redirect user config dir
		oldUserConfigDir := os.Getenv("XDG_CONFIG_HOME")
		oldHome := os.Getenv("HOME")
		defer func() {
			if oldUserConfigDir != "" {
				os.Setenv("XDG_CONFIG_HOME", oldUserConfigDir)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}
			os.Setenv("HOME", oldHome)
		}()

		os.Unsetenv("XDG_CONFIG_HOME")
		os.Setenv("HOME", tempDir)

		logger := createTestLogger(t)

		globalConfig, err := NewGlobalConfig(ctx, logger)
		if err != nil {
			t.Fatalf("Failed to create global config: %v", err)
		}
		defer globalConfig.Close()

		config := globalConfig.Get()
		if config == nil {
			t.Error("Global config Get() returned nil")
		}
		if config.ViewportWidth != 2560 {
			t.Errorf("Loaded ViewportWidth: got %f, want 2560", config.ViewportWidth)
		}
		if config.UnitPrecision != 1 {
			t.Errorf("Loaded UnitPrecision: got %d, want 1", config.UnitPrecision)
		}
	})
}

func TestHandlerLoadEffectiveConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create project config file
	configFile := filepath.Join(tempDir, ".cssrem")
	configContent := `{
		"vwDesign": 1920,
		"fixedDigits": 2
	}`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Create global config
	ctx := context.Background()
	oldUserConfigDir := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if oldUserConfigDir != "" {
			os.Setenv("XDG_CONFIG_HOME", oldUserConfigDir)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	tempGlobalDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tempGlobalDir)

	logger := createTestLogger(t)

	globalConfig, err := NewGlobalConfig(ctx, logger)
	if err != nil {
		t.Fatalf("Failed to create global config: %v", err)
	}
	defer globalConfig.Close()

	// Create handler
	handler := &Handler{
		globalConfig: globalConfig,
	}

	// Test loading effective config with project config
	effectiveConfig := handler.loadEffectiveConfig(globalConfig, tempDir, logger)

	if effectiveConfig.ViewportWidth != 1920 {
		t.Errorf("Effective ViewportWidth: got %f, want 1920 (project config should take priority)", effectiveConfig.ViewportWidth)
	}
	if effectiveConfig.UnitPrecision != 2 {
		t.Errorf("Effective UnitPrecision: got %d, want 2 (project config should take priority)", effectiveConfig.UnitPrecision)
	}
}
