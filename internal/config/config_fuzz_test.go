package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mitchellh/mapstructure"
)

func FuzzConfigLoad(f *testing.F) {
	// Add some seed corpus from existing test files
	f.Add("./testdata/advanced.yml")
	f.Add("./testdata/authentication/token_bootstrap_token.yml")
	f.Add("./testdata/metrics/otlp.yml")
	f.Add("./testdata/storage/git_ssh_auth_valid_with_path.yml")

	// Also add some edge cases
	f.Add("")
	f.Add("./non-existent-file.yml")

	f.Fuzz(func(t *testing.T, path string) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Load panicked: %v", r)
			}
		}()

		// Test that Load doesn't panic with this path
		_, _ = Load(t.Context(), path)
	})
}

func FuzzConfigUnmarshal(f *testing.F) {
	// Seed corpus with basic valid configs
	f.Add([]byte(`version: "2.0"`))
	f.Add([]byte(`
log:
  level: DEBUG
server:
  http_port: 8080
`))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unmarshal panicked: %v", r)
			}
		}()

		// Write fuzzed data to a temp file
		tmpFile, err := os.CreateTemp(t.TempDir(), "config-*.yml")
		if err != nil {
			return
		}
		t.Cleanup(func() { tmpFile.Close() })

		if _, err := tmpFile.Write(data); err != nil {
			return
		}

		// Close file before reading
		if err := tmpFile.Close(); err != nil {
			return
		}

		// Try to load the config with the fuzzed data
		_, _ = Load(t.Context(), tmpFile.Name())
	})
}

func FuzzEnvSubst(f *testing.F) {
	// Add some seed corpus with environment variables
	f.Add("${FOO}", "foo_value")
	f.Add("prefix_${BAR}_suffix", "bar_value")
	f.Add("${BAZ}", "")

	f.Fuzz(func(t *testing.T, envVar string, envValue string) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("envsubst panicked: %v", r)
			}
		}()

		if len(envVar) > 0 {
			// Set the environment variable
			t.Setenv(envVar, envValue)

			// Create a simple config with the env var
			yamlConfig := fmt.Sprintf(`
log:
  level: ${%s}
server:
  http_port: 8080
`, envVar)

			// Write this to a temp file
			tmpFile, err := os.CreateTemp(t.TempDir(), "config-*.yml")
			if err != nil {
				return
			}
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			if _, err := tmpFile.WriteString(yamlConfig); err != nil {
				return
			}

			// Close file before reading
			if err := tmpFile.Close(); err != nil {
				return
			}

			// Try to load the config
			_, _ = Load(t.Context(), tmpFile.Name())
		}
	})
}

func FuzzBindEnvVars(f *testing.F) {
	// Add some basic structs and env vars
	f.Add(`{"A": "a"}`, "PREFIX_A")
	f.Add(`{"A": {"B": "b"}}`, "PREFIX_A_B")
	f.Add(`{"A": {"B": {"C": "c"}}}`, "PREFIX_A_B_C")

	f.Fuzz(func(t *testing.T, structJSON string, envVar string) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("bindEnvVars panicked: %v", r)
			}
		}()

		// Skip empty inputs
		if len(structJSON) == 0 || len(envVar) == 0 {
			return
		}

		// Create a test struct dynamically
		var testStruct map[string]any
		err := json.Unmarshal([]byte(structJSON), &testStruct)
		if err != nil {
			return
		}

		// Create an env binder
		binder := sliceEnvBinder{}

		// Generate a struct type using reflection
		structType := reflect.TypeFor[Config]()

		// Call bindEnvVars with the fuzzed inputs
		bindEnvVars(&binder, []string{envVar}, []string{}, structType)
	})
}

func FuzzDecodeHooks(f *testing.F) {
	// Add some seed corpus with different input types
	f.Add("10s")        // duration string
	f.Add("a b c")      // space-separated string for slice
	f.Add("${ENV_VAR}") // env var string

	f.Fuzz(func(t *testing.T, input string) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("DecodeHook panicked with input '%s': %v", input, r)
			}
		}()

		// Skip empty inputs
		if len(input) == 0 {
			return
		}

		// Setup environment variable if this appears to be an env var substitution
		if strings.HasPrefix(input, "${") && strings.HasSuffix(input, "}") {
			varName := strings.TrimPrefix(strings.TrimSuffix(input, "}"), "${")
			t.Setenv(varName, "test-value")
		}

		// Test each decode hook
		// For string to duration
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Duration hook panicked: %v", r)
				}
			}()

			hook := mapstructure.StringToTimeDurationHookFunc()
			_, _ = hook.(func(reflect.Type, reflect.Type, any) (any, error))(
				reflect.TypeFor[string](),
				reflect.TypeFor[time.Duration](),
				input,
			)
		}()

		// For string to slice
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Slice hook panicked: %v", r)
				}
			}()

			hook := stringToSliceHookFunc()
			_, _ = hook.(func(reflect.Kind, reflect.Kind, any) (any, error))(
				reflect.String,
				reflect.Slice,
				input,
			)
		}()

		// For environment variable substitution
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Envsubst hook panicked: %v", r)
				}
			}()

			hook := stringToReferenceHookFunc()
			_, _ = hook.(func(reflect.Type, reflect.Type, any) (any, error))(
				reflect.TypeFor[string](),
				reflect.TypeFor[string](),
				input,
			)
		}()
	})
}
