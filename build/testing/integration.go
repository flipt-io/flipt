package testing

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v3"
	jjwt "github.com/go-jose/go-jose/v3/jwt"
	"github.com/google/uuid"
	"github.com/hashicorp/cap/jwt"
	"go.flipt.io/build/internal/dagger"
	"go.flipt.io/stew/config"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

var priv *rsa.PrivateKey

func init() {
	// Generate a key to sign JWTs with throughout most test cases.
	// It can be slow sometimes to generate a 4096-bit RSA key, so we only do it once.
	var err error
	priv, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		panic(err)
	}
}

var (
	protocolPorts = map[string]int{"http": 8080, "grpc": 9000}
	replacer      = strings.NewReplacer(" ", "-", "/", "-")
	sema          = make(chan struct{}, 6)

	// AllCases are the top-level filterable integration test cases.
	AllCases = map[string]testCaseFn{
		"authn/token":   authn("token"),
		"authn/k8s":     authn("k8s"),
		"authn/jwt":     authn("jwt"),
		"authz":         authz(),
		"envs":          envsAPI(""),
		"envs_with_dir": envsAPI("root"),
		"ofrep":         ofrepAPI(),
		"snapshot":      snapshotAPI(),
	}
)

type testConfig struct {
	name    string
	address string
	port    int
}

