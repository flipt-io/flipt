package testing

import (
	"context"
	"os"

	"github.com/google/uuid"
	"go.flipt.io/build/internal/dagger"
)

func LoadTest(ctx context.Context, client *dagger.Client, base, flipt *dagger.Container) error {
	seed := base.File("build/testing/integration/readonly/testdata/main/default.yaml")
	importCmd := []string{"/flipt", "import", "import.yaml"}

	flipt = flipt.
		WithEnvVariable("FLIPT_DB_URL", "postgres://postgres:password@postgres:5432?sslmode=disable").
		WithEnvVariable("FLIPT_DB_MAX_OPEN_CONN", "5").
		WithEnvVariable("FLIPT_DB_MAX_IDLE_CONN", "5").
		WithEnvVariable("FLIPT_LOG_LEVEL", "warn").
		WithServiceBinding("postgres", client.Container().
			From("postgres").
			WithEnvVariable("POSTGRES_PASSWORD", "password").
			WithExposedPort(5432).
			WithExec(nil).
			AsService())

	// import some test data
	flipt, err := flipt.WithEnvVariable("UNIQUE", uuid.New().String()).
		WithFile("import.yaml", seed).
		WithExec(importCmd).
		Sync(ctx)

	if err != nil {
		return err
	}

	flipt = flipt.WithEnvVariable("UNIQUE", uuid.New().String()).
		WithExposedPort(8080)

	var authToken string
	if authEnabled := os.Getenv("FLIPT_AUTH_ENABLED"); authEnabled == "true" || authEnabled == "1" {
		authToken = uuid.New().String()
		flipt = flipt.WithEnvVariable("FLIPT_AUTHENTICATION_REQUIRED", "true").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_ENABLED", "true").
			WithEnvVariable("FLIPT_AUTHENTICATION_METHODS_TOKEN_BOOTSTRAP_TOKEN", authToken)
	}

	var cacheEnabled bool
	if cacheEnabledEnv := os.Getenv("FLIPT_CACHE_ENABLED"); cacheEnabledEnv == "true" || cacheEnabledEnv == "1" {
		cacheEnabled = true
		flipt = flipt.WithEnvVariable("FLIPT_CACHE_ENABLED", "true").
			WithEnvVariable("FLIPT_CACHE_BACKEND", "redis").
			WithEnvVariable("FLIPT_CACHE_REDIS_HOST", "redis").
			WithServiceBinding("redis", client.Container().
				From("redis:alpine").
				WithExposedPort(6379).
				WithExec(nil).
				AsService())
	}

	flipt = flipt.WithExec(nil)

	// build the loadtest binary
	loadtest := base.
		WithWorkdir("build/internal").
		WithExec([]string{"go", "build", "-o", "./out/loadtest", "./cmd/loadtest/..."}).
		File("out/loadtest")

	cmd := []string{"./loadtest", "-duration", "60s", "-rate", "300"}
	if authToken != "" {
		cmd = append(cmd, "-flipt-auth-token", authToken)
	}
	if cacheEnabled {
		cmd = append(cmd, "-flipt-cache-enabled")
	}

	// run the loadtest binary from within the pyroscope container and export the adhoc data
	// output to the host
	_, err = client.Container().
		From("pyroscope/pyroscope:latest").
		WithFile("loadtest", loadtest).
		WithServiceBinding("flipt", flipt.AsService()).
		WithExec(append([]string{"adhoc", "--log-level", "info", "--url", "flipt:8080"}, cmd...)).
		Directory("/home/pyroscope/.local/share/pyroscope").
		Export(ctx, "build/internal/out/profiles")

	return err
}
