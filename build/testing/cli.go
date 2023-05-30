package testing

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"dagger.io/dagger"
	"github.com/google/go-cmp/cmp"
)

func CLI(ctx context.Context, flipt *dagger.Container) error {
	if err := assertExec(ctx, flipt, subcommand{"foo"},
		fails,
		stderr(equals(`Error: unknown command "foo" for "flipt"
Run 'flipt --help' for usage.`))); err != nil {
		return err
	}

	expected := contains(`loading configuration	{"error": "loading configuration: open /foo/bar.yml: no such file or directory"}`)
	if err := assertExec(ctx, flipt, subcommand{"--config", "/foo/bar.yml"},
		fails,
		stdout(expected)); err != nil {
		return err
	}

	return nil
}

type subcommand []string

func (s subcommand) command() []string {
	return append([]string{"/flipt"}, s...)
}

type stringAssertion interface {
	assert(string) error
}

type assertConf struct {
	success bool
	stdout  stringAssertion
	stderr  stringAssertion
}

type assertOption func(*assertConf)

func fails(c *assertConf) { c.success = false }

func stdout(a stringAssertion) assertOption {
	return func(c *assertConf) {
		c.stdout = a
	}
}

func stderr(a stringAssertion) assertOption {
	return func(c *assertConf) {
		c.stderr = a
	}
}

type equals string

func (e equals) assert(t string) error {
	if diff := cmp.Diff(string(e), t); diff != "" {
		return fmt.Errorf("unexpected output: diff (-/+):\n%s", diff)
	}

	return nil
}

type contains string

func (c contains) assert(t string) error {
	if !strings.Contains(t, string(c)) {
		return fmt.Errorf("unexpected output: %q does not contain %q", t, c)
	}

	return nil
}

func assertExec(ctx context.Context, flipt *dagger.Container, s subcommand, opts ...assertOption) error {
	conf := assertConf{success: true}
	for _, opt := range opts {
		opt(&conf)
	}

	container, err := flipt.WithExec(s.command()).Sync(ctx)

	var (
		stdout = container.Stdout
		stderr = container.Stderr
	)

	if err != nil {
		if conf.success {
			return fmt.Errorf("unexpected error running flipt %q: %w", s, err)
		}

		var eerr *dagger.ExecError
		// get stdout and stderr from exec error instead
		if errors.As(err, &eerr) {
			stdout = func(context.Context) (string, error) {
				return eerr.Stdout, nil
			}

			stderr = func(context.Context) (string, error) {
				return eerr.Stderr, nil
			}
		}
	}

	if conf.stdout != nil {
		stdout, err := stdout(ctx)
		if err != nil {
			return err
		}

		if err := conf.stdout.assert(string(stdout)); err != nil {
			return err
		}
	}

	if conf.stderr != nil {
		stderr, err := stderr(ctx)
		if err != nil {
			return err
		}

		if err := conf.stderr.assert(string(stderr)); err != nil {
			return err
		}
	}

	return nil
}
