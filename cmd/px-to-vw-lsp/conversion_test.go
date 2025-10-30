package main

import (
	"regexp"
	"strconv"
	"testing"
)

func TestPxToVwConversion(t *testing.T) {
	tests := []struct {
		name          string
		pxValue       float64
		viewportWidth float64
		precision     int
		expectedVw    string
	}{
		{
			name:          "Basic conversion 1440px to vw",
			pxValue:       1440,
			viewportWidth: 1440,
			precision:     3,
			expectedVw:    "100.000",
		},
		{
			name:          "Half viewport width",
			pxValue:       720,
			viewportWidth: 1440,
			precision:     3,
			expectedVw:    "50.000",
		},
		{
			name:          "Quarter viewport width",
			pxValue:       360,
			viewportWidth: 1440,
			precision:     3,
			expectedVw:    "25.000",
		},
		{
			name:          "Decimal px value",
			pxValue:       1536,
			viewportWidth: 1440,
			precision:     3,
			expectedVw:    "106.667",
		},
		{
			name:          "Small px value",
			pxValue:       16,
			viewportWidth: 1440,
			precision:     3,
			expectedVw:    "1.111",
		},
		{
			name:          "Different viewport 1920",
			pxValue:       1920,
			viewportWidth: 1920,
			precision:     3,
			expectedVw:    "100.000",
		},
		{
			name:          "Precision 2",
			pxValue:       100,
			viewportWidth: 1440,
			precision:     2,
			expectedVw:    "6.94",
		},
		{
			name:          "Precision 1",
			pxValue:       100,
			viewportWidth: 1440,
			precision:     1,
			expectedVw:    "6.9",
		},
		{
			name:          "Zero precision (rounded)",
			pxValue:       100,
			viewportWidth: 1440,
			precision:     0,
			expectedVw:    "7",
		},
		{
			name:          "Large viewport width",
			pxValue:       1536,
			viewportWidth: 2560,
			precision:     3,
			expectedVw:    "60.000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vwValue := (tt.pxValue / tt.viewportWidth) * 100
			result := strconv.FormatFloat(vwValue, 'f', tt.precision, 64)

			if result != tt.expectedVw {
				t.Errorf("Conversion failed: %fpx at viewport %f with precision %d = %svw (expected %svw)",
					tt.pxValue, tt.viewportWidth, tt.precision, result, tt.expectedVw)
			}
		})
	}
}

func TestConfigMerging(t *testing.T) {
	tests := []struct {
		name           string
		defaultConfig  Config
		globalConfig   Config
		projectConfig  Config
		expectedConfig Config
	}{
		{
			name: "All configs default",
			defaultConfig: Config{
				ViewportWidth: 1440,
				UnitPrecision: 3,
			},
			globalConfig: Config{
				ViewportWidth: 0,
				UnitPrecision: 0,
			},
			projectConfig: Config{
				ViewportWidth: 0,
				UnitPrecision: 0,
			},
			expectedConfig: Config{
				ViewportWidth: 1440,
				UnitPrecision: 3,
			},
		},
		{
			name: "Global overrides default",
			defaultConfig: Config{
				ViewportWidth: 1440,
				UnitPrecision: 3,
			},
			globalConfig: Config{
				ViewportWidth: 1920,
				UnitPrecision: 2,
			},
			projectConfig: Config{
				ViewportWidth: 0,
				UnitPrecision: 0,
			},
			expectedConfig: Config{
				ViewportWidth: 1920,
				UnitPrecision: 2,
			},
		},
		{
			name: "Project overrides global and default",
			defaultConfig: Config{
				ViewportWidth: 1440,
				UnitPrecision: 3,
			},
			globalConfig: Config{
				ViewportWidth: 1920,
				UnitPrecision: 2,
			},
			projectConfig: Config{
				ViewportWidth: 2560,
				UnitPrecision: 1,
			},
			expectedConfig: Config{
				ViewportWidth: 2560,
				UnitPrecision: 1,
			},
		},
		{
			name: "Partial project config overrides",
			defaultConfig: Config{
				ViewportWidth: 1440,
				UnitPrecision: 3,
			},
			globalConfig: Config{
				ViewportWidth: 1920,
				UnitPrecision: 2,
			},
			projectConfig: Config{
				ViewportWidth: 2560,
				UnitPrecision: 0,
			},
			expectedConfig: Config{
				ViewportWidth: 2560,
				UnitPrecision: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeConfigs(tt.defaultConfig, tt.globalConfig, tt.projectConfig)

			if result.ViewportWidth != tt.expectedConfig.ViewportWidth {
				t.Errorf("ViewportWidth: got %f, want %f", result.ViewportWidth, tt.expectedConfig.ViewportWidth)
			}
			if result.UnitPrecision != tt.expectedConfig.UnitPrecision {
				t.Errorf("UnitPrecision: got %d, want %d", result.UnitPrecision, tt.expectedConfig.UnitPrecision)
			}
		})
	}
}

func TestRegexMatching(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple px value",
			input:    "width: 100px",
			expected: []string{"100px", "100", ""},
		},
		{
			name:     "Decimal px value",
			input:    "width: 100.5px",
			expected: []string{"100.5px", "100.5", ".5"},
		},
		{
			name:     "Media query",
			input:    "@media (min-width: 768px)",
			expected: []string{"768px", "768", ""},
		},
		{
			name:     "Negative px value",
			input:    "margin: -20px",
			expected: []string{"-20px", "-20", ""},
		},
		{
			name:     "No px value",
			input:    "width: 100%",
			expected: nil,
		},
		{
			name:     "Multiple px values",
			input:    "margin: 10px 20px",
			expected: []string{"20px", "20", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := regexp.MustCompile(`(-?\d+(\.\d+)?)px`)
			matches := re.FindAllStringSubmatch(tt.input, -1)

			if tt.expected == nil {
				if matches != nil {
					t.Errorf("Expected no match, got %v", matches)
				}
				return
			}

			if matches == nil {
				t.Errorf("Expected match %v, got nil", tt.expected)
				return
			}

			// Take the last match
			match := matches[len(matches)-1]

			if len(match) != len(tt.expected) {
				t.Errorf("Expected %d groups, got %d", len(tt.expected), len(match))
				return
			}

			for i, expectedValue := range tt.expected {
				if match[i] != expectedValue {
					t.Errorf("Group %d: got %q, want %q", i, match[i], expectedValue)
				}
			}
		})
	}
}
