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
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"path"
	"strings"
	"time"

	"dagger.io/dagger"
	"github.com/containerd/containerd/platforms"
	"github.com/go-jose/go-jose/v3"
	jjwt "github.com/go-jose/go-jose/v3/jwt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/hashicorp/cap/jwt"
	"go.flipt.io/stew/config"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

const bootstrapToken = "s3cr3t"

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
		"api/sqlite":    withSQLite(api),
		"api/libsql":    withLibSQL(api),
		"api/postgres":  withPostgres(api),
		"api/mysql":     withMySQL(api),
		"api/cockroach": withCockroach(api),
		"api/cache":     cache,
		"api/cachetls":  cacheWithTLS,
		"fs/git":        git,
		"fs/local":      local,
		"fs/s3":         s3,
		"fs/oci":        oci,
		"fs/azblob":     azblob,
		"fs/gcs":        gcs,
		"import/export": importExport,
		"auth/sqlite":   auth,
	}
)

type testConfig struct {
	name       string
	address    string
	port       int
	references bool
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

func Integration(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, caseNames ...string) error {
	cases, err := filterCases(caseNames...)
	if err != nil {
		return err
	}

	logs := client.CacheVolume(fmt.Sprintf("logs-%s", uuid.New()))
	_, err = flipt.WithUser("root").
		WithMountedCache("/logs", logs).
		WithExec([]string{"chown", "flipt:flipt", "/logs"}).
		Sync(ctx)
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

	for caseName, fn := range cases {
		for _, config := range configs {
			var (
				fn     = fn
				config = config
				flipt  = flipt
				base   = base
			)

			g.Go(take(func() error {
				{
					// Static token auth configuration
					flipt = flipt.
						WithEnvVariable("FLIPT_AUTHENTICATION_REQUIRED", "true").
						WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_ENABLED", "true").
						WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_BOOTSTRAP_TOKEN", bootstrapToken)
				}
				{
					// K8s auth configuration
					flipt = flipt.
						WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_KUBERNETES_ENABLED", "true")

					var saToken string
					// run an OIDC server which exposes a JWKS url using a private key we own
					// and generate a JWT to act as our SA token
					flipt, saToken, err = serveOIDC(ctx, client, base, flipt)
					if err != nil {
						return err
					}

					// mount service account token into base on expected k8s sa token path
					base = base.WithNewFile("/var/run/secrets/kubernetes.io/serviceaccount/token", dagger.ContainerWithNewFileOpts{
						Contents: saToken,
					})
				}
				{
					// JWT auth configuration
					bytes, err := x509.MarshalPKIXPublicKey(priv.Public())
					if err != nil {
						return err
					}

					bytes = pem.EncodeToMemory(&pem.Block{
						Type:  "public key",
						Bytes: bytes,
					})

					flipt = flipt.
						WithNewFile("/etc/flipt/jwt.pem", dagger.ContainerWithNewFileOpts{Contents: string(bytes)}).
						WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_JWT_ENABLED", "true").
						WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_JWT_PUBLIC_KEY_FILE", "/etc/flipt/jwt.pem").
						WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_JWT_VALIDATE_CLAIMS_ISSUER", "https://flipt.io")

					privBytes := pem.EncodeToMemory(&pem.Block{
						Type:  "RSA PRIVATE KEY",
						Bytes: x509.MarshalPKCS1PrivateKey(priv),
					})

					base = base.WithNewFile("/var/run/secrets/flipt/private.pem", dagger.ContainerWithNewFileOpts{
						Contents: string(privBytes),
					})
				}

				name := strings.ToLower(replacer.Replace(fmt.Sprintf("flipt-test-%s-config-%s", caseName, config.name)))
				flipt = flipt.
					WithEnvVariable("CI", os.Getenv("CI")).
					WithEnvVariable("FLIPT_LOG_LEVEL", "debug").
					WithEnvVariable("FLIPT_LOG_FILE", fmt.Sprintf("/var/opt/flipt/logs/%s.log", name)).
					WithMountedCache("/var/opt/flipt/logs", logs).
					WithExposedPort(config.port)
				return fn(ctx, client, base.Pipeline(name), flipt, config)()
			}))
		}
	}

	err = g.Wait()

	if _, lerr := client.Container().From("alpine:3.16").
		WithMountedCache("/logs", logs).
		WithExec([]string{"cp", "-r", "/logs", "/out"}).
		Directory("/out").Export(ctx, "build/logs"); lerr != nil {
		log.Println("Error copying logs", lerr)
	}

	return err
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

func api(ctx context.Context, _ *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
	return suite(ctx, "api", base,
		// create unique instance for test case
		flipt.WithEnvVariable("UNIQUE", uuid.New().String()).WithExec(nil), conf)
}

func withSQLite(fn testCaseFn) testCaseFn {
	return fn
}

func withLibSQL(fn testCaseFn) testCaseFn {
	return func(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
		return fn(ctx, client, base, flipt.WithEnvVariable("FLIPT_DB_URL", "libsql://file:/etc/flipt/flipt.db"), conf)
	}
}

func withPostgres(fn testCaseFn) testCaseFn {
	return func(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
		return fn(ctx, client, base, flipt.
			WithEnvVariable("FLIPT_DB_URL", "postgres://postgres:password@postgres:5432?sslmode=disable").
			WithServiceBinding("postgres", client.Container().
				From("postgres:alpine").
				WithEnvVariable("POSTGRES_PASSWORD", "password").
				WithExposedPort(5432).
				WithEnvVariable("UNIQUE", uuid.New().String()).
				WithExec(nil).
				AsService()),
			conf,
		)
	}
}

func withMySQL(fn testCaseFn) testCaseFn {
	return func(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
		return fn(ctx, client, base, flipt.
			WithEnvVariable(
				"FLIPT_DB_URL",
				"mysql://flipt:password@mysql:3306/flipt_test?multiStatements=true",
			).
			WithServiceBinding("mysql", client.Container().
				From("mysql:8").
				WithEnvVariable("MYSQL_USER", "flipt").
				WithEnvVariable("MYSQL_PASSWORD", "password").
				WithEnvVariable("MYSQL_DATABASE", "flipt_test").
				WithEnvVariable("MYSQL_ALLOW_EMPTY_PASSWORD", "true").
				WithEnvVariable("UNIQUE", uuid.New().String()).
				WithExposedPort(3306).
				WithExec(nil).
				AsService()),
			conf,
		)
	}
}

func withCockroach(fn testCaseFn) testCaseFn {
	return func(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
		return fn(ctx, client, base, flipt.
			WithEnvVariable("FLIPT_DB_URL", "cockroachdb://root@cockroach:26257/defaultdb?sslmode=disable").
			WithServiceBinding("cockroach", client.Container().
				From("cockroachdb/cockroach:latest-v21.2").
				WithEnvVariable("COCKROACH_USER", "root").
				WithEnvVariable("COCKROACH_DATABASE", "defaultdb").
				WithEnvVariable("UNIQUE", uuid.New().String()).
				WithExposedPort(26257).
				WithExec([]string{"start-single-node", "--single-node", "--insecure", "--store=type=mem,size=0.7Gb", "--accept-sql-without-tls", "--logtostderr=ERROR"}).
				AsService()),
			conf,
		)

	}
}

func cache(ctx context.Context, _ *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
	flipt = flipt.
		WithEnvVariable("FLIPT_LOG_LEVEL", "DEBUG").
		WithEnvVariable("FLIPT_CACHE_ENABLED", "true").
		WithEnvVariable("FLIPT_CACHE_TTL", "1s")

	return suite(ctx, "api", base, flipt.WithExec(nil), conf)
}

func cacheWithTLS(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return func() error { return err }
	}
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		IsCA:         true,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		DNSNames:     []string{"redis"},
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return func() error { return err }
	}
	bytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return func() error { return err }
	}
	crtBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: bytes,
	})

	keyBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	redis := client.Container().
		From("redis:alpine").
		WithExposedPort(6379).
		WithNewFile("/opt/tls/key", dagger.ContainerWithNewFileOpts{Contents: string(keyBytes)}).
		WithNewFile("/opt/tls/crt", dagger.ContainerWithNewFileOpts{Contents: string(crtBytes)}).
		WithExec([]string{
			"redis-server", "--tls-port", "6379", "--port", "0",
			"--tls-key-file", "/opt/tls/key", "--tls-cert-file",
			"/opt/tls/crt", "--tls-ca-cert-file", "/opt/tls/crt",
			"--tls-auth-clients", "no"}).
		AsService()

	flipt = flipt.
		WithEnvVariable("FLIPT_LOG_LEVEL", "DEBUG").
		WithEnvVariable("FLIPT_CACHE_ENABLED", "true").
		WithEnvVariable("FLIPT_CACHE_TTL", "1s").
		WithEnvVariable("FLIPT_CACHE_BACKEND", "redis").
		WithEnvVariable("FLIPT_CACHE_REDIS_REQUIRE_TLS", "true").
		WithEnvVariable("FLIPT_CACHE_REDIS_HOST", "redis").
		WithEnvVariable("FLIPT_CACHE_REDIS_CA_CERT_PATH", "/opt/tls/crt").
		WithNewFile("/opt/tls/crt", dagger.ContainerWithNewFileOpts{Contents: string(crtBytes)}).
		WithServiceBinding("redis", redis)
	return suite(ctx, "api", base, flipt.WithExec(nil), conf)
}

