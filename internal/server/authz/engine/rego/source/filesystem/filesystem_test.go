package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalDataSourceGetReturnsDecodedData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	if err := os.WriteFile(path, []byte(`{"role":"admin"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	data, _, err := DataSourceFromPath(path).Get(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if data["role"] != "admin" {
		t.Fatalf("role = %v, want admin", data["role"])
	}
}
