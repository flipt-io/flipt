package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"go.flipt.io/build/internal/dagger"
	"golang.org/x/sync/errgroup"
)

func Unit(ctx context.Context, client *dagger.Client, flipt *dagger.Container) (*dagger.File, error) {
	// create Redis service container
	redisSrv := client.Container().
		From("redis:alpine").
		WithExposedPort(6379)

	gitea := client.Container().
		From("gitea/gitea:1.21.1").
		WithExposedPort(3000)

	flipt = flipt.
		WithServiceBinding("gitea", gitea.AsService()).
		WithExec([]string{"go", "run", "./build/internal/cmd/gitea/...", "-gitea-url", "http://gitea:3000", "-testdata-dir", "./internal/storage/fs/git/testdata"})

	out, err := flipt.Stdout(ctx)
	if err != nil {
		return nil, err
	}

	var push map[string]string
	if err := json.Unmarshal([]byte(out), &push); err != nil {
		return nil, err
	}

	if goFlags := os.Getenv("GOFLAGS"); goFlags != "" {
		flipt = flipt.WithEnvVariable("GOFLAGS", goFlags)
	}

	flipt, err = flipt.
		WithServiceBinding("redis", redisSrv.AsService()).
		WithEnvVariable("REDIS_HOST", "redis:6379").
		WithEnvVariable("TEST_GIT_REPO_URL", "http://gitea:3000/root/features.git").
		WithEnvVariable("TEST_GIT_REPO_HEAD", push["HEAD"]).
		WithEnvVariable("TEST_GIT_REPO_TAG", push["TAG"]).
		WithExec([]string{"go", "test", "-race", "-coverprofile=coverage.txt", "-covermode=atomic", "-coverpkg=./...", "./..."}).
		Sync(ctx)
	if err != nil {
		return nil, err
	}

	// attempt to export coverage if its exists
	return flipt.File("coverage.txt"), nil
}

// IntegrationCoverage runs integration tests with coverage collection enabled
func IntegrationCoverage(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container, opts ...IntegrationOptions) (*dagger.File, error) {
	var options integrationOptions

	for _, opt := range opts {
		opt(&options)
	}

	cases, err := filterCases(options.cases...)
	if err != nil {
		return nil, err
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

	// Create a coverage volume to collect coverage data from all test containers
	coverageVolume := client.CacheVolume("integration-coverage")

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
				// Configure the Flipt container for coverage collection
				flipt = flipt.
					WithEnvVariable("CI", os.Getenv("CI")).
					WithEnvVariable("FLIPT_LOG_LEVEL", "WARN").
					WithEnvVariable("GOCOVERDIR", "/tmp/coverage").
					WithMountedCache("/tmp/coverage", coverageVolume).
					WithExposedPort(config.port)

				return fn(ctx, client, base, flipt, config)()
			}))
		}
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Extract coverage data and convert to text format
	coverageContainer := client.Container().
		From("golang:1.24-alpine3.21").
		WithMountedCache("/tmp/coverage", coverageVolume).
		WithExec([]string{"go", "tool", "covdata", "textfmt", "-i=/tmp/coverage", "-o=/tmp/coverage.out"})

	return coverageContainer.File("/tmp/coverage.out"), nil
}
