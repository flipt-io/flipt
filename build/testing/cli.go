package testing

import (
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger"
)

func CLI(ctx context.Context, client *dagger.Client, container *dagger.Container) error {
	{
		container := container.Pipeline("flipt --help")
		if _, err := assertExec(ctx, container, flipt("--help"),
			fails,
			stdout(equals(expectedFliptHelp))); err != nil {
			return err
		}
	}

	{
		container := container.Pipeline("flipt --version")
		if _, err := assertExec(ctx, container, flipt("--version"),
			fails,
			stdout(contains("Commit:")),
			stdout(matches(`Build Date: [0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z`)),
			stdout(matches(`Go Version: go[0-9]+\.[0-9]+\.[0-9]`)),
		); err != nil {
			return err
		}
	}

	{
		container := container.Pipeline("flipt --config")
		if _, err := assertExec(ctx, container, flipt("foo"),
			fails,
			stderr(equals(`Error: unknown command "foo" for "flipt"
Run 'flipt --help' for usage.`))); err != nil {
			return err
		}

		if _, err := assertExec(ctx, container, flipt("--config", "/foo/bar.yml"),
			fails,
			stdout(contains(`loading configuration	{"error": "loading configuration: open /foo/bar.yml: no such file or directory", "config_path": "/foo/bar.yml"}`)),
		); err != nil {
			return err
		}

		if _, err := assertExec(ctx, container, flipt("--config", "/tmp"),
			fails,
			stdout(contains(`loading configuration: Unsupported Config Type`)),
		); err != nil {
			return err
		}
	}

	{
		container := container.Pipeline("flipt import/export")

		var err error
		if _, err = assertExec(ctx, container,
			sh("echo FOOBAR | /flipt import --stdin"),
			fails,
		); err != nil {
			return err
		}

		if _, err = assertExec(ctx, container,
			flipt("import", "foo"),
			fails,
			stdout(contains("opening import file: open foo: no such file or directory")),
		); err != nil {
			return err
		}

		container = container.WithFile("/tmp/flipt.yml",
			client.Host().Directory("build/testing/testdata").File("flipt.yml"))

		// import via STDIN succeeds
		if _, err = assertExec(ctx, container, sh("cat /tmp/flipt.yml | /flipt import --stdin")); err != nil {
			return err
		}

		// import valid yaml path and retrieve resulting container
		container, err = assertExec(ctx, container, flipt("import", "/tmp/flipt.yml"))
		if err != nil {
			return err
		}

		if _, err = assertExec(ctx, container,
			flipt("export"),
			stdout(contains(expectedFliptYAML)),
		); err != nil {
			return err
		}

		container, err = assertExec(ctx, container,
			flipt("export", "-o", "/tmp/export.yml"),
		)
		if err != nil {
			return err
		}

		contents, err := container.File("/tmp/export.yml").Contents(ctx)
		if err != nil {
			return err
		}

		if !strings.Contains(contents, expectedFliptYAML) {
			return fmt.Errorf("unexpected output: %q does not contain %q", contents, expectedFliptYAML)
		}
	}

	{
		container := container.Pipeline("flipt migrate")
		if _, err := assertExec(ctx, container, flipt("migrate")); err != nil {
			return err
		}
	}

	return nil
}

const (
	expectedFliptHelp = `Flipt is a modern feature flag solution

Usage:
  flipt [flags]
  flipt [command]

Available Commands:
  export      Export flags/segments/rules to file/stdout
  help        Help about any command
  import      Import flags/segments/rules from file
  migrate     Run pending database migrations

Flags:
      --config string   path to config file
  -h, --help            help for flipt
  -v, --version         version for flipt

Use "flipt [command] --help" for more information about a command.
`
	expectedFliptYAML = `flags:
- key: zUFtS7D0UyMeueYu
  name: UAoZRksg94r1iipa
  description: description
  enabled: true
  variants:
  - key: NGxfcVffpMhBz9n8
    name: fhDHQ7rcxvoaWbHw
  - key: sDGD6NvfCRyaQUn3
  rules:
  - segment: 08UoVJ96LhZblPEx
    rank: 1
    distributions:
    - variant: NGxfcVffpMhBz9n8
      rollout: 100
segments:
- key: 08UoVJ96LhZblPEx
  name: 2oS8SHbrxyFkRg1a
  description: description
  constraints:
  - type: STRING_COMPARISON_TYPE
    property: foo
    operator: eq
    value: baz
  - type: STRING_COMPARISON_TYPE
    property: fizz
    operator: neq
    value: buzz`
)