const (
	rootTestdataDir           = "build/testing/integration/readonly/testdata"
	singleRevisionTestdataDir = rootTestdataDir + "/main"
)

func local(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
	flipt = flipt.
		WithDirectory("/tmp/testdata", base.Directory(singleRevisionTestdataDir)).
		WithEnvVariable("FLIPT_LOG_LEVEL", "DEBUG").
		WithEnvVariable("FLIPT_STORAGE_TYPE", "local").
		WithEnvVariable("FLIPT_STORAGE_LOCAL_PATH", "/tmp/testdata").
		WithEnvVariable("UNIQUE", uuid.New().String())

	return suite(ctx, "readonly", base, flipt.WithExec(nil), conf)
}

func git(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
	gitea := client.Container().
		From("gitea/gitea:1.21.1").
		WithExposedPort(3000).
		WithExec(nil)

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
						Branch:  "main",
						Path:    "/work/base/main",
						Message: "feat: add main directory contents",
					},
					{
						Branch:  "alternate",
						Path:    "/work/base/alternate",
						Message: "feat: add alternate directory contents",
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
		WithDirectory("/work/base", base.Directory(rootTestdataDir)).
		WithNewFile("/etc/stew/config.yml", dagger.ContainerWithNewFileOpts{
			Contents: string(contents),
		}).
		WithServiceBinding("gitea", gitea.AsService()).
		WithExec(nil).
		Sync(ctx)
	if err != nil {
		return func() error { return err }
	}

	flipt = flipt.
		WithServiceBinding("gitea", gitea.AsService()).
		WithEnvVariable("FLIPT_LOG_LEVEL", "DEBUG").
		WithEnvVariable("FLIPT_STORAGE_TYPE", "git").
		WithEnvVariable("FLIPT_STORAGE_GIT_REPOSITORY", "http://gitea:3000/root/features.git").
		WithEnvVariable("FLIPT_STORAGE_GIT_AUTHENTICATION_BASIC_USERNAME", "root").
		WithEnvVariable("FLIPT_STORAGE_GIT_AUTHENTICATION_BASIC_PASSWORD", "password").
		WithEnvVariable("UNIQUE", uuid.New().String())

	// Git backend supports arbitrary references
	conf.references = true

	return suite(ctx, "readonly", base, flipt.WithExec(nil), conf)
}

