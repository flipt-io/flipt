package testing

import (
	"context"
	"fmt"

	_ "embed"

	"go.flipt.io/build/internal/dagger"
)

func redpandaTLSService(ctx context.Context, client *dagger.Client, hostAlias, superuser string) (*dagger.Service, error) {
	key, cert, err := generateTLSCert(hostAlias)
	if err != nil {
		return nil, err
	}
	kafka := client.Container().
		From("redpandadata/redpanda").
		WithNewFile("/etc/redpanda/.bootstrap.yaml", dagger.ContainerWithNewFileOpts{
			Contents: fmt.Sprintf(redpandaBoostrapConfigurationTpl, superuser),
		}).
		WithNewFile("/etc/redpanda/redpanda.yaml", dagger.ContainerWithNewFileOpts{
			Contents: fmt.Sprintf(redpandaConfigurationTpl, hostAlias),
		}).
		WithNewFile("/etc/redpanda/key.pem", dagger.ContainerWithNewFileOpts{
			Contents: string(key),
		}).
		WithNewFile("/etc/redpanda/cert.pem", dagger.ContainerWithNewFileOpts{
			Contents: string(cert),
		}).
		WithExposedPort(9092, dagger.ContainerWithExposedPortOpts{
			Description: "kafka endpoint",
		}).
		WithExposedPort(8081, dagger.ContainerWithExposedPortOpts{
			Description: "schema registry endpoint",
		}).
		WithExposedPort(9644, dagger.ContainerWithExposedPortOpts{
			Description: "admin api endpoint",
		}).
		AsService()
	return kafka, nil
}

var redpandaBoostrapConfigurationTpl = `
superusers:
  - %s
kafka_enable_authorization: true
`

var redpandaConfigurationTpl = `
redpanda:
  developer_mode: true
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
`
