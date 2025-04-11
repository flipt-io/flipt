package info

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/release"
)

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
