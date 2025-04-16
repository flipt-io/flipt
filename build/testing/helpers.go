package testing

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
)

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