type testCaseFn func(_ context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error

func filterCases(caseNames ...string) (map[string]testCaseFn, error) {
	if len(caseNames) == 0 {
		return AllCases, nil
	}

	cases := map[string]testCaseFn{}
	for _, filter := range caseNames {
		if _, ok := AllCases[filter]; !ok {
			return nil, fmt.Errorf("unexpected test case filter: %q", filter)
		}

		cases[filter] = AllCases[filter]
	}

	return cases, nil
}

type IntegrationOptions func(*integrationOptions)

type integrationOptions struct {
	// The test cases to run. If empty, all test cases are run.
	cases []string
}

func WithTestCases(cases ...string) func(*integrationOptions) {
	return func(opts *integrationOptions) {
		opts.cases = cases
	}
}

func Integration(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, opts ...IntegrationOptions) error {
	var options integrationOptions

	for _, opt := range opts {
		opt(&options)
	}

	cases, err := filterCases(options.cases...)
	if err != nil {
		return err
	}

	var configs []testConfig

	for protocol, port := range protocolPorts {
		config := testConfig{
			name:    strings.ToUpper(protocol),
			address: fmt.Sprintf("%s://flipt:%d", protocol, port),
			port:    port,
		}

		configs = append(configs, config)
	}

	var g errgroup.Group

	for _, fn := range cases {
		for _, config := range configs {
			var (
				fn     = fn
				config = config
				flipt  = flipt
				base   = base
			)

			g.Go(take(func() error {
				flipt = flipt.
					WithEnvVariable("CI", os.Getenv("CI")).
					WithEnvVariable("FLIPT_LOG_LEVEL", "WARN").
					WithExposedPort(config.port)

				return fn(ctx, client, base, flipt, config)()
			}))
		}
	}

	return g.Wait()
}

func take(fn func() error) func() error {
	return func() error {
		// insert into semaphore channel to maintain
		// a max concurrency
		sema <- struct{}{}
		defer func() { <-sema }()

		return fn()
	}
}

const (
	testdataDir             = "build/testing/integration/testdata"
	environmentsTestdataDir = testdataDir + "/environments"
	namespacesTestdataDir   = testdataDir + "/namespaces"
)

func envsAPI(directory string) testCaseFn {
	return withGitea(func(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
		flipt = flipt.
			WithEnvVariable("FLIPT_LOG_LEVEL", "WARN").
			WithEnvVariable("FLIPT_ENVIRONMENTS_DEFAULT_STORAGE", "default").
			WithEnvVariable("FLIPT_ENVIRONMENTS_DEFAULT_DIRECTORY", directory).
			WithEnvVariable("FLIPT_ENVIRONMENTS_DEFAULT_DEFAULT", "true").
			WithEnvVariable("FLIPT_ENVIRONMENTS_PRODUCTION_STORAGE", "production").
			WithEnvVariable("FLIPT_ENVIRONMENTS_PRODUCTION_DIRECTORY", directory).
			WithEnvVariable("FLIPT_STORAGE_DEFAULT_REMOTE", "http://gitea:3000/root/features.git").
			WithEnvVariable("FLIPT_STORAGE_PRODUCTION_REMOTE", "http://gitea:3000/root/production.git").
			WithEnvVariable("FLIPT_STORAGE_DEFAULT_BRANCH", "main").
			WithEnvVariable("FLIPT_STORAGE_PRODUCTION_BRANCH", "main").
			WithEnvVariable("FLIPT_STORAGE_DEFAULT_CREDENTIALS", "default").
			WithEnvVariable("FLIPT_STORAGE_PRODUCTION_CREDENTIALS", "default").
			WithEnvVariable("FLIPT_CREDENTIALS_DEFAULT_TYPE", "basic").
			WithEnvVariable("FLIPT_CREDENTIALS_DEFAULT_BASIC_USERNAME", "root").
			WithEnvVariable("FLIPT_CREDENTIALS_DEFAULT_BASIC_PASSWORD", "password").
			WithEnvVariable("UNIQUE", uuid.New().String())

		return suite(ctx, "environments", base, flipt, conf)
	}, environmentsTestdataDir)
}

func authn(method string) testCaseFn {
	return withGitea(func(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
		flipt = flipt.
			WithEnvVariable("FLIPT_LOG_LEVEL", "WARN").
			WithEnvVariable("FLIPT_ENVIRONMENTS_DEFAULT_STORAGE", "default").
			WithEnvVariable("FLIPT_STORAGE_DEFAULT_REMOTE", "http://gitea:3000/root/features.git").
			WithEnvVariable("FLIPT_STORAGE_DEFAULT_BRANCH", "main").
			WithEnvVariable("FLIPT_STORAGE_DEFAULT_CREDENTIALS", "default").
			WithEnvVariable("FLIPT_CREDENTIALS_DEFAULT_TYPE", "basic").
			WithEnvVariable("FLIPT_CREDENTIALS_DEFAULT_BASIC_USERNAME", "root").
			WithEnvVariable("FLIPT_CREDENTIALS_DEFAULT_BASIC_PASSWORD", "password").
			WithEnvVariable("FLIPT_AUTHENTICATION_REQUIRED", "true").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_ENABLED", "true").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_STORAGE_TOKENS_BOOTSTRAP_CREDENTIAL", "s3cr3t").
			WithEnvVariable("UNIQUE", uuid.New().String())

		switch method {
		case "k8s":
			// K8s auth configuration
			flipt = flipt.
				WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_KUBERNETES_ENABLED", "true")

			var (
				priv []byte
				err  error
			)
			// run an OIDC server which exposes a JWKS url and returns
			// the associated private key bytes
			flipt, priv, err = serveOIDC(ctx, client, base, flipt)
			if err != nil {
				return func() error { return err }
			}

			// mount service account token into base on expected k8s sa token path
			base = base.WithNewFile("/var/run/secrets/flipt/k8s.pem", string(priv))
		case "jwt":
			// JWT auth configuration
			bytes, err := x509.MarshalPKIXPublicKey(priv.Public())
			if err != nil {
				return func() error { return err }
			}

			bytes = pem.EncodeToMemory(&pem.Block{
				Type:  "public key",
				Bytes: bytes,
			})

			flipt = flipt.
				WithNewFile("/etc/flipt/jwt.pem", string(bytes)).
				WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_JWT_ENABLED", "true").
				WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_JWT_PUBLIC_KEY_FILE", "/etc/flipt/jwt.pem").
				WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_JWT_VALIDATE_CLAIMS_ISSUER", "https://flipt.io")

			privBytes := pem.EncodeToMemory(&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(priv),
			})

			base = base.WithNewFile("/var/run/secrets/flipt/jwt.pem", string(privBytes))
		}

		return suite(ctx, "authn/"+method, base, flipt, conf)
	}, namespacesTestdataDir)
}

