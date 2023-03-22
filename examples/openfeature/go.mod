module openfeature-example

go 1.20

require (
	github.com/gofrs/uuid v4.3.0+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/open-feature/go-sdk v1.0.0
	github.com/open-feature/go-sdk-contrib/hooks/open-telemetry v0.0.0-20221019164111-0c806c629a9d
	go.flipt.io/flipt-grpc v1.1.0
	go.flipt.io/flipt-openfeature-provider v0.1.3
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.36.4
	go.opentelemetry.io/otel v1.11.1
	go.opentelemetry.io/otel/exporters/jaeger v1.11.1
	go.opentelemetry.io/otel/sdk v1.11.1
	google.golang.org/grpc v1.50.1
)

require (
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.11.3 // indirect
	go.opentelemetry.io/otel/trace v1.11.1 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/genproto v0.0.0-20220930163606-c98284e70a91 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)
