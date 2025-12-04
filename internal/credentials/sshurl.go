package credentials

import (
	"fmt"
	"regexp"
	"strings"
)

// sshURLRegex matches SCP-style SSH URLs like git@github.com:owner/repo.git
// or user@host:path format. Note: This does NOT match ssh:// protocol URLs.
var sshURLRegex = regexp.MustCompile(`^([^@]+)@([^:]+):(.+)$`)

// NormalizeSSHRemoteURL transforms a remote URL for use with SSH credentials.
// It handles the following cases:
//  1. Strips https:// or http:// protocol if present (SSH doesn't use HTTP(S) URLs)
//  2. Converts ssh:// protocol URLs to SCP-style format
//  3. Prepends the SSH user to the URL if not already present
//  4. Returns an error if the URL contains a user that differs from sshUser
//
// The sshUser parameter is the user from SSH credentials configuration.
// If empty, it defaults to "git".
//
// Examples:
//   - https://github.com/org/repo.git + user="git" -> git@github.com:org/repo.git
//   - github.com/org/repo.git + user="git" -> git@github.com:org/repo.git
//   - git@github.com:org/repo.git + user="git" -> git@github.com:org/repo.git (unchanged)
//   - git@github.com:org/repo.git + user="other" -> error (conflicting users)
//   - ssh://git@github.com/org/repo.git + user="git" -> git@github.com:org/repo.git
func NormalizeSSHRemoteURL(remoteURL, sshUser string) (string, error) {
	if sshUser == "" {
		sshUser = "git"
	}

	url := strings.TrimSpace(remoteURL)

	// Handle ssh:// protocol URL format FIRST (before regex check)
	// because ssh://git@host:port/path would incorrectly match the SCP regex
	if strings.HasPrefix(url, "ssh://") {
		return convertSSHProtocolURL(url, sshUser)
	}

	// Check if it's already a valid SCP-style SSH URL (user@host:path)
	if matches := sshURLRegex.FindStringSubmatch(url); matches != nil {
		existingUser := matches[1]
		if existingUser != sshUser {
			return "", fmt.Errorf("conflicting SSH users: URL contains %q but credentials specify %q", existingUser, sshUser)
		}
		// URL is already in correct format
		return url, nil
	}

	// Strip https:// or http:// protocol (including any port like :443 or :80)
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	// Handle URLs with user@ prefix (e.g., from stripping protocol off git@host/path)
	// The condition checks: is there an @ before the first / (if any)?
	// If slashIndex is -1 (no slash), we skip this block and fall through to error handling.
	atIndex := strings.Index(url, "@")
	slashIndex := strings.Index(url, "/")
	if atIndex != -1 && (slashIndex == -1 || atIndex < slashIndex) {
		existingUser := url[:atIndex]
		if existingUser != sshUser {
			return "", fmt.Errorf("conflicting SSH users: URL contains %q but credentials specify %q", existingUser, sshUser)
		}
		// Strip user@ and continue processing
		url = url[atIndex+1:]
		slashIndex = strings.Index(url, "/")
	}

	// At this point, url should be: host/path or host:port/path
	if slashIndex == -1 {
		return "", fmt.Errorf("invalid remote URL format: %q (expected host/path)", remoteURL)
	}

	host := url[:slashIndex]
	path := url[slashIndex+1:]

	// Strip port from host if present (e.g., github.com:443 -> github.com)
	if colonIndex := strings.LastIndex(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	return sshUser + "@" + host + ":" + path, nil
}

// convertSSHProtocolURL converts ssh:// protocol URLs to SCP-style format.
// ssh://[user@]host[:port]/path -> user@host:path
func convertSSHProtocolURL(url, sshUser string) (string, error) {
	url = strings.TrimPrefix(url, "ssh://")

	// Extract user if present (user@host:port/path or user@host/path)
	// The @ must come before the first / to be a user
	var hostPortAndPath string
	atIndex := strings.Index(url, "@")
	slashIndex := strings.Index(url, "/")

	if atIndex != -1 && (slashIndex == -1 || atIndex < slashIndex) {
		existingUser := url[:atIndex]
		if existingUser != sshUser {
			return "", fmt.Errorf("conflicting SSH users: URL contains %q but credentials specify %q", existingUser, sshUser)
		}
		hostPortAndPath = url[atIndex+1:]
	} else {
		hostPortAndPath = url
	}

	// Find the path separator
	slashIndex = strings.Index(hostPortAndPath, "/")
	if slashIndex == -1 {
		return "", fmt.Errorf("invalid ssh:// URL format: %q (expected ssh://host/path)", "ssh://"+url)
	}

	hostPort := hostPortAndPath[:slashIndex]
	path := hostPortAndPath[slashIndex+1:]

	// Strip port from host if present (e.g., github.com:22 -> github.com)
	host := hostPort
	if colonIndex := strings.LastIndex(hostPort, ":"); colonIndex != -1 {
		host = hostPort[:colonIndex]
	}

	return sshUser + "@" + host + ":" + path, nil
}