func authz() testCaseFn {
	return withAuthz(withGitea(func(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
		flipt = flipt.
			WithEnvVariable("FLIPT_LOG_LEVEL", "WARN").
			WithEnvVariable("FLIPT_ENVIRONMENTS_DEFAULT_STORAGE", "default").
			WithEnvVariable("FLIPT_STORAGE_DEFAULT_REMOTE", "http://gitea:3000/root/features.git").
			WithEnvVariable("FLIPT_STORAGE_DEFAULT_BRANCH", "main").
			WithEnvVariable("FLIPT_STORAGE_DEFAULT_CREDENTIALS", "default").
			WithEnvVariable("FLIPT_CREDENTIALS_DEFAULT_TYPE", "basic").
			WithEnvVariable("FLIPT_CREDENTIALS_DEFAULT_BASIC_USERNAME", "root").
			WithEnvVariable("FLIPT_CREDENTIALS_DEFAULT_BASIC_PASSWORD", "password").
			WithEnvVariable("FLIPT_AUTHENTICATION_REQUIRED", "true").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_ENABLED", "true").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_STORAGE_TOKENS_BOOTSTRAP_CREDENTIAL", "s3cr3t").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_STORAGE_TOKENS_ADMIN_CREDENTIAL", "admin123").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_STORAGE_TOKENS_ADMIN_METADATA_ROLE", "admin").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_STORAGE_TOKENS_EDITOR_CREDENTIAL", "editor456").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_STORAGE_TOKENS_EDITOR_METADATA_ROLE", "editor").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_STORAGE_TOKENS_VIEWER_CREDENTIAL", "viewer789").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_STORAGE_TOKENS_VIEWER_METADATA_ROLE", "viewer").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_STORAGE_TOKENS_DEFAULT_VIEWER_CREDENTIAL", "default_viewer1111").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_STORAGE_TOKENS_DEFAULT_VIEWER_METADATA_ROLE", "default_viewer").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_STORAGE_TOKENS_ALTERNATIVE_VIEWER_CREDENTIAL", "alternative_viewer2222").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_STORAGE_TOKENS_ALTERNATIVE_VIEWER_METADATA_ROLE", "alternative_viewer").
			WithEnvVariable("FLIPT_AUTHORIZATION_REQUIRED", "true").
			WithEnvVariable("UNIQUE", uuid.New().String())

		// JWT auth configuration
		bytes, err := x509.MarshalPKIXPublicKey(priv.Public())
		if err != nil {
			return func() error { return err }
		}

		bytes = pem.EncodeToMemory(&pem.Block{
			Type:  "public key",
			Bytes: bytes,
		})

		flipt = flipt.
			WithNewFile("/etc/flipt/jwt.pem", string(bytes)).
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_JWT_ENABLED", "true").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_JWT_PUBLIC_KEY_FILE", "/etc/flipt/jwt.pem").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_JWT_VALIDATE_CLAIMS_ISSUER", "https://flipt.io")

		privBytes := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		})

		base = base.WithNewFile("/var/run/secrets/flipt/jwt.pem", string(privBytes))

		return suite(ctx, "authz", base, flipt, conf)
	}, namespacesTestdataDir))
}

func snapshotAPI() testCaseFn {
	return withGitea(func(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
		flipt = flipt.
			WithEnvVariable("FLIPT_LOG_LEVEL", "DEBUG").
			WithEnvVariable("FLIPT_ENVIRONMENTS_DEFAULT_STORAGE", "default").
			WithEnvVariable("FLIPT_STORAGE_DEFAULT_REMOTE", "http://gitea:3000/root/features.git").
			WithEnvVariable("FLIPT_STORAGE_DEFAULT_BRANCH", "main").
			WithEnvVariable("FLIPT_STORAGE_DEFAULT_CREDENTIALS", "default").
			WithEnvVariable("FLIPT_CREDENTIALS_DEFAULT_TYPE", "basic").
			WithEnvVariable("FLIPT_CREDENTIALS_DEFAULT_BASIC_USERNAME", "root").
			WithEnvVariable("FLIPT_CREDENTIALS_DEFAULT_BASIC_PASSWORD", "password").
			WithEnvVariable("UNIQUE", uuid.New().String())

		return suite(ctx, "snapshot", base, flipt, conf)
	}, namespacesTestdataDir)
}

func ofrepAPI() testCaseFn {
	return withGitea(func(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
		flipt = flipt.
			WithEnvVariable("FLIPT_LOG_LEVEL", "WARN").
			WithEnvVariable("FLIPT_ENVIRONMENTS_DEFAULT_STORAGE", "default").
			WithEnvVariable("FLIPT_STORAGE_DEFAULT_REMOTE", "http://gitea:3000/root/features.git").
			WithEnvVariable("FLIPT_STORAGE_DEFAULT_BRANCH", "main").
			WithEnvVariable("FLIPT_STORAGE_DEFAULT_CREDENTIALS", "default").
			WithEnvVariable("FLIPT_CREDENTIALS_DEFAULT_TYPE", "basic").
			WithEnvVariable("FLIPT_CREDENTIALS_DEFAULT_BASIC_USERNAME", "root").
			WithEnvVariable("FLIPT_CREDENTIALS_DEFAULT_BASIC_PASSWORD", "password").
			WithEnvVariable("UNIQUE", uuid.New().String())

		return suite(ctx, "ofrep", base, flipt, conf)
	}, namespacesTestdataDir)
}

