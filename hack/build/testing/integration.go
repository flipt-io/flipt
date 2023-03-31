package testing

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"path"
	"strings"
	"time"

	"dagger.io/dagger"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

var (
	protocolPorts = map[string]string{"http": "8080", "grpc": "9000"}
	replacer      = strings.NewReplacer(" ", "-", "/", "-")
	sema          = make(chan struct{}, 10)
)

type testConfig struct {
	name      string
	namespace string
	address   string
	token     string
}

func Integration(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container) error {
	logs := client.CacheVolume(fmt.Sprintf("logs-%s", uuid.New()))
	_, err := flipt.WithUser("root").
		WithMountedCache("/logs", logs).
		WithExec([]string{"chown", "flipt:flipt", "/logs"}).
		ExitCode(ctx)
	if err != nil {
		return err
	}

	var cases []testConfig

	for _, namespace := range []string{
		"",
		fmt.Sprintf("%x", rand.Int()),
	} {
		for protocol, port := range protocolPorts {
			address := fmt.Sprintf("%s://flipt:%s", protocol, port)
			cases = append(cases,
				testConfig{
					name:      fmt.Sprintf("%s namespace %s no authentication", strings.ToUpper(protocol), namespace),
					namespace: namespace,
					address:   address,
				},
				testConfig{
					name:      fmt.Sprintf("%s namespace %s with authentication", strings.ToUpper(protocol), namespace),
					namespace: namespace,
					address:   address,
					token:     "some-token",
				},
			)
		}
	}

	rand.Seed(time.Now().Unix())

	var g errgroup.Group

	for _, test := range []struct {
		name string
		fn   func(_ context.Context, base, flipt *dagger.Container, conf testConfig) func() error
	}{
		{
			name: "api",
			fn:   api,
		},
		{
			name: "import/export",
			fn:   importExport,
		},
	} {
		for _, config := range cases {
			config := config

			flipt := flipt
			if config.token != "" {
				flipt = flipt.
					WithEnvVariable("FLIPT_AUTHENTICATION_REQUIRED", "true").
					WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_ENABLED", "true").
					WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_BOOTSTRAP_TOKEN", config.token)
			}

			name := strings.ToLower(replacer.Replace(fmt.Sprintf("flipt-test-%s-config-%s", test.name, config.name)))
			flipt = flipt.
				WithEnvVariable("FLIPT_LOG_LEVEL", "debug").
				WithEnvVariable("FLIPT_LOG_FILE", fmt.Sprintf("/var/opt/flipt/logs/%s.log", name)).
				WithMountedCache("/var/opt/flipt/logs", logs)

			g.Go(take(test.fn(ctx, base, flipt, config)))
		}
	}

	err = g.Wait()

	if _, lerr := client.Container().From("alpine:3.16").
		WithMountedCache("/logs", logs).
		WithExec([]string{"cp", "-r", "/logs", "/out"}).
		Directory("/out").Export(ctx, "hack/build/logs"); lerr != nil {
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

func importExport(ctx context.Context, base, flipt *dagger.Container, conf testConfig) func() error {
	return func() error {
		// import testdata before running readonly suite
		flags := []string{"--address", conf.address}
		if conf.token != "" {
			flags = append(flags, "--token", conf.token)
		}

		if conf.namespace != "" {
			flags = append(flags, "--namespace", conf.namespace)
		}

		var (
			// create unique instance for test case
			fliptToTest = flipt.
					WithEnvVariable("UNIQUE", uuid.New().String()).
					WithExec(nil)

			importCmd = append([]string{"/bin/flipt", "import"}, append(flags, "--create-namespace", "import.yaml")...)
			seed      = base.File("hack/build/testing/integration/readonly/testdata/seed.yaml")
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
			WithExec([]string{"sh", "-c", fmt.Sprintf("sleep 5 && %s", strings.Join(importCmd, " "))}).
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

		// use target flipt binary to invoke import
		generated, err := flipt.
			WithEnvVariable("UNIQUE", uuid.New().String()).
			WithServiceBinding("flipt", fliptToTest).
			WithExec(append([]string{"/bin/flipt", "export"}, flags...)).
			Stdout(ctx)
		if err != nil {
			return err
		}

		if expected != generated {
			fmt.Println("Unexpected difference in exported output:")
			fmt.Println("Expected:")
			fmt.Println(expected + "\n")
			fmt.Println("Found:")
			fmt.Println(generated)

			return errors.New("Exported yaml did not match.")
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
			WithWorkdir(path.Join("hack/build/testing/integration", dir)).
			WithExec(append([]string{"go", "test", "-v", "-race"}, append(flags, ".")...)).
			ExitCode(ctx)

		return err
	}
}
