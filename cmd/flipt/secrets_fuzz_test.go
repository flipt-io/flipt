//go:build go1.18
// +build go1.18

package main

import (
	"context"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"go.flipt.io/flipt/internal/secrets"
)

// MockSecretsManager for fuzz testing
type MockSecretsManager struct {
	secrets map[string][]byte
}

func (m *MockSecretsManager) GetSecretValue(ctx context.Context, ref secrets.Reference) ([]byte, error) {
	key := ref.Provider + ":" + ref.Path + ":" + ref.Key
	if value, exists := m.secrets[key]; exists {
		return value, nil
	}
	return []byte("mock-value"), nil
}

func (m *MockSecretsManager) RegisterProvider(name string, provider secrets.Provider) error { return nil }
func (m *MockSecretsManager) GetProvider(name string) (secrets.Provider, error)             { return nil, nil }
func (m *MockSecretsManager) GetSecret(ctx context.Context, providerName, path string) (*secrets.Secret, error) {
	return nil, nil
}
func (m *MockSecretsManager) ListSecrets(ctx context.Context, providerName, pathPrefix string) ([]string, error) {
	return nil, nil
}
func (m *MockSecretsManager) ListProviders() []string { return nil }
func (m *MockSecretsManager) Close() error             { return nil }

func FuzzSecretReference(f *testing.F) {
	// Add seed corpus with valid secret references
	seeds := []string{
		"${secret:vault:path/to/secret:key}",
		"${secret:file:config:password}",
		"${secret:aws:production/db:username}",
		"${secret:key-only}",
		"${secret:provider:path:key}",
		"${secret:vault:app/prod/db:connection_string}",
		"${secret:file:/etc/secrets:api_key}",
		"prefix-${secret:vault:path:key}-suffix",
		"${secret:vault:nested/deep/path:complex_key_name}",
		"${secret:provider:path/with/slashes:key_with_underscores}",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	// Add edge cases and potential attack vectors
	edgeCases := []string{
		"",
		"${}",
		"${secret:}",
		"${secret:",
		"$secret:}",
		"${secret::",
		"${secret:::}",
		"${secret:provider:}",
		"${secret::path:key}",
		"${secret:provider::key}",
		"${secret:provider:path:}",
		"${secret:a:b:c:d:e:f}",  // Too many parts
		"${secret}" + strings.Repeat(":part", 100), // Very long reference
		"${secret:provider with spaces:path:key}",
		"${secret:provider:path with spaces:key}",
		"${secret:provider:path:key with spaces}",
		"${secret:provider\n:path:key}", // Newlines
		"${secret:provider\t:path:key}", // Tabs
		"${secret:provider${nested}:path:key}", // Nested references
		"${${secret:vault:path:key}}", // Double nesting
		"prefix${secret:vault:path:key}${secret:file:other:key}suffix", // Multiple refs
		// Unicode and special characters
		"${secret:provider:path/émoji🚀:key}",
		"${secret:provider:path:key-with-émojis}",
		// Very long components
		"${secret:" + strings.Repeat("a", 1000) + ":path:key}",
		"${secret:provider:" + strings.Repeat("b", 1000) + ":key}",
		"${secret:provider:path:" + strings.Repeat("c", 1000) + "}",
		// Binary data
		string([]byte{'$', '{', 's', 'e', 'c', 'r', 'e', 't', ':', 0x00, 0x01, 0xff, '}'}),
	}

	for _, edge := range edgeCases {
		f.Add(edge)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Secret reference processing panicked with input '%s': %v", input, r)
			}
		}()

		// Skip extremely large inputs to avoid timeout
		if len(input) > 10*1024 { // 10KB limit
			t.Skip("Input too large")
		}

		// Test the secret reference regex pattern
		secretReference := regexp.MustCompile(`^\${secret:([a-zA-Z0-9_:/-]+)}$`)
		matches := secretReference.MatchString(input)
		if matches {
			// If it matches, try to parse it
			reference := secretReference.ReplaceAllString(input, `$1`)
			parts := strings.Split(reference, ":")
			
			// Test the parsing logic from main.go
			var secretRef secrets.Reference
			switch {
			case len(parts) == 1:
				secretRef = secrets.Reference{
					Provider: "",
					Path:     "",
					Key:      parts[0],
				}
			case len(parts) >= 3:
				secretRef = secrets.Reference{
					Provider: parts[0],
					Path:     strings.Join(parts[1:len(parts)-1], ":"),
					Key:      parts[len(parts)-1],
				}
			}

			// Test validation
			_ = secretRef.Validate()
		}

		// Test the walkConfigForSecrets function with a mock config
		mockManager := &MockSecretsManager{
			secrets: map[string][]byte{
				"vault:path:key":     []byte("secret-value"),
				"file:config:pass":   []byte("password123"),
				"aws:prod:username":  []byte("admin"),
			},
		}

		// Create a test struct with the input value
		testConfig := struct {
			Value string `json:"value"`
		}{
			Value: input,
		}

		ctx := context.Background()
		v := reflect.ValueOf(&testConfig).Elem()
		_ = walkConfigForSecrets(ctx, v, mockManager)
	})
}

func FuzzSecretReferenceRegex(f *testing.F) {
	// Test the regex pattern specifically for ReDoS and edge cases
	seeds := []string{
		"${secret:a}",
		"${secret:a:b:c}",
		"${secret:" + strings.Repeat("a", 10) + "}",
		"${secret:" + strings.Repeat("a:b", 10) + "}",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	// Add potential ReDoS patterns
	redosPatterns := []string{
		"${secret:" + strings.Repeat("a", 100) + strings.Repeat("b", 100) + "}",
		"${secret:" + strings.Repeat("ab", 1000) + "}",
		"${secret:" + strings.Repeat("a/b/c", 100) + "}",
		"${secret:" + strings.Repeat("a_b_c", 100) + "}",
		"${secret:" + strings.Repeat("a-b-c", 100) + "}",
		// Patterns that might cause catastrophic backtracking
		"${secret:" + strings.Repeat("a", 50) + strings.Repeat(":", 50) + "}",
	}

	for _, pattern := range redosPatterns {
		f.Add(pattern)
	}

	f.Fuzz(func(t *testing.T, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Regex matching panicked with input length %d: %v", len(input), r)
			}
		}()

		// Skip extremely large inputs that could cause timeout
		if len(input) > 50*1024 { // 50KB limit
			t.Skip("Input too large for regex test")
		}

		// Test the regex pattern used in main.go
		secretReference := regexp.MustCompile(`^\${secret:([a-zA-Z0-9_:/-]+)}$`)
		
		// This should not panic or take excessive time
		_ = secretReference.MatchString(input)
		
		// Also test finding all matches in case the input has multiple patterns
		globalPattern := regexp.MustCompile(`\${secret:([a-zA-Z0-9_:/-]+)}`)
		_ = globalPattern.FindAllString(input, -1)
	})
}