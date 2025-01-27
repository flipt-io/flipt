package info

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/release"
)

func TestNew(t *testing.T) {
	f := New(
		WithOS("linux", "amd64"),
		WithBuild("commit", "date", "goVersion", "version", true),
		WithLatestRelease(release.Info{LatestVersion: "latestVersion", LatestVersionURL: "latestVersionURL", UpdateAvailable: true}),
		WithConfig(config.Default()),
	)

	assert.Equal(t, "commit", f.Commit)
	assert.Equal(t, "date", f.BuildDate)
	assert.Equal(t, "goVersion", f.GoVersion)
	assert.Equal(t, "version", f.Version)
	assert.True(t, f.IsRelease)
	assert.Equal(t, "latestVersion", f.LatestVersion)
	assert.Equal(t, "latestVersionURL", f.LatestVersionURL)
	assert.True(t, f.UpdateAvailable)
	assert.Equal(t, "linux", f.OS)
	assert.Equal(t, "amd64", f.Arch)
	assert.False(t, f.Authentication.Required)
	assert.False(t, f.Analytics.Enabled)
}

func TestHttpHandler(t *testing.T) {
	f := New()
	r := httptest.NewRequest("GET", "/info", nil)
	w := httptest.NewRecorder()
	f.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `{"updateAvailable":false,"isRelease":false,"authentication":{"required":false},"storage":{"type":"database"}}`, w.Body.String())
}
