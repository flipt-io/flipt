package testing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/containerd/containerd/platforms"
	"go.flipt.io/build/internal/dagger"
)

func CLI(ctx context.Context, client *dagger.Client, source *dagger.Directory, container *dagger.Container) error {
	{
		container := container.Pipeline("flipt --help")
		expected, err := source.File("build/testing/testdata/cli.txt").Contents(ctx)
		if err != nil {
			return err
		}

		if _, err := assertExec(ctx, container, flipt("--help"),
			stdout(equals(expected))); err != nil {
			return err
		}
	}

	{
		container := container.Pipeline("flipt --version")
		if _, err := assertExec(ctx, container, flipt("--version"),
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
		container := container.Pipeline("flipt (no config present) defaults")
		container = container.
			// in order to stop a blocking process via SIGTERM and capture a successful exit code
			// we use a shell script to start flipt in the background, sleep for two seconds,
			// send the SIGTERM signal, wait for process to exit and then propagate Flipts exit code
			WithNewFile("/test.sh", `#!/bin/sh
rm -rf /etc/flipt/config/default.yml
/flipt &

sleep 2

kill -s TERM $!

wait $!

exit $?`,
				dagger.ContainerWithNewFileOpts{
					Owner:       "flipt",
					Permissions: 0777,
				})

		if _, err := assertExec(ctx, container, []string{"/test.sh"},
			stdout(contains("no configuration file found, using defaults")),
		); err != nil {
			return err
		}
	}

	{
		container := container.Pipeline("flipt (user config directory)")
		container = container.
			WithExec([]string{"mkdir", "-p", "/home/flipt/.config/flipt"}).
			WithFile("/home/flipt/.config/flipt/config.yml", source.Directory("build/testing/testdata").File("default.yml")).
			// in order to stop a blocking process via SIGTERM and capture a successful exit code
			// we use a shell script to start flipt in the background, sleep for two seconds,
			// send the SIGTERM signal, wait for process to exit and then propagate Flipts exit code
			WithNewFile("/test.sh", `#!/bin/sh

/flipt &

sleep 2

kill -s TERM $!

wait $!

exit $?`,
				dagger.ContainerWithNewFileOpts{
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
		container := container.Pipeline("flipt (remote config)")
		minio := container.
			From("quay.io/minio/minio:latest").
			WithExposedPort(9009).
			WithEnvVariable("MINIO_ROOT_USER", "user").
			WithEnvVariable("MINIO_ROOT_PASSWORD", "password").
			WithEnvVariable("MINIO_BROWSER", "off").
			WithExec([]string{"server", "/data", "--address", ":9009", "--quiet"}, dagger.ContainerWithExecOpts{UseEntrypoint: true}).
			AsService()

		if _, err := assertExec(ctx,
			container.WithServiceBinding("minio", minio).
				WithEnvVariable("AWS_ACCESS_KEY_ID", "user").
				WithEnvVariable("AWS_SECRET_ACCESS_KEY", "password"),
			flipt("--config", "s3://mybucket/local.yml?region=minio&endpoint=http://minio:9009"),
			fails,
			stderr(contains(`NoSuchBucket`)),
			stderr(contains(`loading configuration: open local.yml`)),
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
			source.Directory("build/testing/testdata").File("flipt.yml"), opts)

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
		container := container.Pipeline("flipt export with mutually exclusive flags")

		_, err := assertExec(ctx, container,
			flipt("export", "--all-namespaces", "--namespaces", "foo,bar"),
			fails,
			stderr(contains("if any flags in the group [all-namespaces namespaces namespace]")),
		)
		if err != nil {
			return err
		}
	}

	{
		container := container.Pipeline("flipt import")

		opts := dagger.ContainerWithFileOpts{
			Owner: "flipt",
		}

		container = container.WithFile("/tmp/flipt.yml",
			source.Directory("build/testing/testdata").File("flipt-namespace-foo.yml"), opts)

		// import via STDIN succeeds
		_, err := assertExec(ctx, container, sh("cat /tmp/flipt.yml | /flipt import --stdin"))
		if err != nil {
			return err
		}

		// import valid yaml path and retrieve resulting container
		container, err = assertExec(ctx, container, flipt("import", "/tmp/flipt.yml"))
		if err != nil {
			return err
		}

		if _, err = assertExec(ctx, container,
			flipt("export", "--namespaces", "foo"),
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
		container := container.Pipeline("flipt import YAML stream")

		opts := dagger.ContainerWithFileOpts{
			Owner: "flipt",
		}

		container = container.WithFile("/tmp/flipt.yml",
			source.Directory("build/testing/testdata").File("flipt-yaml-stream.yml"),
			opts,
		)

		container, err := assertExec(ctx, container, sh("cat /tmp/flipt.yml | /flipt import --stdin"))
		if err != nil {
			return err
		}

		if _, err := assertExec(ctx, container,
			flipt("export", "--all-namespaces"),
			stdout(contains(expectedYAMLStreamOutput)),
		); err != nil {
			return err
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

	{
		container = container.Pipeline("flipt bundle").
			WithWorkdir("build/testing/testdata/bundle")

		var err error
		container, err = assertExec(ctx, container, flipt("bundle", "build", "mybundle:latest"),
			stdout(matches(`sha256:[a-f0-9]{64}`)))
		if err != nil {
			return err
		}

		container, err = assertExec(ctx, container, flipt("bundle", "list"),
			stdout(matches(`DIGEST[\s]+REPO[\s]+TAG[\s]+CREATED`)),
			stdout(matches(`[a-f0-9]{7}[\s]+mybundle[\s]+latest[\s]+[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}`)))
		if err != nil {
			return err
		}

		// we need to wait for a second to ensure the timestamp ticks over so that
		// a new digest is created on the second build
		time.Sleep(time.Second)

		// rebuild the same image
		container, err = assertExec(ctx, container, flipt("bundle", "build", "mybundle:latest"),
			stdout(matches(`sha256:[a-f0-9]{64}`)))
		if err != nil {
			return err
		}

		// image has been rebuilt and now there are two
		container, err = assertExec(ctx, container, flipt("bundle", "list"),
			stdout(matches(`DIGEST[\s]+REPO[\s]+TAG[\s]+CREATED`)),
			stdout(matches(`[a-f0-9]{7}[\s]+mybundle[\s]+latest[\s]+[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}`)),
			stdout(matches(`[a-f0-9]{7}[\s]+mybundle[\s]+[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}`)))
		if err != nil {
			return err
		}

		// push image to itself at a different tag
		container, err = assertExec(ctx, container, flipt("bundle", "push", "mybundle:latest", "myotherbundle:latest"),
			stdout(matches(`sha256:[a-f0-9]{64}`)))
		if err != nil {
			return err
		}

		// now there are three
		container, err = assertExec(ctx, container, flipt("bundle", "list"),
			stdout(matches(`DIGEST[\s]+REPO[\s]+TAG[\s]+CREATED`)),
			stdout(matches(`[a-f0-9]{7}[\s]+myotherbundle[\s]+latest[\s]+[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}`)),
			stdout(matches(`[a-f0-9]{7}[\s]+mybundle[\s]+latest[\s]+[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}`)),
			stdout(matches(`[a-f0-9]{7}[\s]+mybundle[\s]+[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}`)))
		if err != nil {
			return err
		}

		platform, err := client.DefaultPlatform(ctx)
		if err != nil {
			return err
		}

		// switch out zot images based on host platform
		image := fmt.Sprintf("ghcr.io/project-zot/zot-linux-%s:latest",
			platforms.MustParse(string(platform)).Architecture)

		// push to remote name
		container, err = assertExec(ctx,
			container.WithServiceBinding("zot",
				client.Container().
					From(image).
					WithExposedPort(5000).
					AsService()),
			flipt("bundle", "push", "mybundle:latest", "http://zot:5000/myremotebundle:latest"),
			stdout(matches(`sha256:[a-f0-9]{64}`)),
		)
		if err != nil {
			return err
		}

		// pull remote bundle
		container, err = assertExec(ctx,
			container.WithServiceBinding("zot",
				client.Container().
					From(image).
					WithExposedPort(5000).
					AsService()),
			flipt("bundle", "pull", "http://zot:5000/myremotebundle:latest"),
			stdout(matches(`sha256:[a-f0-9]{64}`)),
		)
		if err != nil {
			return err
		}

		// now there are four including local copy of remote name
		_, err = assertExec(ctx, container, flipt("bundle", "list"),
			stdout(matches(`DIGEST[\s]+REPO[\s]+TAG[\s]+CREATED`)),
			stdout(matches(`[a-f0-9]{7}[\s]+mybundle[\s]+latest[\s]+[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}`)),
			stdout(matches(`[a-f0-9]{7}[\s]+mybundle[\s]+[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}`)),
			stdout(matches(`[a-f0-9]{7}[\s]+myotherbundle[\s]+latest[\s]+[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}`)),
			stdout(matches(`[a-f0-9]{7}[\s]+myremotebundle[\s]+latest[\s]+[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}`)))
		if err != nil {
			return err
		}
	}

	{
		// TODO: add tests for flipt cloud commands
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
    value: buzz
`

	expectedYAMLStreamOutput = `version: "1.4"
namespace:
  key: default
  name: Default
  description: Default namespace
---
namespace:
  key: foo
  name: foo
flags:
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
    value: buzz
  match_type: ALL_MATCH_TYPE
---
namespace:
  key: bar
  name: bar
flags:
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
    value: buzz
  match_type: ALL_MATCH_TYPE
`
)
