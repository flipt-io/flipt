package util

import (
	"os/exec"
	"runtime"
)

// OpenBrowser opens the specified URL in the default browser of the user.
func OpenBrowser(url string) error {
	var (
		cmd  string
		args []string
	)

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
		args = append(args, "url.dll,FileProtocolHandler")
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}

	args = append(args, url)
	command := exec.Command(cmd, args...) //nolint noctx
	if err := command.Start(); err != nil {
		return err
	}
	return command.Process.Release()
}
