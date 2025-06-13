package info

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/product"
	"go.flipt.io/flipt/internal/release"
)

// mockLicenseManager is a test double for the license manager interface
type mockLicenseManager struct{ val product.Product }

func (m mockLicenseManager) Product() product.Product { return m.val }

func TestNew(t *testing.T) {
	f := New(
		WithBuild("commit", "date", "goVersion", "version", true),
		WithLatestRelease(release.Info{LatestVersion: "latestVersion", LatestVersionURL: "latestVersionURL", UpdateAvailable: true}),
		WithConfig(config.Default()),
	)

	assert.Equal(t, "commit", f.Build.Commit)
	assert.Equal(t, "date", f.Build.BuildDate)
	assert.Equal(t, "goVersion", f.Build.GoVersion)
	assert.Equal(t, "version", f.Build.Version)
	assert.True(t, f.Build.IsRelease)
	assert.Equal(t, "latestVersion", f.Build.LatestVersion)
	assert.Equal(t, "latestVersionURL", f.Build.LatestVersionURL)
	assert.True(t, f.Build.UpdateAvailable)
	assert.False(t, f.Authentication.Required)
	assert.False(t, f.Analytics.Enabled)
}

func TestFlipt_ProductField_Marshaling(t *testing.T) {
	tests := []struct {
		name    string
		product product.Product
		expect  product.Product
	}{
		{"oss", product.OSS, product.OSS},
		{"pro", product.Pro, product.Pro},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := New(
				WithBuild("commit", "date", "goVersion", "version", true),
				WithLatestRelease(release.Info{LatestVersion: "latestVersion", LatestVersionURL: "latestVersionURL", UpdateAvailable: true}),
				WithConfig(config.Default()),
				WithLicenseManager(mockLicenseManager{tt.product}),
			)
			data, err := json.Marshal(f)
			assert.NoError(t, err)
			var out map[string]any
			assert.NoError(t, json.Unmarshal(data, &out))
			assert.Equal(t, string(tt.expect), out["product"])
		})
	}
}