func s3(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
	minio := client.Container().
		From("quay.io/minio/minio:latest").
		WithEnvVariable("UNIQUE", uuid.New().String()).
		WithExposedPort(9009).
		WithEnvVariable("MINIO_ROOT_USER", "user").
		WithEnvVariable("MINIO_ROOT_PASSWORD", "password").
		WithEnvVariable("MINIO_BROWSER", "off").
		WithExec([]string{"server", "/data", "--address", ":9009", "--quiet"}).
		AsService()

	_, err := base.
		WithServiceBinding("minio", minio).
		WithEnvVariable("AWS_ACCESS_KEY_ID", "user").
		WithEnvVariable("AWS_SECRET_ACCESS_KEY", "password").
		WithExec([]string{"go", "run", "./build/internal/cmd/minio/...", "-minio-url", "http://minio:9009", "-testdata-dir", singleRevisionTestdataDir}).
		Sync(ctx)
	if err != nil {
		return func() error { return err }
	}

	flipt = flipt.
		WithServiceBinding("minio", minio).
		WithEnvVariable("FLIPT_LOG_LEVEL", "DEBUG").
		WithEnvVariable("AWS_ACCESS_KEY_ID", "user").
		WithEnvVariable("AWS_SECRET_ACCESS_KEY", "password").
		WithEnvVariable("FLIPT_STORAGE_TYPE", "object").
		WithEnvVariable("FLIPT_STORAGE_OBJECT_TYPE", "s3").
		WithEnvVariable("FLIPT_STORAGE_OBJECT_S3_ENDPOINT", "http://minio:9009").
		WithEnvVariable("FLIPT_STORAGE_OBJECT_S3_BUCKET", "testdata").
		WithEnvVariable("UNIQUE", uuid.New().String())

	return suite(ctx, "readonly", base, flipt.WithExec(nil), conf)
}

