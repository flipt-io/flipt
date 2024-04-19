package policies

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/stretchr/testify/require"
)

func TestJSONSchema(t *testing.T) {
	schema, err := jsonschema.Compile("policy.schema.json")
	require.NoError(t, err)

	file, err := os.ReadFile("default.json")
	require.NoError(t, err)

	data := map[string]interface{}{}

	err = json.Unmarshal(file, &data)
	require.NoError(t, err)

	err = schema.Validate(data)
	require.NoError(t, err)
}
