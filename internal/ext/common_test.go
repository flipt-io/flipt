package ext

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNamespaceEmbed_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    *NamespaceEmbed
		expected string
		wantErr  bool
	}{
		{
			name:     "marshal namespace key",
			input:    &NamespaceEmbed{IsNamespace: NamespaceKey("test-ns")},
			expected: `"test-ns"`,
		},
		{
			name: "marshal full namespace",
			input: &NamespaceEmbed{IsNamespace: &Namespace{
				Key:         "test-ns",
				Name:        "Test Namespace",
				Description: "A test namespace",
			}},
			expected: `{"key":"test-ns","name":"Test Namespace","description":"A test namespace"}`,
		},
		{
			name:    "marshal nil namespace",
			input:   &NamespaceEmbed{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}

func TestNamespaceEmbed_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected IsNamespace
	}{
		{
			name:     "unmarshal namespace key",
			input:    `"test-ns"`,
			expected: NamespaceKey("test-ns"),
		},
		{
			name:  "unmarshal full namespace",
			input: `{"key":"test-ns","name":"Test Namespace","description":"A test namespace"}`,
			expected: &Namespace{
				Key:         "test-ns",
				Name:        "Test Namespace",
				Description: "A test namespace",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ns NamespaceEmbed
			err := json.Unmarshal([]byte(tt.input), &ns)

			require.NoError(t, err)
			assert.Equal(t, tt.expected.GetKey(), ns.IsNamespace.GetKey())

			// For full namespace objects, check all fields
			if fullNs, ok := tt.expected.(*Namespace); ok {
				actualNs := ns.IsNamespace.(*Namespace)
				assert.Equal(t, fullNs.Name, actualNs.Name)
				assert.Equal(t, fullNs.Description, actualNs.Description)
			}
		})
	}
}

func TestNamespaceEmbed_YAML(t *testing.T) {
	tests := []struct {
		name     string
		input    *NamespaceEmbed
		expected string
		wantErr  bool
	}{
		{
			name:     "marshal namespace key",
			input:    &NamespaceEmbed{IsNamespace: NamespaceKey("test-ns")},
			expected: "test-ns\n",
		},
		{
			name: "marshal full namespace",
			input: &NamespaceEmbed{IsNamespace: &Namespace{
				Key:         "test-ns",
				Name:        "Test Namespace",
				Description: "A test namespace",
			}},
			expected: "key: test-ns\nname: Test Namespace\ndescription: A test namespace\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			data, err := yaml.Marshal(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))

			// Test unmarshaling
			var ns NamespaceEmbed
			err = yaml.Unmarshal(data, &ns)
			require.NoError(t, err)
			assert.Equal(t, tt.input.IsNamespace.GetKey(), ns.IsNamespace.GetKey())
		})
	}
}

func TestNamespace_GetKey(t *testing.T) {
	tests := []struct {
		name     string
		ns       *Namespace
		expected string
	}{
		{
			name:     "nil namespace",
			ns:       nil,
			expected: "",
		},
		{
			name: "valid namespace",
			ns: &Namespace{
				Key: "test-ns",
			},
			expected: "test-ns",
		},
		{
			name: "empty key",
			ns: &Namespace{
				Key: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.ns.GetKey())
		})
	}
}

func TestDocument_Serialization(t *testing.T) {
	doc := &Document{
		Version: "1.5",
		Namespace: &NamespaceEmbed{
			IsNamespace: &Namespace{
				Key:         "test-ns",
				Name:        "Test Namespace",
				Description: "A test namespace",
			},
		},
		Flags: []*Flag{
			{
				Key:         "test-flag",
				Name:        "Test Flag",
				Type:        "boolean",
				Description: "A test flag",
				Enabled:     true,
			},
		},
	}

	t.Run("JSON serialization", func(t *testing.T) {
		data, err := json.Marshal(doc)
		require.NoError(t, err)

		var decoded Document
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, doc.Version, decoded.Version)
		assert.Equal(t, doc.Namespace.GetKey(), decoded.Namespace.GetKey())
		assert.Equal(t, len(doc.Flags), len(decoded.Flags))
		assert.Equal(t, doc.Flags[0].Key, decoded.Flags[0].Key)
	})

	t.Run("YAML serialization", func(t *testing.T) {
		data, err := yaml.Marshal(doc)
		require.NoError(t, err)

		var decoded Document
		err = yaml.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, doc.Version, decoded.Version)
		assert.Equal(t, doc.Namespace.GetKey(), decoded.Namespace.GetKey())
		assert.Equal(t, len(doc.Flags), len(decoded.Flags))
		assert.Equal(t, doc.Flags[0].Key, decoded.Flags[0].Key)
	})
}