func oci(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
	platform, err := client.DefaultPlatform(ctx)
	if err != nil {
		return func() error { return err }
	}

	var (
		// username == "username" password == "password"
		htpasswd  = "username:$2y$05$0krVCN7KfnmV5MwdD6Z7CuFuFnmbqP8.14iEV/nhNLM4V3VFF7NVK"
		zotConfig = `{
    "storage": {
        "rootDirectory": "/var/lib/registry"
    },
    "http": {
        "address": "0.0.0.0",
        "port": "5000",
        "auth": {
            "htpasswd": {
                "path": "/etc/zot/htpasswd"
            }
        }
    },
    "log": {
        "level": "debug"
    }
}`
	)
	// switch out zot images based on host platform
	// and push to remote name
	zot := client.Container().
		From(fmt.Sprintf("ghcr.io/project-zot/zot-linux-%s:latest",
			platforms.MustParse(string(platform)).Architecture)).
		WithExposedPort(5000).
		WithNewFile("/etc/zot/htpasswd", dagger.ContainerWithNewFileOpts{Contents: htpasswd}).
		WithNewFile("/etc/zot/config.json", dagger.ContainerWithNewFileOpts{Contents: zotConfig})

	if _, err := flipt.
		WithDirectory("/tmp/testdata", base.Directory(singleRevisionTestdataDir)).
		WithWorkdir("/tmp/testdata").
		WithServiceBinding("zot", zot.AsService()).
		WithEnvVariable("FLIPT_STORAGE_OCI_AUTHENTICATION_USERNAME", "username").
		WithEnvVariable("FLIPT_STORAGE_OCI_AUTHENTICATION_PASSWORD", "password").
		WithExec([]string{"/flipt", "bundle", "build", "readonly:latest"}).
		WithExec([]string{"/flipt", "bundle", "push", "readonly:latest", "http://zot:5000/readonly:latest"}).
		Sync(ctx); err != nil {
		return func() error {
			return err
		}
	}

	flipt = flipt.
		WithServiceBinding("zot", zot.AsService()).
		WithEnvVariable("FLIPT_LOG_LEVEL", "DEBUG").
		WithEnvVariable("FLIPT_STORAGE_TYPE", "oci").
		WithEnvVariable("FLIPT_STORAGE_OCI_REPOSITORY", "http://zot:5000/readonly:latest").
		WithEnvVariable("FLIPT_STORAGE_OCI_AUTHENTICATION_USERNAME", "username").
		WithEnvVariable("FLIPT_STORAGE_OCI_AUTHENTICATION_PASSWORD", "password").
		WithEnvVariable("UNIQUE", uuid.New().String())

	return suite(ctx, "readonly", base, flipt.WithExec(nil), conf)
}

