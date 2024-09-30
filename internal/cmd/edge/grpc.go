package edge

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.flipt.io/flipt/internal/server/ofrep"

	otlpRuntime "go.opentelemetry.io/contrib/instrumentation/runtime"

	"go.opentelemetry.io/contrib/propagators/autoprop"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/info"
	"go.flipt.io/flipt/internal/metrics"
	authnmiddlewaregrpc "go.flipt.io/flipt/internal/server/authn/middleware/grpc"
	"go.flipt.io/flipt/internal/server/authz"
	authzbundle "go.flipt.io/flipt/internal/server/authz/engine/bundle"
	authzrego "go.flipt.io/flipt/internal/server/authz/engine/rego"
	authzmiddlewaregrpc "go.flipt.io/flipt/internal/server/authz/middleware/grpc"
	"go.flipt.io/flipt/internal/server/evaluation"
	evaluationdata "go.flipt.io/flipt/internal/server/evaluation/data"
	"go.flipt.io/flipt/internal/server/metadata"
	middlewaregrpc "go.flipt.io/flipt/internal/server/middleware/grpc"
	"go.flipt.io/flipt/internal/storage"
	fsstore "go.flipt.io/flipt/internal/storage/fs/store"
	"go.flipt.io/flipt/internal/tracing"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	grpc_health "google.golang.org/grpc/health/grpc_health_v1"
)

type grpcRegister interface {
	RegisterGRPC(*grpc.Server)
}

type grpcRegisterers []grpcRegister

func (g *grpcRegisterers) Add(r grpcRegister) {
	*g = append(*g, r)
}

func (g grpcRegisterers) RegisterGRPC(s *grpc.Server) {
	for _, register := range g {
		register.RegisterGRPC(s)
	}
}

// GRPCServer configures the dependencies associated with the Flipt GRPC Service.
// It provides an entrypoint to start serving the gRPC stack (Run()).
// Along with a teardown function (Shutdown(ctx)).
type GRPCServer struct {
	*grpc.Server

	logger *zap.Logger
	cfg    *config.Config
	ln     net.Listener

	shutdownFuncs []func(context.Context) error
}

