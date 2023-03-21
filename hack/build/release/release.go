package release

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"dagger.io/dagger"
	"golang.org/x/mod/semver"
)

func Submodules(ctx context.Context, client *dagger.Client, tag string) error {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return errors.New("GITHUB_TOKEN is required")
	}

	base := client.Container().From("golang:1.20-alpine3.17").
		WithExec([]string{"apk", "add", "-U", "--no-cache", "git"}).
		WithExec([]string{"git", "clone", "https://github.com/flipt-io/flipt.git", "/src/flipt"}).
		WithWorkdir("/src/flipt")

	authenticated := base.WithExec([]string{"git", "config", "--global", "user.name", "flipt-automation[bot]"}).
		WithExec([]string{"git", "config", "--global", "user.email", "dev@flipt.io"}).
		WithExec([]string{"git", "config", "--global",
			"http.https://github.com/.extraheader",
			fmt.Sprintf("AUTHORIZATION: Bearer %s", token),
		})

	tagList, err := authenticated.WithExec([]string{"git", "tag", "--list", "v*"}).Stdout(ctx)
	if err != nil {
		return err
	}

	tags := strings.Split(tagList, "\n")
	semver.Sort(tags)

	lastTwo := tags[len(tags)-2:]

	// ensure we only attempt to publish updates when requested
	// tag matches the latest semver in default branch.
	if lastTwo[1] != tag {
		return fmt.Errorf("Publish tag %q does not match latest semver tag %q. Aborting.", lastTwo[1], tag)
	}

	for _, submodule := range []string{
		"errors",
		"rpc/flipt",
	} {
		if err := tagSubmodule(ctx, authenticated, submodule, lastTwo[0], lastTwo[1]); err != nil {
			return err
		}
	}

	return nil
}

func tagSubmodule(ctx context.Context, container *dagger.Container, submodule, fromTag, toTag string) error {
	diff, err := container.WithExec([]string{"git", "diff", "--shortstat", fromTag, toTag, "--", submodule}).Stdout(ctx)
	if err != nil {
		return err
	}

	if strings.TrimSpace(diff) == "" {
		fmt.Printf("Nothing changed between %q and %q. Skipping sub-module release.", fromTag, toTag)
		return nil
	}

	target := path.Join(submodule, toTag)
	container = container.
		// checkout destination version tag
		WithExec([]string{"git", "checkout", toTag}).
		// tag with target submodule tag
		WithExec([]string{
			"git", "tag", "-am",
			fmt.Sprintf("Releasing go.flipt.io/flipt/%s version %s", submodule, toTag),
			target,
		})

	_, err = container.WithExec([]string{"git", "tag", "-n", target}).ExitCode(ctx)
	if err != nil {
		return err
	}

	_, err = container.WithExec([]string{"git", "push", "origin", target}).ExitCode(ctx)

	return err
}
