package testing

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.flipt.io/build/internal/dagger"
)

func flipt(args ...string) []string {
	return append([]string{"/flipt"}, args...)
}

func sh(cmd string) []string {
	return []string{"sh", "-c", cmd}
}

type stringAssertion interface {
	assert(string) error
}

type assertConf struct {
	success bool
	stdout  []stringAssertion
	stderr  []stringAssertion
}

type assertOption func(*assertConf)

func fails(c *assertConf) { c.success = false }

func stdout(a stringAssertion) assertOption {
	return func(c *assertConf) {
		c.stdout = append(c.stdout, a)
	}
}

func stderr(a stringAssertion) assertOption {
	return func(c *assertConf) {
		c.stderr = append(c.stderr, a)
	}
}

type equals string

func (e equals) assert(t string) error {
	if diff := cmp.Diff(string(e), t); diff != "" {
		return fmt.Errorf("unexpected output: diff (-/+):\n%s", diff)
	}

	return nil
}

type contains string

func (c contains) assert(t string) error {
	if !strings.Contains(t, string(c)) {
		return fmt.Errorf("unexpected output: %q does not contain %q", t, c)
	}

	return nil
}

type matches string

func (m matches) assert(t string) error {
	r := regexp.MustCompile(string(m))
	if !r.MatchString(t) {
		return fmt.Errorf("unexpected output %q does not match %q", t, m)
	}

	return nil
}

func assertExec(ctx context.Context, flipt *dagger.Container, args []string, opts ...assertOption) (*dagger.Container, error) {
	conf := assertConf{success: true}
	for _, opt := range opts {
		opt(&conf)
	}

	container, err := flipt.WithExec(args).Sync(ctx)

	var (
		execStdout = container.Stdout
		execStderr = container.Stderr
	)

	if err != nil {
		if conf.success {
			return nil, fmt.Errorf("unexpected error running flipt %q: %w", args, err)
		}

		var eerr *dagger.ExecError
		// get stdout and stderr from exec error instead
		if errors.As(err, &eerr) {
			execStdout = func(context.Context) (string, error) {
				return eerr.Stdout, nil
			}

			execStderr = func(context.Context) (string, error) {
				return eerr.Stderr, nil
			}
		}
	}

	if err == nil && !conf.success {
		return nil, fmt.Errorf("expected error running flipt %q: found success", args)
	}

	for _, a := range conf.stdout {
		stdout, err := execStdout(ctx)
		if err != nil {
			return nil, err
		}

		if err := a.assert(string(stdout)); err != nil {
			return nil, err
		}
	}

	for _, a := range conf.stderr {
		stderr, err := execStderr(ctx)
		if err != nil {
			return nil, err
		}

		if err := a.assert(string(stderr)); err != nil {
			return nil, err
		}
	}

	return container, nil
}

// generateTLSCert generates a TLS certificate and private key.
func generateTLSCert(dnsname ...string) (keyBytes, certBytes []byte, err error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, err
	}
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		IsCA:         true,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		DNSNames:     dnsname,
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	bytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	certBytes = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: bytes,
	})

	keyBytes = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	return
}