// NewGRPCServer constructs the core Flipt gRPC service including its dependencies
// (e.g. tracing, metrics, and cleanup).
// It returns an instance of *GRPCServer which callers can Run().
func NewGRPCServer(
	ctx context.Context,
	logger *zap.Logger,
	cfg *config.Config,
	info info.Flipt,
) (*GRPCServer, error) {
	logger = logger.With(zap.String("server", "grpc"))
	server := &GRPCServer{
		logger: logger,
		cfg:    cfg,
	}

	var err error
	server.ln, err = net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GRPCPort))
	if err != nil {
		return nil, fmt.Errorf("creating grpc listener: %w", err)
	}

	server.onShutdown(func(context.Context) error {
		return server.ln.Close()
	})

	var store storage.Store

	store, err = fsstore.NewStore(ctx, logger, cfg)
	if err != nil {
		return nil, err
	}

	logger.Debug("store enabled", zap.Stringer("store", store))

	// Initialize metrics exporter if enabled
	if cfg.Metrics.Enabled {
		metricsResource, err := metrics.GetResources(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating metrics resource: %w", err)
		}

		metricExp, metricExpShutdown, err := metrics.GetExporter(ctx, &cfg.Metrics)
		if err != nil {
			return nil, fmt.Errorf("creating metrics exporter: %w", err)
		}

		server.onShutdown(metricExpShutdown)

		meterProvider := metricsdk.NewMeterProvider(
			metricsdk.WithResource(metricsResource),
			metricsdk.WithReader(metricExp),
		)
		otel.SetMeterProvider(meterProvider)
		server.onShutdown(meterProvider.Shutdown)

		// We only want to start the runtime metrics by open telemetry if the user have chosen
		// to use OTLP because the Prometheus endpoint already exposes those metrics.
		if cfg.Metrics.Exporter == config.MetricsOTLP {
			err = otlpRuntime.Start(otlpRuntime.WithMeterProvider(meterProvider))
			if err != nil {
				return nil, fmt.Errorf("starting runtime metric exporter: %w", err)
			}
		}

		logger.Debug("otel metrics enabled", zap.String("exporter", string(cfg.Metrics.Exporter)))
	}

	// Initialize tracingProvider regardless of configuration. No extraordinary resources
	// are consumed, or goroutines initialized until a SpanProcessor is registered.
	tracingProvider, err := tracing.NewProvider(ctx, info.Version, cfg.Tracing)
	if err != nil {
		return nil, err
	}
	server.onShutdown(func(ctx context.Context) error {
		return tracingProvider.Shutdown(ctx)
	})

	if cfg.Tracing.Enabled {
		exp, traceExpShutdown, err := tracing.GetExporter(ctx, &cfg.Tracing)
		if err != nil {
			return nil, fmt.Errorf("creating tracing exporter: %w", err)
		}

		server.onShutdown(traceExpShutdown)

		tracingProvider.RegisterSpanProcessor(tracesdk.NewBatchSpanProcessor(exp, tracesdk.WithBatchTimeout(1*time.Second)))

		logger.Debug("otel tracing enabled", zap.String("exporter", cfg.Tracing.Exporter.String()))
	}

	// base inteceptors
	interceptors := []grpc.UnaryServerInterceptor{
		grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			logger.Error("panic recovered", zap.Any("panic", p))
			return status.Errorf(codes.Internal, "%v", p)
		})),
		grpc_ctxtags.UnaryServerInterceptor(),
		grpc_zap.UnaryServerInterceptor(logger, grpc_zap.WithDecider(func(methodFullName string, err error) bool {
			// will not log gRPC calls if it was a call to healthcheck and no error was raised
			if err == nil && methodFullName == "/grpc.health.v1.Health/Check" {
				return false
			}
			return true
		})),
		grpc_prometheus.UnaryServerInterceptor,
		middlewaregrpc.ErrorUnaryInterceptor,
	}

	var (
		metasrv     = metadata.New(cfg, info)
		evalsrv     = evaluation.New(logger, store)
		evaldatasrv = evaluationdata.New(logger, store)
		healthsrv   = health.NewServer()
		ofrepsrv    = ofrep.New(logger, cfg.Cache, evalsrv)
	)

	var (
		// authnOpts is a slice of options that will be passed to the authentication service.
		// it's initialized with the default option of skipping authentication for the health service which should never require authentication.
		authnOpts = []containers.Option[authnmiddlewaregrpc.InterceptorOptions]{
			authnmiddlewaregrpc.WithServerSkipsAuthentication(healthsrv),
		}
		skipAuthnIfExcluded = func(server any, excluded bool) {
			if excluded {
				authnOpts = append(authnOpts, authnmiddlewaregrpc.WithServerSkipsAuthentication(server))
			}
		}
	)

	skipAuthnIfExcluded(evalsrv, cfg.Authentication.Exclude.Evaluation)
	skipAuthnIfExcluded(evaldatasrv, cfg.Authentication.Exclude.Evaluation)
	skipAuthnIfExcluded(ofrepsrv, cfg.Authentication.Exclude.OFREP)

	register, authInterceptors, err := authenticationGRPC(
		ctx,
		logger,
		cfg,
		authnOpts...,
	)
	if err != nil {
		return nil, err
	}

	// initialize servers
	register.Add(metasrv)
	register.Add(evalsrv)
	register.Add(evaldatasrv)
	register.Add(ofrepsrv)

	// forward internal gRPC logging to zap
	grpcLogLevel, err := zapcore.ParseLevel(cfg.Log.GRPCLevel)
	if err != nil {
		return nil, fmt.Errorf("parsing grpc log level (%q): %w", cfg.Log.GRPCLevel, err)
	}

	grpc_zap.ReplaceGrpcLoggerV2(logger.WithOptions(zap.IncreaseLevel(grpcLogLevel)))

	// add auth interceptors to the server
	interceptors = append(interceptors,
		append(authInterceptors,
			middlewaregrpc.FliptAcceptServerVersionUnaryInterceptor(logger),
			middlewaregrpc.EvaluationUnaryInterceptor(cfg.Analytics.Enabled()),
		)...,
	)

	server.onShutdown(func(ctx context.Context) error {
		return tracingProvider.Shutdown(ctx)
	})

	otel.SetTracerProvider(tracingProvider)

	textMapPropagator, err := autoprop.TextMapPropagator(getStringSlice(cfg.Tracing.Propagators)...)
	if err != nil {
		return nil, fmt.Errorf("error constructing tracing text map propagator: %w", err)
	}
	otel.SetTextMapPropagator(textMapPropagator)

	if cfg.Authorization.Required {
		authzOpts := []containers.Option[authzmiddlewaregrpc.InterceptorOptions]{
			authzmiddlewaregrpc.WithServerSkipsAuthorization(healthsrv),
		}

		var (
			authzEngine   authz.Verifier
			authzShutdown errFunc
			err           error
		)

		authzEngine, authzShutdown, err = getAuthz(ctx, logger, cfg)
		if err != nil {
			return nil, err
		}

		server.onShutdown(authzShutdown)

		interceptors = append(interceptors, authzmiddlewaregrpc.AuthorizationRequiredInterceptor(logger, authzEngine, authzOpts...))

		logger.Info("authorization middleware enabled")
	}

	// we validate requests before after authn and authz
	interceptors = append(interceptors, middlewaregrpc.ValidationUnaryInterceptor)

	grpcOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(interceptors...),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     cfg.Server.GRPCConnectionMaxIdleTime,
			MaxConnectionAge:      cfg.Server.GRPCConnectionMaxAge,
			MaxConnectionAgeGrace: cfg.Server.GRPCConnectionMaxAgeGrace,
		}),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	}

	if cfg.Server.Protocol == config.HTTPS {
		creds, err := credentials.NewServerTLSFromFile(cfg.Server.CertFile, cfg.Server.CertKey)
		if err != nil {
			return nil, fmt.Errorf("loading TLS credentials: %w", err)
		}

		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}

	// initialize grpc server
	server.Server = grpc.NewServer(grpcOpts...)
	grpc_health.RegisterHealthServer(server.Server, healthsrv)

	// register grpcServer graceful stop on shutdown
	server.onShutdown(func(context.Context) error {
		healthsrv.Shutdown()
		server.GracefulStop()
		return nil
	})

	// register each grpc service onto the grpc server
	register.RegisterGRPC(server.Server)

	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(server.Server)
	reflection.Register(server.Server)

	return server, nil
}

