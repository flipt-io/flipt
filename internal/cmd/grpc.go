package cmd

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	otlpRuntime "go.opentelemetry.io/contrib/instrumentation/runtime"

	"go.opentelemetry.io/contrib/propagators/autoprop"

	"github.com/fullstorydev/grpchan"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/info"
	"go.flipt.io/flipt/internal/metrics"
	serverfliptv1 "go.flipt.io/flipt/internal/server"
	analytics "go.flipt.io/flipt/internal/server/analytics"
	"go.flipt.io/flipt/internal/server/analytics/clickhouse"
	"go.flipt.io/flipt/internal/server/analytics/prometheus"
	authnmiddlewaregrpc "go.flipt.io/flipt/internal/server/authn/middleware/grpc"
	"go.flipt.io/flipt/internal/server/authz"
	authzbundle "go.flipt.io/flipt/internal/server/authz/engine/bundle"
	authzrego "go.flipt.io/flipt/internal/server/authz/engine/rego"
	authzmiddlewaregrpc "go.flipt.io/flipt/internal/server/authz/middleware/grpc"
	serverenvironments "go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/internal/server/evaluation"
	evaluationdata "go.flipt.io/flipt/internal/server/evaluation/data"
	"go.flipt.io/flipt/internal/server/metadata"
	middlewaregrpc "go.flipt.io/flipt/internal/server/middleware/grpc"
	"go.flipt.io/flipt/internal/server/ofrep"
	"go.flipt.io/flipt/internal/storage/environments"
	"go.flipt.io/flipt/internal/tracing"
	rpcflipt "go.flipt.io/flipt/rpc/flipt"
	rpcanalytics "go.flipt.io/flipt/rpc/flipt/analytics"
	rpcevaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	rpcmeta "go.flipt.io/flipt/rpc/flipt/meta"
	rpcoffrep "go.flipt.io/flipt/rpc/flipt/ofrep"
	rpcenv "go.flipt.io/flipt/rpc/v2/environments"
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

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	grpc_health "google.golang.org/grpc/health/grpc_health_v1"
)

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
// (e.g. tracing, metrics, storage, migrations, caching and cleanup).
// It returns an instance of *GRPCServer which callers can Run().
func NewGRPCServer(
	ctx context.Context,
	logger *zap.Logger,
	cfg *config.Config,
	ipch *inprocgrpc.Channel,
	info info.Flipt,
	forceMigrate bool,
) (*GRPCServer, error) {
	logger = logger.With(zap.String("server", "grpc"))
	server := &GRPCServer{
		logger: logger,
		cfg:    cfg,
	}

	// acts as a registry for all grpc services so they can be shared between
	// the grpc server and the in-process client connection
	handlers := &grpchan.HandlerMap{}

	var err error
	server.ln, err = net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GRPCPort))
	if err != nil {
		return nil, fmt.Errorf("creating grpc listener: %w", err)
	}

	server.onShutdown(func(context.Context) error {
		return server.ln.Close()
	})

	// configure a declarative backend store
	environmentStore, err := environments.NewStore(ctx, logger, cfg)
	if err != nil {
		return nil, err
	}

	logger.Debug("store enabled")

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
	tracingProvider, err := tracing.NewProvider(ctx, info.Build.Version, cfg.Tracing)
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

		logger.Debug("otel tracing enabled")
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
		//nolint:staticcheck // Deprecated but inprocgrpc does not support stats handlers
		otelgrpc.UnaryServerInterceptor(),
	}

	var (
		// legacy services
		metasrv     = metadata.New(cfg, info)
		evalsrv     = evaluation.New(logger, environmentStore)
		evaldatasrv = evaluationdata.New(logger, environmentStore)
		fliptv1srv  = serverfliptv1.New(logger, environmentStore)
		ofrepsrv    = ofrep.New(logger, evalsrv, nil)

		// health service
		healthsrv = health.NewServer()
	)

	envsrv, err := serverenvironments.NewServer(logger, environmentStore)
	if err != nil {
		return nil, fmt.Errorf("building configuration server: %w", err)
	}

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

	skipAuthnIfExcluded(fliptv1srv, cfg.Authentication.Exclude.Management)
	skipAuthnIfExcluded(evalsrv, cfg.Authentication.Exclude.Evaluation)
	skipAuthnIfExcluded(evaldatasrv, cfg.Authentication.Exclude.Evaluation)

	authInterceptors, authShutdown, err := authenticationGRPC(
		ctx,
		logger,
		cfg,
		handlers,
		authnOpts...,
	)
	if err != nil {
		return nil, err
	}

	server.onShutdown(authShutdown)

	if cfg.Analytics.Enabled() {
		if cfg.Analytics.Storage.Clickhouse.Enabled {
			client, err := clickhouse.New(logger, cfg, forceMigrate)
			if err != nil {
				return nil, fmt.Errorf("connecting to clickhouse: %w", err)
			}
			analyticsExporter := analytics.NewAnalyticsSinkSpanExporter(logger, client)
			tracingProvider.RegisterSpanProcessor(
				tracesdk.NewBatchSpanProcessor(
					analyticsExporter,
					tracesdk.WithBatchTimeout(cfg.Analytics.Buffer.FlushPeriod)),
			)
			server.onShutdown(func(ctx context.Context) error {
				return analyticsExporter.Shutdown(ctx)
			})
			rpcanalytics.RegisterAnalyticsServiceServer(handlers, analytics.New(logger, client))
			logger.Debug("analytics enabled", zap.String("database", client.String()), zap.String("flush_period", cfg.Analytics.Buffer.FlushPeriod.String()))
		} else if cfg.Analytics.Storage.Prometheus.Enabled {
			client, err := prometheus.New(logger, cfg)
			if err != nil {
				return nil, err
			}
			rpcanalytics.RegisterAnalyticsServiceServer(handlers, analytics.New(logger, client))
			logger.Debug("analytics enabled", zap.String("database", client.String()))
		}
	}

	// register servers
	rpcflipt.RegisterFliptServer(handlers, fliptv1srv)
	rpcenv.RegisterEnvironmentsServiceServer(handlers, envsrv)
	rpcmeta.RegisterMetadataServiceServer(handlers, metasrv)
	rpcevaluation.RegisterEvaluationServiceServer(handlers, evalsrv)
	rpcevaluation.RegisterDataServiceServer(handlers, evaldatasrv)
	rpcoffrep.RegisterOFREPServiceServer(handlers, ofrepsrv)

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

	// we validate requests after authn and authz
	interceptors = append(interceptors, middlewaregrpc.ValidationUnaryInterceptor)

	grpcOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(interceptors...),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     cfg.Server.GRPCConnectionMaxIdleTime,
			MaxConnectionAge:      cfg.Server.GRPCConnectionMaxAge,
			MaxConnectionAgeGrace: cfg.Server.GRPCConnectionMaxAgeGrace,
		}),
	}

	ipch.WithServerUnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...))

	if cfg.Server.Protocol == config.HTTPS {
		creds, err := credentials.NewServerTLSFromFile(cfg.Server.CertFile, cfg.Server.CertKey)
		if err != nil {
			return nil, fmt.Errorf("loading TLS credentials: %w", err)
		}

		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}

	ipch = ipch.WithServerUnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...))

	// initialize grpc server
	grpcServer := grpc.NewServer(grpcOpts...)
	grpc_health.RegisterHealthServer(handlers, healthsrv)

	// register grpc services onto the in-process client connection and the grpc server
	handlers.ForEach(ipch.RegisterService)
	handlers.ForEach(grpcServer.RegisterService)

	// register grpcServer graceful stop on shutdown
	server.onShutdown(func(context.Context) error {
		healthsrv.Shutdown()
		grpcServer.GracefulStop()
		return nil
	})

	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(grpcServer)
	reflection.Register(grpcServer)

	server.Server = grpcServer
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
