//go:build go1.18
// +build go1.18

package fs

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

func FuzzParseConfig(f *testing.F) {
	// Add seed corpus with valid configurations
	seeds := []string{
		`version: "1.1"
namespace: default
flags:
  - key: example-flag
    name: Example Flag
    type: VARIANT_FLAG_TYPE
    enabled: true`,

		`version: "1.0"
namespace: production
flags: []
segments: []`,

		`version: "1.1"
namespace: test
flags:
  - key: test-flag
    name: Test Flag
    enabled: false
    variants:
      - key: variant1
        name: Variant 1`,

		// YAML with includes
		`version: "1.1"
namespace: main
include:
  - "flags/*.yml"
  - "segments/*.yml"`,

		// Complex nested structure
		`version: "1.1"
namespace: complex
flags:
  - key: complex-flag
    name: Complex Flag
    type: VARIANT_FLAG_TYPE
    enabled: true
    variants:
      - key: v1
        name: Variant 1
        attachment: |
          {"key": "value", "nested": {"data": true}}
    rules:
      - segment: segment1
        distributions:
          - variant: v1
            rollout: 100`,

		// Minimal config
		`version: "1.0"`,

		// Config with metadata
		`version: "1.1"
namespace: metadata-test
metadata:
  created_by: "system"
  environment: "test"
flags: []`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	// Add edge cases
	f.Add("") // Empty config
	f.Add("invalid yaml: [")
	f.Add("version: invalid")
	f.Add(`version: "1.1"
namespace: test
flags:
  - key: 
    name: Empty Key`)
	f.Add("---\n---\n---") // Multiple YAML documents
	f.Add("version: \"1.1\"\nnamespace: test\nflags:\n  - key: flag\n    name: Flag\n    enabled: not-a-boolean")

	f.Fuzz(func(t *testing.T, configData string) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("parseConfig panicked with input length %d: %v", len(configData), r)
			}
		}()

		// Test parsing configuration
		reader := strings.NewReader(configData)
		_, _ = parseConfig(zap.NewNop(), reader)
	})
}

func FuzzParseConfigLargeInputs(f *testing.F) {
	// Test with potentially large inputs that could cause memory issues
	baseConfig := `version: "1.1"
namespace: large-test
flags:`

	// Generate various sizes of input
	f.Add(baseConfig)

	// Add a config with many flags
	largeConfig := baseConfig + "\n"
	for i := 0; i < 10; i++ {
		largeConfig += `  - key: flag` + string(rune('0'+i)) + `
    name: Flag ` + string(rune('0'+i)) + `
    enabled: true
`
	}
	f.Add(largeConfig)

	f.Fuzz(func(t *testing.T, configData string) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("parseConfig panicked with large input length %d: %v", len(configData), r)
			}
		}()

		// Skip extremely large inputs to avoid timeout
		if len(configData) > 1024*1024 { // 1MB limit
			t.Skip("Input too large")
		}

		reader := strings.NewReader(configData)
		_, _ = parseConfig(zap.NewNop(), reader)
	})
}

func FuzzParseConfigSpecialCharacters(f *testing.F) {
	// Test with various special characters and encodings
	seeds := []string{
		`version: "1.1"
namespace: "special-chars"
flags:
  - key: "flag-with-unicode-🚀"
    name: "Flag with émojis and spëcial chars"`,

		`version: "1.1"
namespace: "quotes"
flags:
  - key: 'single-quoted'
    name: "double-quoted"
    description: 'mixed "quotes" here'`,

		// YAML with special characters
		`version: "1.1"
namespace: "special"
flags:
  - key: "flag\nwith\nnewlines"
    name: "Flag\twith\ttabs"`,

		// Binary-like data
		string([]byte{0x00, 0x01, 0x02, 0xff, 0xfe}),

		// Very long strings
		`version: "1.1"
namespace: "` + strings.Repeat("a", 1000) + `"`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, configData string) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("parseConfig panicked with special characters input: %v", r)
			}
		}()

		// Skip inputs that are too large
		if len(configData) > 100*1024 { // 100KB limit
			t.Skip("Input too large")
		}

		reader := strings.NewReader(configData)
		_, _ = parseConfig(zap.NewNop(), reader)
	})
}