func withGitea(fn testCaseFn, dataDir string) testCaseFn {
	return func(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
		gitea := client.Container().
			From("gitea/gitea:1.23.3").
			WithExposedPort(3000).
			WithEnvVariable("UNIQUE", time.Now().String()).
			AsService()

		stew := config.Config{
			URL: "http://gitea:3000",
			Admin: struct {
				Username string "json:\"username\""
				Email    string "json:\"email\""
				Password string "json:\"password\""
			}{
				Username: "root",
				Password: "password",
				Email:    "dev@flipt.io",
			},
			Repositories: []config.Repository{
				{
					Name: "features",
					Contents: []config.Content{
						{
							// we always at-least create "main"
							Branch:  "main",
							Path:    "/work/default",
							Message: "feat: add directory contents",
						},
					},
				},
				{
					Name: "production",
					Contents: []config.Content{
						{
							Branch:  "main",
							Path:    "/work/production",
							Message: "feat: add directory contents",
						},
					},
				},
			},
		}

		contents, err := yaml.Marshal(&stew)
		if err != nil {
			return func() error { return err }
		}

		_, err = client.Container().
			From("ghcr.io/flipt-io/stew:latest").
			WithWorkdir("/work").
			WithDirectory("/work/default", base.Directory(dataDir)).
			WithDirectory("/work/production", base.Directory(dataDir)).
			WithNewFile("/etc/stew/config.yml", string(contents)).
			WithServiceBinding("gitea", gitea).
			WithExec([]string{"/usr/local/bin/stew", "-config", "/etc/stew/config.yml"}).
			Sync(ctx)
		if err != nil {
			return func() error { return err }
		}

		return fn(
			ctx,
			client,
			base,
			flipt.WithServiceBinding("gitea", gitea),
			conf,
		)
	}
}

func withAuthz(fn testCaseFn) testCaseFn {
	return func(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
		var (
			policyPath = "/etc/flipt/authz/policy.rego"
			policyData = "/etc/flipt/authz/data.json"
		)

		policy := `
package flipt.authz.v2

import rego.v1

default allow := false

# Allow if any matching rule grants permission
allow if {
    some rule in has_rules
    allow_rule(rule)
}

# Helper to check if a rule allows the request
allow_rule(rule) if {
    # For admin role, grant all permissions
    input.authentication.metadata["io.flipt.auth.role"] == "admin"
}

allow_rule(rule) if {
    # For admin role, grant all permissions
    input.authentication.metadata["role"] == "admin"
}

# Helper to check if a rule allows the request with specific permissions
allow_rule(rule) if {
    # Check scope
    permit_string(rule.scope, input.request.scope)
    # Check environment if specified
    check_environment(rule)
    # Check namespace if specified
    check_namespace(rule)
    # Check actions
    permit_slice(rule.actions, input.request.action)
}

# Check if environment matches or is not specified
check_environment(rule) if {
    not rule.environment
}

check_environment(rule) if {
    rule.environment
    permit_string(rule.environment, input.request.environment)
}

# Check if namespace matches or is not specified
check_namespace(rule) if {
    not rule.namespace
}

check_namespace(rule) if {
    rule.namespace
    permit_string(rule.namespace, input.request.namespace)
}

has_rules contains rules if {
    some role in data.roles
    role.name == input.authentication.metadata["io.flipt.auth.role"]
    rules := role.rules[_]
}

has_rules contains rules if {
    some role in data.roles
    role.name == input.authentication.metadata["role"]
    rules := role.rules[_]
}

permit_string(allowed, _) if {
    allowed == "*"
}

permit_string(allowed, requested) if {
    allowed == requested
}

permit_slice(allowed, _) if {
    allowed[_] = "*"
}

permit_slice(allowed, requested) if {
    allowed[_] = requested
}
`

		data := `
{
    "version": "0.1.0",
    "roles": [
        {
            "name": "admin",
            "rules": [
                {
                    "scope": "*",
                    "actions": ["*"]
                }
            ]
        },
        {
            "name": "editor",
            "rules": [
                {
                    "scope": "*",
                    "actions": ["read"]
                },
                {
                    "scope": "namespace",
                    "actions": ["create", "read", "update", "delete"]
                }
            ]
        },
        {
            "name": "viewer",
            "rules": [
                {
                    "scope": "*",
                    "actions": ["read"]
                }
            ]
        },
        {
            "name": "default_viewer",
            "rules": [
                {
                    "scope": "*",
                    "actions": ["read"],
                    "namespace": "default"
                }
            ]
        },
        {
            "name": "alternative_viewer",
            "rules": [
                {
                    "scope": "*",
                    "actions": ["read"],
                    "namespace": "alternative"
                }
            ]
        }
    ]
}
`

		return fn(ctx, client, base, flipt.
			WithEnvVariable("FLIPT_AUTHORIZATION_BACKEND", "local").
			WithEnvVariable("FLIPT_AUTHORIZATION_LOCAL_POLICY_PATH", policyPath).
			WithNewFile(policyPath, string(policy)).
			WithEnvVariable("FLIPT_AUTHORIZATION_LOCAL_DATA_PATH", policyData).
			WithNewFile(policyData, string(data)),
			conf,
		)
	}
}

