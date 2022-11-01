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
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

type response struct {
	RequestID string `json:"request_id"`
	User      string `json:"user"`
	Language  string `json:"language"`
	Greeting  string `json:"greeting"`
}

var (
	fliptServer  string
	jaegerServer string
	greetings    = map[string]string{
		"en": "Hello",
		"fr": "Bonjour",
		"es": "Hola",
		"de": "Hallo",
		"jp": "こんにちは",
	}
)

const (
	service     = "example-api"
	environment = "development"
)

func init() {
	flag.StringVar(&fliptServer, "server", "flipt:9000", "address of Flipt backend server")
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
	log.SetFlags(0)

	tp, err := tracerProvider(jaegerServer)
	if err != nil {
		log.Fatal(err)
	}

	defer tp.Shutdown(context.Background())

	// register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// set the opentelemetry hook
	openfeature.AddHooks(otelHook.NewHook())

	// setup Flipt OpenFeature provider with GRPC client
	provider := flipt.NewProvider(
		flipt.WithAddress(fliptServer),
		flipt.WithServiceType(flipt.ServiceTypeGRPC),
	)
	openfeature.SetProvider(provider)

	client := openfeature.NewClient(service + "-client")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/greeting", func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.Must(uuid.NewV4()).String()
		}

		key := r.URL.Query().Get("user")
		if key == "" {
			key = uuid.Must(uuid.NewV4()).String()
		}

		value, err := client.StringValue(r.Context(), "language", "en", openfeature.NewEvaluationContext(
			key,
			map[string]interface{}{},
		))

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("key: %s, language: %s", key, value)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response{
			RequestID: requestID,
			Language:  value,
			Greeting:  greetings[value],
			User:      key,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	log.Println("Flipt UI available at http://localhost:8080")
	log.Println("Demo API available at http://localhost:8000/api")
	log.Println("Jaeger UI available at http://localhost:16686")
	log.Print("\n -> run `curl -v http://localhost:8000/api/greeting?user={foo}`\n")
	log.Fatal(http.ListenAndServe(":8000", router))
}
