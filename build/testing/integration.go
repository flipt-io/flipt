package testing

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"

	"dagger.io/dagger"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

var (
	protocolPorts = map[string]string{"http": "8080", "grpc": "9000"}
	replacer      = strings.NewReplacer(" ", "-", "/", "-")
	sema          = make(chan struct{}, 6)

	// AllCases are the top-level filterable integration test cases.
	AllCases = map[string]testCaseFn{
		"api":           api,
		"import/export": importExport,
	}
)

type testConfig struct {
	name      string
	namespace string
	address   string
	token     string
}

type testCaseFn func(_ context.Context, base, flipt *dagger.Container, conf testConfig) func() error

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
		ExitCode(ctx)
	if err != nil {
		return err
	}

	var configs []testConfig

	for _, namespace := range []string{"", "production"} {
		for protocol, port := range protocolPorts {
			for _, token := range []string{"", "some-token"} {
				name := fmt.Sprintf("%s namespace %s", strings.ToUpper(protocol), namespace)
				if token != "" {
					name = fmt.Sprintf("%s with token %s", name, token)
				}

				configs = append(configs,
					testConfig{
						name:      name,
						namespace: namespace,
						address:   fmt.Sprintf("%s://flipt:%s", protocol, port),
						token:     token,
					},
				)
			}
		}
	}

	rand.Seed(time.Now().Unix())

	var g errgroup.Group

	for caseName, fn := range cases {
		for _, config := range configs {
			var (
				fn     = fn
				config = config
			)

			flipt := flipt
			if config.token != "" {
				flipt = flipt.
					WithEnvVariable("FLIPT_AUTHENTICATION_REQUIRED", "true").
					WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_ENABLED", "true").
					WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_BOOTSTRAP_TOKEN", config.token)
			}

			name := strings.ToLower(replacer.Replace(fmt.Sprintf("flipt-test-%s-config-%s", caseName, config.name)))
			flipt = flipt.
				WithEnvVariable("CI", os.Getenv("CI")).
				WithEnvVariable("FLIPT_LOG_LEVEL", "debug").
				WithEnvVariable("FLIPT_LOG_FILE", fmt.Sprintf("/var/opt/flipt/logs/%s.log", name)).
				WithMountedCache("/var/opt/flipt/logs", logs)

			g.Go(take(fn(ctx, base, flipt, config)))
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

func api(ctx context.Context, base, flipt *dagger.Container, conf testConfig) func() error {
	return suite(ctx, "api", base,
		// create unique instance for test case
		flipt.
			WithEnvVariable("UNIQUE", uuid.New().String()).
			WithExec(nil), conf)
}

const (
	testdataDir     = "build/testing/integration/readonly/testdata"
	testdataPathFmt = testdataDir + "/%s.yaml"
)

func importExport(ctx context.Context, base, flipt *dagger.Container, conf testConfig) func() error {
	return func() error {
		// import testdata before running readonly suite
		flags := []string{"--address", conf.address}
		if conf.token != "" {
			flags = append(flags, "--token", conf.token)
		}

		ns := "default"
		if conf.namespace != "" {
			ns = conf.namespace
			flags = append(flags, "--namespace", conf.namespace)
		}

		seed := base.File(fmt.Sprintf(testdataPathFmt, ns))

		var (
			// create unique instance for test case
			fliptToTest = flipt.
					WithEnvVariable("UNIQUE", uuid.New().String()).
					WithExec(nil)

			importCmd = append([]string{"/flipt", "import"}, append(flags, "--create-namespace", "import.yaml")...)
		)
		// use target flipt binary to invoke import
		_, err := flipt.
			WithEnvVariable("UNIQUE", uuid.New().String()).
			// copy testdata import yaml from base
			WithFile("import.yaml", seed).
			WithServiceBinding("flipt", fliptToTest).
			// it appears it takes a little while for Flipt to come online
			// For the go tests they have to compile and that seems to be enough
			// time for the target Flipt to come up.
			// However, in this case the flipt binary is prebuilt and needs a little sleep.
			WithExec([]string{"sh", "-c", fmt.Sprintf("sleep 2 && %s", strings.Join(importCmd, " "))}).
			ExitCode(ctx)
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
			expected = strings.ReplaceAll(expected, "version: \"1.0\"\n", fmt.Sprintf("version: \"1.0\"\nnamespace: %s\n", namespace))
		}

		// use target flipt binary to invoke import
		generated, err := flipt.
			WithEnvVariable("UNIQUE", uuid.New().String()).
			WithServiceBinding("flipt", fliptToTest).
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
			fmt.Println("Unexpected difference in exported output:")
			fmt.Println(diff)
			return errors.New("exported yaml did not match")
		}

		return nil
	}
}

func suite(ctx context.Context, dir string, base, flipt *dagger.Container, conf testConfig) func() error {
	return func() error {
		flags := []string{"--flipt-addr", conf.address}
		if conf.namespace != "" {
			flags = append(flags, "--flipt-namespace", conf.namespace)
		}

		if conf.token != "" {
			flags = append(flags, "--flipt-token", conf.token)
		}

		_, err := base.
			WithServiceBinding("flipt", flipt).
			WithWorkdir(path.Join("build/testing/integration", dir)).
			WithEnvVariable("UNIQUE", uuid.New().String()).
			WithExec(append([]string{"go", "test", "-v", "-race"}, append(flags, ".")...)).
			ExitCode(ctx)

		return err
	}
}
