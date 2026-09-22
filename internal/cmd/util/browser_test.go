package util

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOpenBrowser(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper uses a POSIX shell script")
	}

	const url = "https://example.com/flags"
	dir := t.TempDir()
	marker := filepath.Join(dir, "opened-url")
	executable := "xdg-open"
	if runtime.GOOS == "darwin" {
		executable = "open"
	}

	launcher := filepath.Join(dir, executable)
	require.NoError(t, os.WriteFile(launcher, []byte("#!/bin/sh\nprintf '%s' \"$1\" > \"$BROWSER_MARKER\"\n"), 0o600))
	require.NoError(t, os.Chmod(launcher, 0o700))
	t.Setenv("PATH", dir)
	t.Setenv("BROWSER_MARKER", marker)

	require.NoError(t, OpenBrowser(url))
	require.Eventually(t, func() bool {
		openedURL, err := os.ReadFile(marker)
		return err == nil && string(openedURL) == url
	}, time.Second, 10*time.Millisecond)
}

func TestOpenBrowserReturnsLauncherError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PATH executable lookup differs on Windows")
	}

	t.Setenv("PATH", t.TempDir())

	require.Error(t, OpenBrowser("https://example.com/flags"))
}
