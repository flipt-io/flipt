package testing

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"dagger.io/dagger"
	"github.com/containerd/containerd/platforms"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

const bootstrapToken = "s3cr3t"

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
		"fs/git":        git,
		"fs/local":      local,
		"fs/s3":         s3,
		"fs/oci":        oci,
		"fs/azblob":     azblob,
		"import/export": importExport,
	}
)

type testConfig struct {
	name      string
	namespace string
	address   string
	auth      authConfig
	port      int
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

type authConfig int

const (
	noAuth authConfig = iota
	authNoNamespace
	authNamespaced
)

func (a authConfig) enabled() bool {
	return a != noAuth
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

	for _, namespace := range []string{"", "production"} {
		for protocol, port := range protocolPorts {
			for _, auth := range []authConfig{noAuth, authNoNamespace, authNamespaced} {
				config := testConfig{
					name:      fmt.Sprintf("%s namespace %s", strings.ToUpper(protocol), namespace),
					namespace: namespace,
					auth:      auth,
					address:   fmt.Sprintf("%s://flipt:%d", protocol, port),
					port:      port,
				}

				switch auth {
				case noAuth:
					config.name = fmt.Sprintf("%s without auth", config.name)
				case authNoNamespace:
					config.name = fmt.Sprintf("%s with auth no namespaced token", config.name)
				case authNamespaced:
					config.name = fmt.Sprintf("%s with auth namespaced token", config.name)
				}

				configs = append(configs, config)
			}
		}
	}

	var g errgroup.Group

	for caseName, fn := range cases {
		for _, config := range configs {
			var (
				fn     = fn
				config = config
			)

			flipt := flipt
			if config.auth.enabled() {
				flipt = flipt.
					WithEnvVariable("FLIPT_AUTHENTICATION_REQUIRED", "true").
					WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_ENABLED", "true").
					WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_BOOTSTRAP_TOKEN", bootstrapToken)
			}

			name := strings.ToLower(replacer.Replace(fmt.Sprintf("flipt-test-%s-config-%s", caseName, config.name)))
			flipt = flipt.
				WithEnvVariable("CI", os.Getenv("CI")).
				WithEnvVariable("FLIPT_LOG_LEVEL", "debug").
				WithEnvVariable("FLIPT_LOG_FILE", fmt.Sprintf("/var/opt/flipt/logs/%s.log", name)).
				WithMountedCache("/var/opt/flipt/logs", logs).
				WithExposedPort(config.port)

			g.Go(take(fn(ctx, client, base.Pipeline(name), flipt, config)))
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
				WithExec([]string{"start-single-node", "--insecure"}).
				AsService()),
			conf,
		)

	}
}

func cache(ctx context.Context, _ *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
	flipt = flipt.
		WithEnvVariable("FLIPT_LOG_LEVEL", "DEBUG").
		WithEnvVariable("FLIPT_CACHE_ENABLED", "true")

	return suite(ctx, "api", base, flipt.WithExec(nil), conf)
}

const (
	testdataDir     = "build/testing/integration/readonly/testdata"
	testdataPathFmt = testdataDir + "/%s.yaml"
)

func local(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
	flipt = flipt.
		WithDirectory("/tmp/testdata", base.Directory(testdataDir)).
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

	_, err := base.
		WithServiceBinding("gitea", gitea.AsService()).
		WithExec([]string{"go", "run", "./build/internal/cmd/gitea/...", "-gitea-url", "http://gitea:3000", "-testdata-dir", testdataDir}).
		Sync(ctx)
	if err != nil {
		return func() error { return err }
	}

	flipt = flipt.
		WithServiceBinding("gitea", gitea.AsService()).
		WithEnvVariable("FLIPT_LOG_LEVEL", "DEBUG").
		WithEnvVariable("FLIPT_STORAGE_TYPE", "git").
		WithEnvVariable("FLIPT_STORAGE_GIT_REPOSITORY", "http://gitea:3000/root/features.git").
		WithEnvVariable("UNIQUE", uuid.New().String())

	return suite(ctx, "readonly", base, flipt.WithExec(nil), conf)
}

func s3(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
	minio := client.Container().
		From("quay.io/minio/minio:latest").
		WithExposedPort(9009).
		WithEnvVariable("MINIO_ROOT_USER", "user").
		WithEnvVariable("MINIO_ROOT_PASSWORD", "password").
		WithEnvVariable("MINIO_BROWSER", "off").
		WithExec([]string{"server", "/data", "--address", ":9009", "--quiet"})

	_, err := base.
		WithServiceBinding("minio", minio.AsService()).
		WithEnvVariable("AWS_ACCESS_KEY_ID", "user").
		WithEnvVariable("AWS_SECRET_ACCESS_KEY", "password").
		WithExec([]string{"go", "run", "./build/internal/cmd/minio/...", "-minio-url", "http://minio:9009", "-testdata-dir", testdataDir}).
		Sync(ctx)
	if err != nil {
		return func() error { return err }
	}

	flipt = flipt.
		WithServiceBinding("minio", minio.AsService()).
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
		WithDirectory("/tmp/testdata", base.Directory(testdataDir)).
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

func importExport(ctx context.Context, _ *dagger.Client, base, flipt *dagger.Container, conf testConfig) func() error {
	return func() error {
		// import testdata before running readonly suite
		flags := []string{"--address", conf.address}
		if conf.auth.enabled() {
			flags = append(flags, "--token", bootstrapToken)
		}

		ns := "default"
		if conf.namespace != "" {
			ns = conf.namespace
		}

		seed := base.File(fmt.Sprintf(testdataPathFmt, ns))

		var (
			// create unique instance for test case
			fliptToTest = flipt.
					WithEnvVariable("UNIQUE", uuid.New().String()).
					WithExec(nil)

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

		// run readonly suite against imported Flipt instance
		if err := suite(ctx, "readonly", base, fliptToTest, conf)(); err != nil {
			return err
		}

		expected, err := seed.Contents(ctx)
		if err != nil {
			return err
		}

		namespace := conf.namespace
		if namespace == "" {
			namespace = "default"
			// replace namespace in expected yaml
			expected = strings.ReplaceAll(expected, "version: \"1.2\"\n", fmt.Sprintf("version: \"1.2\"\nnamespace: %s\n", namespace))
		}

		if namespace != "default" {
			flags = append(flags, "--namespaces", conf.namespace)
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

		return nil
	}
}

func suite(ctx context.Context, dir string, base, flipt *dagger.Container, conf testConfig) func() error {
	return func() (err error) {
		flags := []string{"--flipt-addr", conf.address}
		if conf.namespace != "" {
			flags = append(flags, "--flipt-namespace", conf.namespace)
		}

		if conf.auth.enabled() {
			flags = append(flags, "--flipt-token", bootstrapToken)
			if conf.auth == authNamespaced {
				flags = append(flags, "--flipt-create-namespaced-token")
			}
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
		WithExposedPort(10000).
		WithExec([]string{"azurite-blob", "--blobHost", "0.0.0.0", "--silent"}).
		AsService()

	_, err := base.
		WithServiceBinding("azurite", azurite).
		WithEnvVariable("AZURE_STORAGE_ACCOUNT", "devstoreaccount1").
		WithEnvVariable("AZURE_STORAGE_KEY", "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==").
		WithExec([]string{"go", "run", "./build/internal/cmd/azurite/...", "-url", "http://azurite:10000/devstoreaccount1", "-testdata-dir", testdataDir}).
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
