package cmd

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	csrf "filippo.io/csrf/gorilla"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/gateway"
	"go.flipt.io/flipt/internal/info"
	"go.flipt.io/flipt/internal/server/authn/method"
	ofrep_middleware "go.flipt.io/flipt/internal/server/evaluation/ofrep"
	grpc_middleware "go.flipt.io/flipt/internal/server/middleware/grpc"
	http_middleware "go.flipt.io/flipt/internal/server/middleware/http"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.flipt.io/flipt/rpc/flipt/meta"
	"go.flipt.io/flipt/rpc/flipt/ofrep"
	"go.flipt.io/flipt/rpc/v2/analytics"
	"go.flipt.io/flipt/rpc/v2/environments"
	evaluationv2 "go.flipt.io/flipt/rpc/v2/evaluation"
	"go.flipt.io/flipt/ui"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// HTTPServer is a wrapper around the construction and registration of Flipt's HTTP server.
type HTTPServer struct {
	*http.Server

	logger *zap.Logger

	listenAndServe func() error
}

// NewHTTPServer constructs and configures the HTTPServer instance.
// The HTTPServer depends upon a running gRPC server instance which is why
// it explicitly requires and established gRPC connection as an argument.
func NewHTTPServer(
	ctx context.Context,
	logger *zap.Logger,
	cfg *config.Config,
	conn grpc.ClientConnInterface,
	info info.Flipt,
) (*HTTPServer, error) {
	logger = logger.With(zap.Stringer("server", cfg.Server.Protocol))

	var (
		server = &HTTPServer{
			logger: logger,
		}

		r   = chi.NewRouter()
		api = gateway.NewGatewayServeMux(logger,
			runtime.WithMetadata(grpc_middleware.ForwardFliptEnvironment))

		evaluateAPI = gateway.NewGatewayServeMux(logger,
			runtime.WithMetadata(grpc_middleware.ForwardFliptEnvironment))

		clientEvaluationAPI = gateway.NewGatewayServeMux(logger,
			runtime.WithMetadata(grpc_middleware.ForwardFliptEnvironment),
			runtime.WithForwardResponseOption(http_middleware.HttpResponseModifier),
		)

		analyticsAPI = gateway.NewGatewayServeMux(logger, runtime.WithMetadata(grpc_middleware.ForwardFliptEnvironment))

		ofrepAPI = gateway.NewGatewayServeMux(logger,
			runtime.WithMetadata(grpc_middleware.ForwardFliptEnvironment),
			runtime.WithMetadata(grpc_middleware.ForwardFliptNamespace),
			runtime.WithErrorHandler(ofrep_middleware.ErrorHandler(logger)),
		)

		environmentsAPI = gateway.NewGatewayServeMux(logger)

		httpPort = cfg.Server.HTTPPort
	)

	if cfg.Server.Protocol == config.HTTPS {
		httpPort = cfg.Server.HTTPSPort
	}

	// v1
	if err := flipt.RegisterFliptHandlerClient(ctx, api, flipt.NewFliptClient(conn)); err != nil {
		return nil, fmt.Errorf("registering grpc gateway: %w", err)
	}

	if err := evaluation.RegisterEvaluationServiceHandlerClient(ctx, evaluateAPI, evaluation.NewEvaluationServiceClient(conn)); err != nil {
		return nil, fmt.Errorf("registering grpc gateway: %w", err)
	}

	if err := ofrep.RegisterOFREPServiceHandlerClient(ctx, ofrepAPI, ofrep.NewOFREPServiceClient(conn)); err != nil {
		return nil, fmt.Errorf("registering grpc gateway: %w", err)
	}

	// v2
	if err := analytics.RegisterAnalyticsServiceHandlerClient(ctx, analyticsAPI, analytics.NewAnalyticsServiceClient(conn)); err != nil {
		return nil, fmt.Errorf("registering grpc gateway: %w", err)
	}

	if err := environments.RegisterEnvironmentsServiceHandlerClient(ctx, environmentsAPI, environments.NewEnvironmentsServiceClient(conn)); err != nil {
		return nil, fmt.Errorf("registering grpc gateway: %w", err)
	}

	if err := evaluationv2.RegisterClientEvaluationServiceHandlerClient(ctx, clientEvaluationAPI, evaluationv2.NewClientEvaluationServiceClient(conn)); err != nil {
		return nil, fmt.Errorf("registering grpc gateway: %w", err)
	}

	if cfg.Cors.Enabled {
		cors := cors.New(cors.Options{
			AllowedOrigins:   cfg.Cors.AllowedOrigins,
			AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
			AllowedHeaders:   cfg.Cors.AllowedHeaders,
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300,
		})

		r.Use(cors.Handler)
		logger.Debug("CORS enabled", zap.Strings("allowed_origins", cfg.Cors.AllowedOrigins))
	}

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// checking Values as map[string][]string also catches ?pretty and ?pretty=
			// r.URL.Query().Get("pretty") would not.
			if _, ok := r.URL.Query()["pretty"]; ok {
				r.Header.Set("Accept", "application/json+pretty")
			}

			h.ServeHTTP(w, r)
		})
	})
	r.Use(middleware.Compress(gzip.DefaultCompression))
	r.Use(http_middleware.HandleNoBodyResponse)
	r.Use(middleware.Recoverer)

	if cfg.Diagnostics.Profiling.Enabled {
		r.Mount("/debug", middleware.Profiler())
	}

	r.Mount("/metrics", promhttp.Handler())

	r.Group(func(r chi.Router) {
		r.Use(removeTrailingSlash)

		if cfg.Tracing.Enabled {
			r.Use(func(handler http.Handler) http.Handler {
				return otelhttp.NewHandler(handler, "grpc-gateway")
			})

			r.Use(func(handler http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ctx := r.Context()
					// Make sure headers contain the trace context set up by otelhttp for
					// later extraction by our custom grpc-gateway incoming header matcher.
					// Usually this would be taken care of by otelgrpc client interceptors,
					// but inprocgrpc does not provide client interceptors.
					otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))
					// Use the new context with the span
					r = r.WithContext(ctx)
					handler.ServeHTTP(w, r)
				})
			})
		}

		if key := cfg.Authentication.Session.CSRF.Key; key != "" {
			logger.Debug("enabling CSRF prevention")
			r.Use(csrf.Protect([]byte(key), csrf.TrustedOrigins(cfg.Authentication.Session.CSRF.TrustedOrigins)))
		}

		r.Mount("/api/v1", api)
		r.Mount("/evaluate/v1", evaluateAPI)
		r.Mount("/ofrep", ofrepAPI)
		r.Mount("/internal/v1", clientEvaluationAPI) // for backwards compatibility

		r.Mount("/internal/v2/analytics", analyticsAPI)
		r.Mount("/api/v2/environments", environmentsAPI)
		r.Mount("/client/v2", clientEvaluationAPI)

		// mount all authentication related HTTP components
		// to the chi router.
		authenticationHTTPMount(ctx, logger, cfg.Authentication, r, conn)

		r.Group(func(r chi.Router) {
			// mount the metadata service to the chi router under /meta.
			r.Mount("/meta", runtime.NewServeMux(
				register(
					ctx,
					meta.NewMetadataServiceClient(conn),
					meta.RegisterMetadataServiceHandlerClient,
				),
				runtime.WithMetadata(method.ForwardPrefix),
			))
		})
	})

	// mount health endpoint to use the grpc health check service
	r.Mount("/health", runtime.NewServeMux(runtime.WithHealthEndpointAt(grpc_health_v1.NewHealthClient(conn), "/health")))

	if cfg.UI.Enabled {
		fs, err := ui.FS()
		if err != nil {
			return nil, fmt.Errorf("mounting ui: %w", err)
		}

		r.With(func(next http.Handler) http.Handler {
			// set additional headers enabling the UI to be served securely
			// ie: Content-Security-Policy, X-Content-Type-Options, etc.
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for k, v := range ui.AdditionalHeaders() {
					w.Header().Set(k, v)
				}
				next.ServeHTTP(w, r)
			})
		}).Mount("/", http.FileServer(http.FS(fs)))
	}

	server.Server = &http.Server{
		Addr:           fmt.Sprintf("%s:%d", cfg.Server.Host, httpPort),
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	logger.Info("starting http server", zap.String("address", fmt.Sprintf("%s:%d", cfg.Server.Host, httpPort)))

	if cfg.Server.Protocol != config.HTTPS {
		server.listenAndServe = server.ListenAndServe
		return server, nil
	}

	server.TLSConfig = &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		},
	}

	server.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))

	server.listenAndServe = func() error {
		return server.ListenAndServeTLS(cfg.Server.CertFile, cfg.Server.CertKey)
	}

	return server, nil
}

// Run starts listening and serving the Flipt HTTP API.
// It blocks until the server is shutdown.
func (h *HTTPServer) Run() error {
	if err := h.listenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server: %w", err)
	}

	return nil
}

// Shutdown triggers the shutdown operation of the HTTP API.
func (h *HTTPServer) Shutdown(ctx context.Context) error {
	h.logger.Info("shutting down HTTP server...")

	return h.Server.Shutdown(ctx)
}

func removeTrailingSlash(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		h.ServeHTTP(w, r)
	})
}
