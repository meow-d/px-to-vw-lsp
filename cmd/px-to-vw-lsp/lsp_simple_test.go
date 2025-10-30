package main

import (
	"regexp"
	"strconv"
	"testing"
)

func TestRegexEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectMatch bool
		expectedPx  string
	}{
		{
			name:        "Media query with spaces",
			input:       "@media (min-width: 1536px",
			expectMatch: true,
			expectedPx:  "1536",
		},
		{
			name:        "Property value at end of string",
			input:       "width: 800px",
			expectMatch: true,
			expectedPx:  "800",
		},
		{
			name:        "Property with decimal",
			input:       "margin: 20.5px",
			expectMatch: true,
			expectedPx:  "20.5",
		},
		{
			name:        "Negative value",
			input:       "margin: -10px",
			expectMatch: true,
			expectedPx:  "-10",
		},
		{
			name:        "Multiple px values - should match last",
			input:       "margin: 10px 20px",
			expectMatch: true,
			expectedPx:  "20",
		},
		{
			name:        "Percentage - no match",
			input:       "width: 100%",
			expectMatch: false,
		},
		{
			name:        "Em value - no match",
			input:       "font-size: 16em",
			expectMatch: false,
		},
		{
			name:        "Rem value - no match",
			input:       "font-size: 1.2rem",
			expectMatch: false,
		},
		{
			name:        "Incomplete px",
			input:       "width: 100p",
			expectMatch: false,
		},
		{
			name:        "No unit",
			input:       "width: 100",
			expectMatch: false,
		},
	}

	re := regexp.MustCompile(`(-?\d+(\.\d+)?)px`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := re.FindAllStringSubmatch(tt.input, -1)

			if !tt.expectMatch {
				if matches != nil {
					t.Errorf("Expected no match, got %v", matches)
				}
				return
			}

			if matches == nil {
				t.Errorf("Expected match for %q, got nil", tt.input)
				return
			}

			// Take the last match
			match := matches[len(matches)-1]
			if match[1] != tt.expectedPx {
				t.Errorf("Expected px value %q, got %q", tt.expectedPx, match[1])
			}
		})
	}
}

func TestCursorPositionScenarios(t *testing.T) {
	tests := []struct {
		name           string
		line           string
		cursorPos      int
		expectedMatch  bool
		expectedPx     string
		expectedPrefix string
	}{
		{
			name:           "Cursor at end of px",
			line:           "@media (min-width: 1536px",
			cursorPos:      25, // End of string
			expectedMatch:  true,
			expectedPx:     "1536",
			expectedPrefix: "@media (min-width: 1536px",
		},
		{
			name:          "Cursor in middle of px value",
			line:          "@media (min-width: 1536px",
			cursorPos:     24, // After '6'
			expectedMatch: false,
		},
		{
			name:          "Cursor at end of p in px",
			line:          "@media (min-width: 1536p",
			cursorPos:     24, // After 'p'
			expectedMatch: false,
		},
		{
			name:          "Cursor just before px",
			line:          "@media (min-width: 1536 px",
			cursorPos:     26, // After space
			expectedMatch: false,
		},
	}

	re := regexp.MustCompile(`(-?\d+(\.\d+)?)px`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix := tt.line[:tt.cursorPos]
			match := re.FindStringSubmatch(prefix)

			if !tt.expectedMatch {
				if match != nil {
					t.Errorf("Expected no match, got %v", match)
				}
				return
			}

			if match == nil {
				t.Errorf("Expected match for prefix %q, got nil", prefix)
				return
			}

			if match[1] != tt.expectedPx {
				t.Errorf("Expected px value %q, got %q", tt.expectedPx, match[1])
			}

			if tt.expectedPrefix != "" && prefix != tt.expectedPrefix {
				t.Errorf("Expected prefix %q, got %q", tt.expectedPrefix, prefix)
			}
		})
	}
}

func TestConfigIntegration(t *testing.T) {
	// Test different viewport sizes and precision values
	tests := []struct {
		name          string
		pxValue       float64
		viewportWidth float64
		precision     int
		expectedVw    string
	}{
		{
			name:          "Standard desktop 1536px at 1440 viewport",
			pxValue:       1536,
			viewportWidth: 1440,
			precision:     3,
			expectedVw:    "106.667",
		},
		{
			name:          "Standard desktop 1536px at 1920 viewport",
			pxValue:       1536,
			viewportWidth: 1920,
			precision:     3,
			expectedVw:    "80.000",
		},
		{
			name:          "Same as viewport width",
			pxValue:       1440,
			viewportWidth: 1440,
			precision:     3,
			expectedVw:    "100.000",
		},
		{
			name:          "Very small value",
			pxValue:       1,
			viewportWidth: 1440,
			precision:     5,
			expectedVw:    "0.06944",
		},
		{
			name:          "High precision",
			pxValue:       100,
			viewportWidth: 1440,
			precision:     5,
			expectedVw:    "6.94444",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				ViewportWidth: tt.viewportWidth,
				UnitPrecision: tt.precision,
			}

			// Simulate conversion logic
			vwValue := (tt.pxValue / config.ViewportWidth) * 100
			result := formatFloat(vwValue, config.UnitPrecision)

			if result != tt.expectedVw {
				t.Errorf("Conversion: %fpx at %.0f viewport with precision %d = %svw (expected %svw)",
					tt.pxValue, tt.viewportWidth, tt.precision, result, tt.expectedVw)
			}
		})
	}
}

// formatFloat helper function for testing
func formatFloat(f float64, precision int) string {
	return strconv.FormatFloat(f, 'f', precision, 64)
}
