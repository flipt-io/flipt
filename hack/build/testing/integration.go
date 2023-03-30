package testing

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"dagger.io/dagger"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

var protocolPorts = map[string]string{"http": "8080", "grpc": "9000"}

type testCase struct {
	name       string
	namespace  string
	address    string
	extraEnvs  [][2]string
	extraFlags []string
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

	var cases []testCase

	for _, namespace := range []string{
		"",
		fmt.Sprintf("%x", rand.Int()),
	} {
		for protocol, port := range protocolPorts {
			address := fmt.Sprintf("%s://flipt:%s", protocol, port)
			cases = append(cases,
				testCase{
					name:      fmt.Sprintf("%s no authentication", strings.ToUpper(protocol)),
					namespace: namespace,
					address:   address,
				},
				testCase{
					name:      fmt.Sprintf("%s with authentication", strings.ToUpper(protocol)),
					namespace: namespace,
					address:   address,
					extraEnvs: [][2]string{
						{"FLIPT_AUTHENTICATION_REQUIRED", "true"},
						{"FLIPT_AUTHENTICATION_METHODS_TOKEN_ENABLED", "true"},
						{"FLIPT_AUTHENTICATION_METHODS_TOKEN_BOOTSTRAP_TOKEN", "some-token"},
					},
					extraFlags: []string{"-flipt-token", "some-token"},
				},
			)
		}
	}

	rand.Seed(time.Now().Unix())

	var g errgroup.Group
	for _, test := range cases {
		test := test

		flipt := flipt
		for _, env := range test.extraEnvs {
			flipt = flipt.WithEnvVariable(env[0], env[1])
		}

		g.Go(integration(
			ctx,
			test.name,
			base,
			flipt,
			logs,
			append(test.extraFlags,
				"-flipt-addr", test.address,
				"-flipt-namespace", test.namespace,
			)...,
		))
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

func integration(ctx context.Context, name string, base, flipt *dagger.Container, logs *dagger.CacheVolume, flags ...string) func() error {
	name = fmt.Sprintf("flipt-%s", strings.ToLower(strings.ReplaceAll(name, " ", "-")))
	logFile := fmt.Sprintf("/var/opt/flipt/logs/%s.log", name)

	flipt = flipt.
		// this ensures a unique instance of flipt is created per
		// integration test
		WithEnvVariable("UNIQUE", uuid.New().String()).
		WithEnvVariable("FLIPT_LOG_LEVEL", "debug").
		WithMountedCache("/var/opt/flipt/logs", logs).
		WithEnvVariable("FLIPT_LOG_FILE", logFile).
		WithExec(nil)

	return func() error {
		flags := append([]string{"-v", "-race"}, flags...)
		_, err := base.
			WithServiceBinding("flipt", flipt).
			WithWorkdir("hack/build/integration").
			WithExec(append([]string{"go", "test"}, append(flags, ".")...)).
			ExitCode(ctx)

		return err
	}
}