func importInto(ctx context.Context, base, flipt, fliptToTest *dagger.Container, flags ...string) error {
	for _, ns := range []string{"default", "production"} {
		seed := base.File(path.Join(singleRevisionTestdataDir, ns+".yaml"))

		var (
			importCmd = append([]string{"/flipt", "import"}, append(flags, "import.yaml")...)
		)
		// use target flipt binary to invoke import
		_, err := flipt.
			WithEnvVariable("UNIQUE", uuid.New().String()).
			// copy testdata import yaml from base
			WithFile("import.yaml", seed).
			WithServiceBinding("flipt", fliptToTest.AsService()).
			// it appears it takes a little while for Flipt to come online
			// For the go tests they have to compile and that seems to be enough
			// time for the target Flipt to come up.
			// However, in this case the flipt binary is prebuilt and needs a little sleep.
			WithExec([]string{"sh", "-c", fmt.Sprintf("sleep 2 && %s", strings.Join(importCmd, " "))}).
			Sync(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func importExport(ctx context.Context, _ *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
	return func() error {
		// import testdata before running readonly suite
		flags := []string{"--address", conf.address, "--token", bootstrapToken}

		// create unique instance for test case
		fliptToTest := flipt.
			WithEnvVariable("UNIQUE", uuid.New().String()).
			WithExec(nil)

		if err := importInto(ctx, base, flipt, fliptToTest, flags...); err != nil {
			return err
		}

		// run readonly suite against imported Flipt instance
		if err := suite(ctx, "readonly", base, fliptToTest, conf)(); err != nil {
			return err
		}

		for _, ns := range []string{"default", "production"} {
			seed := base.File(path.Join(singleRevisionTestdataDir, ns+".yaml"))
			expected, err := seed.Contents(ctx)
			if err != nil {
				return err
			}

			if ns == "default" {
				// replace namespace in expected yaml
				expected = strings.ReplaceAll(expected, "version: \"1.2\"\n", fmt.Sprintf("version: \"1.2\"\nnamespace: %s\n", ns))
			}

			if ns != "default" {
				flags = append(flags, "--namespaces", ns)
			}

			// use target flipt binary to invoke import
			generated, err := flipt.
				WithEnvVariable("UNIQUE", uuid.New().String()).
				WithServiceBinding("flipt", fliptToTest.AsService()).
				WithExec(append([]string{"/flipt", "export", "-o", "/tmp/output.yaml"}, flags...)).
				File("/tmp/output.yaml").
				Contents(ctx)
			if err != nil {
				return err
			}

			// remove line that starts with comment character '#' and newline after
			generated = generated[strings.Index(generated, "\n")+2:]

			diff := cmp.Diff(expected, generated)
			if diff != "" {
				fmt.Printf("Unexpected difference in %q exported output: \n", conf.name)
				fmt.Println(diff)
				return errors.New("exported yaml did not match")
			}
		}

		return nil
	}
}

func auth(ctx context.Context, _ *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
	// create unique instance for test case
	fliptToTest := flipt.WithEnvVariable("UNIQUE", uuid.New().String()).WithExec(nil)
	// import state into instance before running test
	if err := importInto(ctx, base, flipt, fliptToTest, "--address", conf.address, "--token", bootstrapToken); err != nil {
		return func() error { return err }
	}

	return suite(ctx, "auth", base, fliptToTest, conf)
}

func suite(ctx context.Context, dir string, base, flipt *dagger.Container, conf testConfig) func() error {
	return func() (err error) {
		flags := []string{"--flipt-addr", conf.address, "--flipt-token", bootstrapToken}
		if conf.references {
			flags = append(flags, "--flipt-supports-references")
		}

		_, err = base.
			WithWorkdir(path.Join("build/testing/integration", dir)).
			WithEnvVariable("UNIQUE", uuid.New().String()).
			WithServiceBinding("flipt", flipt.AsService()).
			WithExec([]string{"sh", "-c", fmt.Sprintf("go test -v -timeout=1m -race %s .", strings.Join(flags, " "))}).
			Sync(ctx)

		return err
	}
}

// azurite simulates the Azure blob service
func azblob(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
	azurite := client.Container().
		From("mcr.microsoft.com/azure-storage/azurite").
		WithEnvVariable("UNIQUE", uuid.New().String()).
		WithExposedPort(10000).
		WithExec([]string{"azurite-blob", "--blobHost", "0.0.0.0", "--silent"}).
		AsService()

	_, err := base.
		WithServiceBinding("azurite", azurite).
		WithEnvVariable("AZURE_STORAGE_ACCOUNT", "devstoreaccount1").
		WithEnvVariable("AZURE_STORAGE_KEY", "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==").
		WithExec([]string{"go", "run", "./build/internal/cmd/azurite/...", "-url", "http://azurite:10000/devstoreaccount1", "-testdata-dir", singleRevisionTestdataDir}).
		Sync(ctx)
	if err != nil {
		return func() error { return err }
	}

	flipt = flipt.
		WithServiceBinding("azurite", azurite).
		WithEnvVariable("FLIPT_LOG_LEVEL", "DEBUG").
		WithEnvVariable("FLIPT_STORAGE_TYPE", "object").
		WithEnvVariable("FLIPT_STORAGE_OBJECT_TYPE", "azblob").
		WithEnvVariable("FLIPT_STORAGE_OBJECT_AZBLOB_ENDPOINT", "http://azurite:10000/devstoreaccount1").
		WithEnvVariable("FLIPT_STORAGE_OBJECT_AZBLOB_CONTAINER", "testdata").
		WithEnvVariable("AZURE_STORAGE_ACCOUNT", "devstoreaccount1").
		WithEnvVariable("AZURE_STORAGE_KEY", "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==").
		WithEnvVariable("UNIQUE", uuid.New().String())

	return suite(ctx, "readonly", base, flipt.WithExec(nil), conf)
}

// gcs simulates the Google Cloud Storage service
func gcs(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
	gcs := client.Container().
		From("fsouza/fake-gcs-server").
		WithEnvVariable("UNIQUE", uuid.New().String()).
		WithExposedPort(4443).
		WithExec([]string{"-scheme", "http", "-public-host", "gcs:4443"}).
		AsService()

	_, err := base.
		WithServiceBinding("gcs", gcs).
		WithEnvVariable("STORAGE_EMULATOR_HOST", "gcs:4443").
		WithExec([]string{"go", "run", "./build/internal/cmd/gcs/...", "-testdata-dir", singleRevisionTestdataDir}).
		Sync(ctx)
	if err != nil {
		return func() error { return err }
	}

	flipt = flipt.
		WithServiceBinding("gcs", gcs).
		WithEnvVariable("FLIPT_LOG_LEVEL", "DEBUG").
		WithEnvVariable("FLIPT_STORAGE_TYPE", "object").
		WithEnvVariable("FLIPT_STORAGE_OBJECT_TYPE", "googlecloud").
		WithEnvVariable("FLIPT_STORAGE_OBJECT_GOOGLECLOUD_BUCKET", "testdata").
		WithEnvVariable("STORAGE_EMULATOR_HOST", "gcs:4443").
		WithEnvVariable("UNIQUE", uuid.New().String())

	return suite(ctx, "readonly", base, flipt.WithExec(nil), conf)
}

func signJWT(key crypto.PrivateKey, claims interface{}) string {
	sig, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.SignatureAlgorithm(string(jwt.RS256)), Key: key},
		(&jose.SignerOptions{}).WithType("JWT"),
	)

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
func serveOIDC(_ context.Context, _ *dagger.Client, base, flipt *dagger.Container) (*dagger.Container, string, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, "", err
	}

	rsaSigningKey := &bytes.Buffer{}
	if err := pem.Encode(rsaSigningKey, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}); err != nil {
		return nil, "", err
	}

	// generate a SA style JWT for identifying the Flipt service
	fliptSAToken := signJWT(priv, map[string]any{
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iss": "https://discover.srv",
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
		return nil, "", err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, "", err
	}

	var caCert bytes.Buffer
	if err := pem.Encode(&caCert, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}); err != nil {
		return nil, "", err
	}

	var caPrivKeyPEM bytes.Buffer
	pem.Encode(&caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	return flipt.
			WithEnvVariable("FLIPT_LOG_LEVEL", "debug").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_KUBERNETES_DISCOVERY_URL", "https://discover.svc").
			WithServiceBinding("discover.svc", base.
				WithNewFile("/server.crt", dagger.ContainerWithNewFileOpts{
					Contents: caCert.String(),
				}).
				WithNewFile("/server.key", dagger.ContainerWithNewFileOpts{
					Contents: caPrivKeyPEM.String(),
				}).
				WithNewFile("/priv.pem", dagger.ContainerWithNewFileOpts{
					Contents: rsaSigningKey.String(),
				}).
				WithExposedPort(443).
				WithExec([]string{
					"sh",
					"-c",
					"go run ./build/internal/cmd/discover/... --private-key /priv.pem",
				}).
				AsService()).
			WithNewFile("/var/run/secrets/kubernetes.io/serviceaccount/token",
				dagger.ContainerWithNewFileOpts{Contents: fliptSAToken}).
			WithNewFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
				dagger.ContainerWithNewFileOpts{Contents: caCert.String()}),
		// generate a JWT to used to identify the workload communicating with Flipt
		// using the private key components of the key pair served by the OIDC server
		signJWT(priv, map[string]any{
			"exp": time.Now().Add(24 * time.Hour).Unix(),
			"iss": "https://discover.svc",
			"kubernetes.io": map[string]any{
				"namespace": "integration",
				"pod": map[string]any{
					"name": "integration-test-7d26f049-kdurb",
					"uid":  "bd8299f9-c50f-4b76-af33-9d8e3ef2b850",
				},
				"serviceaccount": map[string]any{
					"name": "integration-test",
					"uid":  "4f18914e-f276-44b2-aebd-27db1d8f8def",
				},
			},
		}), nil
}
