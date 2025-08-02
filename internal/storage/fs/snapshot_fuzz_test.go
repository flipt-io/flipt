//go:build go1.18
// +build go1.18

package fs

import (
	"strings"
	"testing"
	"testing/fstest"

	"go.uber.org/zap"
)

func FuzzSnapshotFromFS(f *testing.F) {
	// Add seed corpus with valid file structures

	// Valid YAML flag file
	validYAML := `version: "1.1"
namespace: default
flags:
  - key: test-flag
    name: Test Flag
    type: VARIANT_FLAG_TYPE
    enabled: true
    variants:
      - key: on
        name: "On"
      - key: off  
        name: "Off"
    rules:
      - segment: test-segment
        distributions:
          - variant: on
            rollout: 50
          - variant: off
            rollout: 50
segments:
  - key: test-segment
    name: Test Segment
    match_type: ALL_MATCH_TYPE
    constraints:
      - type: STRING_COMPARISON_TYPE
        property: user_id
        operator: eq
        value: "test"`

	// Valid JSON flag file
	validJSON := `{
  "version": "1.1",
  "namespace": "default",
  "flags": [
    {
      "key": "json-flag",
      "name": "JSON Flag",
      "type": "BOOLEAN_FLAG_TYPE",
      "enabled": true
    }
  ]
}`

	// YAML stream (multiple documents)
	yamlStream := `---
version: "1.1"
namespace: stream1
flags:
  - key: flag1
    name: Flag 1
    enabled: true
---
version: "1.1" 
namespace: stream2
flags:
  - key: flag2
    name: Flag 2
    enabled: false`

	// Edge case: empty files
	f.Add("", ".yml")
	f.Add("", ".json")
	f.Add("{}", ".json")
	f.Add("---", ".yml")

	// Valid files
	f.Add(validYAML, ".yml")
	f.Add(validJSON, ".json")
	f.Add(yamlStream, ".yml")

	f.Fuzz(func(t *testing.T, content, extension string) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Snapshot creation panicked with %s file (length %d): %v",
					extension, len(content), r)
			}
		}()

		// Skip extremely large inputs
		if len(content) > 512*1024 { // 512KB limit
			t.Skip("Input too large")
		}

		// Create a test filesystem with the fuzzed content
		filename := "features" + extension
		testFS := fstest.MapFS{
			filename: &fstest.MapFile{
				Data: []byte(content),
			},
		}

		logger := zap.NewNop()
		config := &Config{} // Use default config

		// Test creating snapshot from filesystem
		_, _ = SnapshotFromFS(logger, config, testFS)
	})
}

func FuzzSnapshotJSONParsing(f *testing.F) {
	// Specific tests for JSON parsing edge cases
	seeds := []string{
		`{"version": "1.1", "flags": []}`,
		`{"version": "1.1", "namespace": "test"}`,
		`{"flags": null}`,
		`{"flags": [null]}`,
		`{"flags": [{"key": null}]}`,
		`{"flags": [{"enabled": "not-boolean"}]}`,
		// Deeply nested structures
		`{"flags": [{"rules": [{"distributions": [{"rollout": 99.999999999999999}]}]}]}`,
		// Large numbers
		`{"flags": [{"rules": [{"distributions": [{"rollout": 999999999999999999999}]}]}]}`,
		// Unicode and special characters
		`{"flags": [{"key": "🚀", "name": "émoji-flag"}]}`,
		// Malformed JSON
		`{"flags": [}`,
		`{"version": "1.1"`,
		`{"flags": [{"key": "test",}]}`,
		// Empty arrays and objects
		`{}`,
		`{"flags": [{}]}`,
		`{"segments": [{"constraints": []}]}`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, jsonContent string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("JSON parsing panicked with input length %d: %v", len(jsonContent), r)
			}
		}()

		if len(jsonContent) > 256*1024 { // 256KB limit
			t.Skip("Input too large")
		}

		testFS := fstest.MapFS{
			"features.json": &fstest.MapFile{
				Data: []byte(jsonContent),
			},
		}

		logger := zap.NewNop()
		config := &Config{} // Use default config

		_, _ = SnapshotFromFS(logger, config, testFS)
	})
}

