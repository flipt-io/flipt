package testing

import (
	"context"

	"dagger.io/dagger"
)

func Config(ctx context.Context, client *dagger.Client, container *dagger.Container) error {
	_, err := client.Container().
		Pipeline("validate advanced.yml with CUE").
		From("golang:1.20-alpine3.16").
		WithExec([]string{"go", "install", "cuelang.org/go/cmd/cue@v0.5.0"}).
		WithFile("/flipt.schema.cue", container.File("config/flipt.schema.cue")).
		WithFile("/config.yml", container.File("internal/config/testdata/advanced.yml")).
		WithExec([]string{"cue", "eval", "-d", "#FliptSpec", "-c", "--strict", "-E", "/flipt.schema.cue", "/config.yml"}).
		Sync(ctx)

	return err
}
