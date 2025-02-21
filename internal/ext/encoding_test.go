package ext

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncoding_NewEncoder(t *testing.T) {
	tests := []struct {
		name     string
		encoding Encoding
		data     interface{}
		want     string
		wantErr  bool
	}{
		{
			name:     "JSON encoding",
			encoding: EncodingJSON,
			data:     map[string]string{"key": "value"},
			want:     "{\"key\":\"value\"}\n",
		},
		{
			name:     "YAML encoding",
			encoding: EncodingYAML,
			data:     map[string]string{"key": "value"},
			want:     "key: value\n",
		},
		{
			name:     "YML encoding",
			encoding: EncodingYML,
			data:     map[string]string{"key": "value"},
			want:     "key: value\n",
		},
		{
			name:     "Invalid encoding",
			encoding: "invalid",
			data:     map[string]string{"key": "value"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			encoder := tt.encoding.NewEncoder(buf)
			if tt.wantErr {
				assert.Nil(t, encoder)
				return
			}
			require.NotNil(t, encoder)

			err := encoder.Encode(tt.data)
			require.NoError(t, err)

			err = encoder.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want, buf.String())
		})
	}
}

func TestEncoding_NewDecoder(t *testing.T) {
	tests := []struct {
		name     string
		encoding Encoding
		input    string
		want     map[string]string
		wantErr  bool
	}{
		{
			name:     "JSON decoding",
			encoding: EncodingJSON,
			input:    "{\"key\":\"value\"}",
			want:     map[string]string{"key": "value"},
		},
		{
			name:     "YAML decoding",
			encoding: EncodingYAML,
			input:    "key: value",
			want:     map[string]string{"key": "value"},
		},
		{
			name:     "YML decoding",
			encoding: EncodingYML,
			input:    "key: value",
			want:     map[string]string{"key": "value"},
		},
		{
			name:     "Invalid encoding",
			encoding: "invalid",
			input:    "key: value",
			wantErr:  true,
		},
		{
			name:     "Invalid JSON",
			encoding: EncodingJSON,
			input:    "{invalid json}",
			wantErr:  true,
		},
		{
			name:     "Invalid YAML",
			encoding: EncodingYAML,
			input:    "invalid: yaml: [",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewBufferString(tt.input)
			decoder := tt.encoding.NewDecoder(reader)
			if tt.wantErr {
				if decoder == nil {
					return
				}
				var result map[string]string
				err := decoder.Decode(&result)
				assert.Error(t, err)
				return
			}

			require.NotNil(t, decoder)
			var result map[string]string
			err := decoder.Decode(&result)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestNopCloseEncoder(t *testing.T) {
	t.Run("Close should return nil", func(t *testing.T) {
		encoder := NopCloseEncoder{}
		assert.NoError(t, encoder.Close())
	})

	t.Run("Should wrap encoder correctly", func(t *testing.T) {
		buf := &bytes.Buffer{}
		jsonEncoder := EncodingJSON.NewEncoder(buf)
		require.NotNil(t, jsonEncoder)

		data := map[string]string{"key": "value"}
		err := jsonEncoder.Encode(data)
		require.NoError(t, err)

		assert.Equal(t, "{\"key\":\"value\"}\n", buf.String())
	})
}
