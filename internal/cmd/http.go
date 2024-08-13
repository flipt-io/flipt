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

	"go.flipt.io/flipt/rpc/flipt/ofrep"

	"github.com/fatih/color"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/gorilla/csrf"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/gateway"
	"go.flipt.io/flipt/internal/info"
	"go.flipt.io/flipt/internal/server/authn/method"
	grpc_middleware "go.flipt.io/flipt/internal/server/middleware/grpc"
	http_middleware "go.flipt.io/flipt/internal/server/middleware/http"
	ofrep_middleware "go.flipt.io/flipt/internal/server/ofrep"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/analytics"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.flipt.io/flipt/rpc/flipt/meta"
	"go.flipt.io/flipt/ui"
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
	conn *grpc.ClientConn,
	info info.Flipt,
) (*HTTPServer, error) {
	logger = logger.With(zap.Stringer("server", cfg.Server.Protocol))

	var (
		server = &HTTPServer{
			logger: logger,
		}
		isConsole = cfg.Log.Encoding == config.LogEncodingConsole

		r               = chi.NewRouter()
		api             = gateway.NewGatewayServeMux(logger)
		evaluateAPI     = gateway.NewGatewayServeMux(logger)
		evaluateDataAPI = gateway.NewGatewayServeMux(logger, runtime.WithMetadata(grpc_middleware.ForwardFliptAcceptServerVersion), runtime.WithForwardResponseOption(http_middleware.HttpResponseModifier))
		analyticsAPI    = gateway.NewGatewayServeMux(logger)
		ofrepAPI        = gateway.NewGatewayServeMux(logger, runtime.WithErrorHandler(ofrep_middleware.ErrorHandler(logger)))
		httpPort        = cfg.Server.HTTPPort
	)

	if cfg.Server.Protocol == config.HTTPS {
		httpPort = cfg.Server.HTTPSPort
	}

	if err := flipt.RegisterFliptHandler(ctx, api, conn); err != nil {
		return nil, fmt.Errorf("registering grpc gateway: %w", err)
	}

	if err := evaluation.RegisterEvaluationServiceHandler(ctx, evaluateAPI, conn); err != nil {
		return nil, fmt.Errorf("registering grpc gateway: %w", err)
	}

	if err := evaluation.RegisterDataServiceHandler(ctx, evaluateDataAPI, conn); err != nil {
		return nil, fmt.Errorf("registering grpc gateway: %w", err)
	}

	if err := analytics.RegisterAnalyticsServiceHandler(ctx, analyticsAPI, conn); err != nil {
		return nil, fmt.Errorf("registering grpc gateway: %w", err)
	}

	if err := ofrep.RegisterOFREPServiceHandler(ctx, ofrepAPI, conn); err != nil {
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

	// set additional headers enabling the UI to be served securely
	// ie: Content-Security-Policy, X-Content-Type-Options, etc.
	for k, v := range ui.AdditionalHeaders() {
		r.Use(middleware.SetHeader(k, v))
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

		if key := cfg.Authentication.Session.CSRF.Key; key != "" {
			logger.Debug("enabling CSRF prevention")

			// skip csrf if the request does not set the origin header
			// for a potentially mutating http method.
			// This allows us to forgo CSRF for non-browser based clients.
			r.Use(func(handler http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodGet &&
						r.Method != http.MethodHead &&
						r.Header.Get("origin") == "" {
						r = csrf.UnsafeSkipCheck(r)
					}

					handler.ServeHTTP(w, r)
				})
			})
			r.Use(csrf.Protect([]byte(key), csrf.Path("/")))
		}

		r.Mount("/api/v1", api)
		r.Mount("/evaluate/v1", evaluateAPI)
		r.Mount("/internal/v1/analytics", analyticsAPI)
		r.Mount("/internal/v1", evaluateDataAPI)
		r.Mount("/ofrep", ofrepAPI)

		// mount all authentication related HTTP components
		// to the chi router.
		authenticationHTTPMount(ctx, logger, cfg.Authentication, r, conn)

		r.Group(func(r chi.Router) {
			r.Use(func(handler http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if cfg.Authentication.Session.CSRF.Key != "" {
						w.Header().Set("X-CSRF-Token", csrf.Token(r))
					}

					handler.ServeHTTP(w, r)
				})
			})

			// mount the metadata service to the chi router under /meta.
			r.Mount("/meta", runtime.NewServeMux(
				registerFunc(
					ctx,
					conn,
					meta.RegisterMetadataServiceHandler,
				),
				runtime.WithMetadata(method.ForwardPrefix),
			))
		})
	})

	// mount health endpoint to use the grpc health check service
	r.Mount("/health", runtime.NewServeMux(runtime.WithHealthEndpointAt(grpc_health_v1.NewHealthClient(conn), "/health")))

	fs, err := ui.FS()
	if err != nil {
		return nil, fmt.Errorf("mounting ui: %w", err)
	}

	r.Mount("/", http.FileServer(http.FS(fs)))

	server.Server = &http.Server{
		Addr:           fmt.Sprintf("%s:%d", cfg.Server.Host, httpPort),
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	logger.Debug("starting http server")

	var (
		apiAddr = fmt.Sprintf("%s://%s:%d/api/v1", cfg.Server.Protocol, cfg.Server.Host, httpPort)
		uiAddr  = fmt.Sprintf("%s://%s:%d", cfg.Server.Protocol, cfg.Server.Host, httpPort)
	)

	if isConsole {
		color.Green("\nAPI: %s", apiAddr)
		color.Green("UI: %s", uiAddr)

		fmt.Println()
	} else {
		logger.Info("api available", zap.String("address", apiAddr))
		logger.Info("ui available", zap.String("address", uiAddr))
	}

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
