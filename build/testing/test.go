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
	var services []*dagger.Service // Keep track of all started services to stop them later

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
				coverageFliptContainer := flipt.
					WithEnvVariable("CI", os.Getenv("CI")).
					WithEnvVariable("FLIPT_LOG_LEVEL", "DEBUG").  // Increase log level to see coverage activity
					WithEnvVariable("GOCOVERDIR", "/tmp/coverage").
					WithMountedCache("/tmp/coverage", coverageVolume).
					WithExposedPort(config.port)

				// Create the service and start it explicitly to ensure proper lifecycle management
				fliptService := coverageFliptContainer.AsService()
				startedService, err := fliptService.Start(ctx)
				if err != nil {
					return fmt.Errorf("failed to start Flipt service: %w", err)
				}

				// Add to services list for cleanup (Note: this has race condition, but for debugging purposes)
				services = append(services, startedService)

				// Run the original test function, but pass a container that references the started service
				// The test function will call its internal suite() which will bind to the service
				testFunc := fn(ctx, client, base, coverageFliptContainer, config)
				
				// Execute the test
				err = testFunc()
				
				// Stop the service after test completes to flush coverage data
				fmt.Printf("Stopping Flipt service to flush coverage data...\n")
				if _, stopErr := startedService.Stop(ctx); stopErr != nil {
					fmt.Printf("Warning: failed to stop Flipt service: %v\n", stopErr)
				} else {
					fmt.Printf("Successfully stopped Flipt service\n")
				}
				
				return err
			}))
		}
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Try a simple test: run the coverage-enabled binary directly for a few seconds to see if it writes anything
	testContainer := client.Container().
		From("golang:1.24-alpine3.21").
		WithMountedCache("/tmp/coverage", coverageVolume).
		WithEnvVariable("GOCOVERDIR", "/tmp/coverage").
		WithFile("/flipt", flipt.File("/flipt")).
		WithExec([]string{"sh", "-c", "echo 'Testing coverage binary directly...' && timeout 5 /flipt server --config /dev/null || echo 'Server stopped' && echo 'Coverage directory after direct run:' && ls -la /tmp/coverage && find /tmp/coverage -type f"}).
		WithExec([]string{"sh", "-c", "if [ -n \"$(find /tmp/coverage -type f 2>/dev/null)\" ]; then echo 'Found files from direct test, converting...' && go tool covdata textfmt -i=/tmp/coverage -o=/tmp/coverage.out 2>&1; else echo 'Debug: Direct binary test found no coverage files' > /tmp/coverage.out; fi"})

	return testContainer.File("/tmp/coverage.out"), nil
}
