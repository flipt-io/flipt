package testing

import (
	"context"
	"fmt"

	_ "embed"

	"go.flipt.io/build/internal/dagger"
)

func redpandaTLSService(_ context.Context, client *dagger.Client, hostAlias, superuser string) (*dagger.Service, error) {
	key, cert, err := generateTLSCert(hostAlias)
	if err != nil {
		return nil, err
	}
	kafka := client.Container().
		From("redpandadata/redpanda:v24.1.13").
		WithNewFile("/etc/redpanda/.bootstrap.yaml", fmt.Sprintf(redpandaBoostrapConfigurationTpl, superuser)).
		WithNewFile("/etc/redpanda/redpanda.yaml", fmt.Sprintf(redpandaConfigurationTpl, hostAlias)).
		WithNewFile("/etc/redpanda/key.pem", string(key)).
		WithNewFile("/etc/redpanda/cert.pem", string(cert)).
		WithExposedPort(9092, dagger.ContainerWithExposedPortOpts{
			Description: "kafka endpoint",
		}).
		WithExposedPort(8081, dagger.ContainerWithExposedPortOpts{
			Description: "schema registry endpoint",
		}).
		WithExposedPort(9644, dagger.ContainerWithExposedPortOpts{
			Description: "admin api endpoint",
		}).
		WithExec([]string{"redpanda",
			"start",
			"--mode=dev-container",
			"--smp=1",
			"--overprovisioned",
			"--check=false",
			"--memory=200M"}, dagger.ContainerWithExecOpts{UseEntrypoint: true}).
		AsService()
	return kafka, nil
}

var redpandaBoostrapConfigurationTpl = `
superusers:
  - %s
kafka_enable_authorization: true
storage_min_free_bytes: 10485760
`

var redpandaConfigurationTpl = `
redpanda:
  developer_mode: true
  storage_min_free_bytes: 10485760
  admin:
    address: 0.0.0.0
    port: 9644

  kafka_api:
    - address: 0.0.0.0
      name: external
      port: 9092
      authentication_method: sasl
    - address: 127.0.0.1
      name: internal
      port: 29092
      authentication_method: sasl

  advertised_kafka_api:
    - address: %s
      name: external
      port: 9092
    - address: localhost
      name: internal
      port: 29092

  kafka_api_tls:
    - name: external
      enabled: true
      cert_file: /etc/redpanda/cert.pem
      key_file: /etc/redpanda/key.pem
      require_client_auth: false

schema_registry:
  schema_registry_api:
    - address: "0.0.0.0"
      name: main
      port: 8081
      authentication_method: sasl

schema_registry_client:
  brokers:
    - address: localhost
      port: 29092
rpk:
  tune_network: false
  tune_disk_scheduler: false
  tune_disk_nomerges: false
  tune_disk_irq: false
  tune_fstrim: false
  tune_cpu: false
  tune_aio_events: false
  tune_clocksource: false
  tune_swappiness: false
  tnable_memory_locking: false
  tune_coredump: false
  coredump_dir: "/var/lib/redpanda/coredump"
logger:
  level: WARN
`