func suite(ctx context.Context, dir string, base, flipt *dagger.Container, conf testConfig) func() error {
	return func() (err error) {
		flags := []string{"--flipt-addr", conf.address}

		_, err = base.
			WithWorkdir(path.Join("build/testing/integration", dir)).
			WithEnvVariable("UNIQUE", uuid.New().String()).
			WithServiceBinding("flipt", flipt.AsService()).
			WithExec([]string{"sh", "-c", fmt.Sprintf("go test -v -timeout=1m -race %s .", strings.Join(flags, " "))}).
			Sync(ctx)

		return err
	}
}

func signJWT(key crypto.PrivateKey, claims interface{}) string {
	sig, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.SignatureAlgorithm(string(jwt.RS256)), Key: key},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		panic(err)
	}

	raw, err := jjwt.Signed(sig).
		Claims(claims).
		CompactSerialize()
	if err != nil {
		panic(err)
	}

	return raw
}

// serveOIDC runs a mini OIDC-style key provider and mounts it as a service onto Flipt.
// This provider is designed to mimic how kubernetes exposes JWKS endpoints for its service account tokens.
// The function creates signing keys and TLS CA certificates which is shares with the provider and
// with Flipt itself. This is to facilitate Flipt using the custom CA to authenticate the provider.
// The function generates two JWTs, one for Flipt to identify itself and one which is returned to the caller.
// The caller can use this as the service account token identity to be mounted into the container with the
// client used for running the test and authenticating with Flipt.
func serveOIDC(_ context.Context, _ *dagger.Client, base, flipt *dagger.Container) (*dagger.Container, []byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	rsaSigningKey := &bytes.Buffer{}
	if err := pem.Encode(rsaSigningKey, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}); err != nil {
		return nil, nil, err
	}

	// generate a SA style JWT for identifying the Flipt service
	fliptSAToken := signJWT(priv, map[string]any{
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iss": "https://discover.svc",
		"kubernetes.io": map[string]any{
			"namespace": "flipt",
			"pod": map[string]any{
				"name": "flipt-7d26f049-kdurb",
				"uid":  "bd8299f9-c50f-4b76-af33-9d8e3ef2b850",
			},
			"serviceaccount": map[string]any{
				"name": "flipt",
				"uid":  "4f18914e-f276-44b2-aebd-27db1d8f8def",
			},
		},
	})

	// generate a CA certificate to share between Flipt and the mini OIDC server
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization:  []string{"Flipt, INC."},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"North Carolina"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		DNSNames:              []string{"discover.svc"},
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	var caCert bytes.Buffer
	if err := pem.Encode(&caCert, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}); err != nil {
		return nil, nil, err
	}

	var caPrivKeyPEM bytes.Buffer
	_ = pem.Encode(&caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	return flipt.
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_KUBERNETES_DISCOVERY_URL", "https://discover.svc").
			WithServiceBinding("discover.svc", base.
				WithNewFile("/server.crt", caCert.String()).
				WithNewFile("/server.key", caPrivKeyPEM.String()).
				WithNewFile("/priv.pem", rsaSigningKey.String()).
				WithExposedPort(443).
				WithDefaultArgs([]string{
					"sh",
					"-c",
					"go run ./build/internal/cmd/discover/... --private-key /priv.pem",
				}).
				AsService()).
			WithNewFile("/var/run/secrets/kubernetes.io/serviceaccount/token", fliptSAToken).
			WithNewFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt", caCert.String()),
		rsaSigningKey.Bytes(), nil
}