func FuzzSnapshotYAMLParsing(f *testing.F) {
	// Specific tests for YAML parsing edge cases
	seeds := []string{
		"version: \"1.1\"\nflags: []",
		"version: 1.1\nnamespace: test",
		"flags:\n- key: test\n  enabled: yes",
		"flags:\n- key: test\n  enabled: no",
		"flags:\n- key: test\n  enabled: true",
		"flags:\n- key: test\n  enabled: false",
		// YAML with different boolean representations
		"flags:\n- enabled: on\n- enabled: off",
		"flags:\n- enabled: True\n- enabled: False",
		// YAML anchors and aliases
		"flags:\n- &flag\n  key: test\n- <<: *flag\n  key: test2",
		// Multi-line strings
		"flags:\n- description: |\n    This is a\n    multi-line\n    description",
		// YAML with numbers
		"flags:\n- rollout: 50\n- rollout: 0.5\n- rollout: 1e2",
		// Malformed YAML
		"flags:\n- key: test\n    invalid indent",
		"flags: [\n  key: test",
		"version: \"1.1\n", // Unclosed quote
		// Special YAML constructs
		"flags: !!null",
		"flags:\n- key: !!str 123",
		// Very deep nesting
		strings.Repeat("nested:\n  ", 50) + "value: deep",
		// Binary data
		"data: !!binary |\n  " + string([]byte{0x00, 0x01, 0x02, 0xff}),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, yamlContent string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("YAML parsing panicked with input length %d: %v", len(yamlContent), r)
			}
		}()

		if len(yamlContent) > 256*1024 { // 256KB limit
			t.Skip("Input too large")
		}

		testFS := fstest.MapFS{
			"features.yml": &fstest.MapFile{
				Data: []byte(yamlContent),
			},
		}

		logger := zap.NewNop()
		config := &Config{} // Use default config

		_, _ = SnapshotFromFS(logger, config, testFS)
	})
}

func FuzzSnapshotMultipleFiles(f *testing.F) {
	// Test with multiple files that could have conflicts or interactions
	f.Add("file1.yml", "version: \"1.1\"\nnamespace: ns1", "file2.yml", "version: \"1.1\"\nnamespace: ns2")
	f.Add("flags.yml", "flags:\n- key: flag1", "segments.yml", "segments:\n- key: seg1")

	f.Fuzz(func(t *testing.T, file1Name, file1Content, file2Name, file2Content string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Multiple file parsing panicked: %v", r)
			}
		}()

		// Skip if filenames are too long or empty
		if len(file1Name) == 0 || len(file2Name) == 0 || len(file1Name) > 100 || len(file2Name) > 100 {
			t.Skip("Invalid filenames")
		}

		// Skip large content
		if len(file1Content) > 64*1024 || len(file2Content) > 64*1024 {
			t.Skip("Content too large")
		}

		// Ensure files have proper extensions
		if !strings.HasSuffix(file1Name, ".yml") && !strings.HasSuffix(file1Name, ".yaml") && !strings.HasSuffix(file1Name, ".json") {
			file1Name += ".yml"
		}
		if !strings.HasSuffix(file2Name, ".yml") && !strings.HasSuffix(file2Name, ".yaml") && !strings.HasSuffix(file2Name, ".json") {
			file2Name += ".yml"
		}

		testFS := fstest.MapFS{
			file1Name: &fstest.MapFile{Data: []byte(file1Content)},
			file2Name: &fstest.MapFile{Data: []byte(file2Content)},
		}

		logger := zap.NewNop()
		config := &Config{} // Use default config

		_, _ = SnapshotFromFS(logger, config, testFS)
	})
}
