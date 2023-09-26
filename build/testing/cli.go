package testing

import (
	"context"
	"fmt"
	"os"
	"strings"

	"dagger.io/dagger"
)

func CLI(ctx context.Context, client *dagger.Client, container *dagger.Container) error {
	{
		container := container.Pipeline("flipt --help")
		expected, err := os.ReadFile("build/testing/testdata/cli.txt")
		if err != nil {
			return err
		}

		if _, err := assertExec(ctx, container, flipt("--help"),
			fails,
			stdout(equals(expected))); err != nil {
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
			stderr(contains(`unknown command "foo" for "flipt"
Run 'flipt --help' for usage.`))); err != nil {
			return err
		}

		if _, err := assertExec(ctx, container, flipt("--config", "/foo/bar.yml"),
			fails,
			stderr(contains("loading configuration: open /foo/bar.yml: no such file or directory")),
		); err != nil {
			return err
		}

		if _, err := assertExec(ctx, container, flipt("--config", "/tmp"),
			fails,
			stderr(contains(`loading configuration: Unsupported Config Type`)),
		); err != nil {
			return err
		}
	}

	{
		container := container.Pipeline("flipt (no config)")
		if _, err := assertExec(ctx, container.WithExec([]string{"rm", "/etc/flipt/config/default.yml"}), flipt(),
			fails,
			stderr(contains(`loading configuration: open /etc/flipt/config/default.yml: no such file or directory`)),
		); err != nil {
			return err
		}
	}

	{
		container := container.Pipeline("flipt (user config directory)")
		container = container.
			WithExec([]string{"mkdir", "-p", "/home/flipt/.config/flipt"}).
			WithFile("/home/flipt/.config/flipt/config.yml", client.Host().Directory("build/testing/testdata").File("default.yml")).
			// in order to stop a blocking process via SIGTERM and capture a successful exit code
			// we use a shell script to start flipt in the background, sleep for two seconds,
			// send the SIGTERM signal, wait for process to exit and then propagate Flipts exit code
			WithNewFile("/test.sh", dagger.ContainerWithNewFileOpts{
				Contents: `#!/bin/sh

/flipt &

sleep 2

kill -s TERM $!

wait $!

exit $?`,
				Owner:       "flipt",
				Permissions: 0777,
			}).
			WithEnvVariable("FLIPT_LOG_LEVEL", "debug")

		if _, err := assertExec(ctx, container, []string{"/test.sh"},
			stdout(contains("configuration source\t{\"path\": \"/home/flipt/.config/flipt/config.yml\"}")),
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
			stderr(contains("opening import file: open foo: no such file or directory")),
		); err != nil {
			return err
		}
		opts := dagger.ContainerWithFileOpts{
			Owner: "flipt",
		}

		container = container.WithFile("/tmp/flipt.yml",
			client.Host().Directory("build/testing/testdata").File("flipt.yml"), opts)

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
		container := container.Pipeline("flipt import create namespace")

		opts := dagger.ContainerWithFileOpts{
			Owner: "flipt",
		}

		container = container.WithFile("/tmp/flipt.yml",
			client.Host().Directory("build/testing/testdata").File("flipt-namespace-foo.yml"), opts)

		// import via STDIN succeeds
		_, err := assertExec(ctx, container, sh("cat /tmp/flipt.yml | /flipt import --create-namespace --stdin"))
		if err != nil {
			return err
		}

		// import valid yaml path and retrieve resulting container
		container, err = assertExec(ctx, container, flipt("import", "--create-namespace", "/tmp/flipt.yml"))
		if err != nil {
			return err
		}

		if _, err = assertExec(ctx, container,
			flipt("export", "--namespace", "foo"),
			stdout(contains(expectedFliptYAML)),
		); err != nil {
			return err
		}

		container, err = assertExec(ctx, container,
			flipt("export", "--namespace", "foo", "-o", "/tmp/export.yml"),
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

	{
		container := container.Pipeline("flipt config init")

		container, err := assertExec(ctx, container, flipt("config", "init", "-y"))
		if err != nil {
			return err
		}

		contents, err := container.File("/home/flipt/.config/flipt/config.yml").Contents(ctx)
		if err != nil {
			return err
		}

		expected := `# yaml-language-server: $schema=https://raw.githubusercontent.com/flipt-io/flipt/main/config/flipt.schema.json`

		if !strings.Contains(contents, expected) {
			return fmt.Errorf("unexpected output: %q does not contain %q", contents, expected)
		}
	}

	return nil
}

const (
	expectedFliptYAML = `flags:
- key: zUFtS7D0UyMeueYu
  name: UAoZRksg94r1iipa
  type: VARIANT_FLAG_TYPE
  description: description
  enabled: true
  variants:
  - key: NGxfcVffpMhBz9n8
    name: fhDHQ7rcxvoaWbHw
  - key: sDGD6NvfCRyaQUn3
  rules:
  - segment: 08UoVJ96LhZblPEx
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