// Run begins serving gRPC requests.
// This methods blocks until Shutdown is called.
func (s *GRPCServer) Run() error {
	s.logger.Debug("starting grpc server")
	return s.Serve(s.ln)
}

// Shutdown tearsdown the entire gRPC stack including dependencies.
func (s *GRPCServer) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down GRPC server...")

	// call in reverse order to emulate pop semantics of a stack
	for i := len(s.shutdownFuncs) - 1; i >= 0; i-- {
		if fn := s.shutdownFuncs[i]; fn != nil {
			if err := fn(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

type errFunc func(context.Context) error

func (s *GRPCServer) onShutdown(fn errFunc) {
	s.shutdownFuncs = append(s.shutdownFuncs, fn)
}

var (
	authzOnce sync.Once
	validator authz.Verifier
	authzFunc errFunc = func(context.Context) error { return nil }
	authzErr  error
)

func getAuthz(ctx context.Context, logger *zap.Logger, cfg *config.Config) (authz.Verifier, errFunc, error) {
	authzOnce.Do(func() {
		var err error
		switch cfg.Authorization.Backend {
		case config.AuthorizationBackendLocal:
			validator, err = authzrego.NewEngine(ctx, logger, cfg)

		default:
			validator, err = authzbundle.NewEngine(ctx, logger, cfg)
		}

		if err != nil {
			authzErr = fmt.Errorf("creating authorization policy engine: %w", err)
			return
		}
	})

	return validator, authzFunc, authzErr
}

// getStringSlice receives any slice which the underline member type is "string"
// and return a new slice with the same members but transformed to "string" type.
// This is useful when we want to convert an enum slice of strings.
func getStringSlice[AnyString ~string, Slice []AnyString](slice Slice) []string {
	strSlice := make([]string, 0, len(slice))
	for _, anyString := range slice {
		strSlice = append(strSlice, string(anyString))
	}

	return strSlice
}
