package credentials

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"path/filepath"
	"testing"

	gitssh "github.com/go-git/go-git/v6/plumbing/transport/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.uber.org/zap/zaptest"
	"golang.org/x/crypto/ssh"

	"github.com/go-git/go-git/v6/plumbing/transport"
)

func testPrivateKeyPEM(t *testing.T) string {
	t.Helper()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	blk, err := ssh.MarshalPrivateKey(priv, "")
	require.NoError(t, err)

	return string(pem.EncodeToMemory(blk))
}

// withoutKnownHosts points HOME and SSH_KNOWN_HOSTS at locations with no known_hosts
// file, reproducing a container environment.
func withoutKnownHosts(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SSH_KNOWN_HOSTS", filepath.Join(t.TempDir(), "does-not-exist"))
}

// TestInsecureHostKeyAuthClientConfig asserts that insecure_ignore_host_key yields a
// usable *ssh.ClientConfig on a host with no known_hosts file.
//
// go-git reads known_hosts solely to pre-populate HostKeyAlgorithms and returns an error
// when the file is absent, even though a HostKeyCallback has been supplied. Without
// HostKeyAlgorithms set, connecting fails with "unable to find any valid known_hosts
// file, set SSH_KNOWN_HOSTS env variable", making insecure_ignore_host_key inoperative.
func TestInsecureHostKeyAuthClientConfig(t *testing.T) {
	withoutKnownHosts(t)

	method, err := gitssh.NewPublicKeys("git", []byte(testPrivateKeyPEM(t)), "")
	require.NoError(t, err)
	// nolint:gosec
	method.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	cfg, err := insecureHostKeyAuth{PublicKeys: method}.ClientConfig(t.Context(), &transport.Request{})
	require.NoError(t, err, "ClientConfig must not require a known_hosts file")
	assert.NotNil(t, cfg.HostKeyCallback)
	assert.NotEmpty(t, cfg.HostKeyAlgorithms,
		"HostKeyAlgorithms must be pre-populated so go-git does not consult known_hosts")
}

// TestGitAuthenticationSSHInsecureIgnoreHostKey covers the same case through the public
// entry point, ensuring the wrapper is actually applied when the option is set.
func TestGitAuthenticationSSHInsecureIgnoreHostKey(t *testing.T) {
	withoutKnownHosts(t)

	c := &Credential{
		logger: zaptest.NewLogger(t),
		config: &config.CredentialConfig{
			Type: config.CredentialTypeSSH,
			SSH: &config.SSHAuthConfig{
				User:                  "git",
				PrivateKeyBytes:       testPrivateKeyPEM(t),
				InsecureIgnoreHostKey: true,
			},
		},
	}

	opt, err := c.GitAuthentication()
	require.NoError(t, err)
	require.NotNil(t, opt)
}
