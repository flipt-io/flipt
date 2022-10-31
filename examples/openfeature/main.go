//go:build example
// +build example

package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	otelHook "github.com/open-feature/go-sdk-contrib/hooks/open-telemetry/pkg"
	"github.com/open-feature/go-sdk/pkg/openfeature"
	"go.flipt.io/flipt-openfeature-provider/pkg/provider/flipt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

type data struct {
	RequestID string `json:"request_id"`
	Language  string `json:"language"`
}

var (
	fliptServer  string
	jaegerServer string
)

const (
	service     = "example-api"
	environment = "development"
)

func init() {
	flag.StringVar(&fliptServer, "server", "http://flipt:8080", "address of Flipt backend server")
	flag.StringVar(&jaegerServer, "jaeger", "http://jaeger:14268/api/traces", "address of Jaeger server")
}

// tracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func tracerProvider(url string) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		// record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
			attribute.String("environment", environment),
		)),
	)
	return tp, nil
}

func main() {
	flag.Parse()

	tp, err := tracerProvider(jaegerServer)
	if err != nil {
		log.Fatal(err)
	}

	defer tp.Shutdown(context.Background())

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	// set the opentelemetry hook
	openfeature.AddHooks(otelHook.NewHook())
	openfeature.SetProvider(flipt.NewProvider(flipt.WithAddress(fliptServer)))

	client := openfeature.NewClient(service + "-client")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.Must(uuid.NewV4()).String()
		}

		value, err := client.StringValue(r.Context(), "language", "en", openfeature.NewEvaluationContext(
			requestID,
			map[string]interface{}{},
		))

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("requestID: %s, language: %s", requestID, value)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(data{
			RequestID: requestID,
			Language:  value,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	log.Println("Flip UI available at http://localhost:8080")
	log.Println("Demo API available at http://localhost:8000/api")
	log.Println("Jaeger UI available at http://localhost:16686")
	log.Print("\n -> run `curl -v http://localhost:8000/api`\n")
	log.Fatal(http.ListenAndServe(":8000", router))
}
