package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/info"
	fliptserver "go.flipt.io/flipt/internal/server"
	"go.flipt.io/flipt/internal/server/cache"
	"go.flipt.io/flipt/internal/server/cache/memory"
	"go.flipt.io/flipt/internal/server/cache/redis"
	"go.flipt.io/flipt/internal/server/metadata"
	middlewaregrpc "go.flipt.io/flipt/internal/server/middleware/grpc"
	fliptotel "go.flipt.io/flipt/internal/server/otel"
	"go.flipt.io/flipt/internal/storage"
	authsql "go.flipt.io/flipt/internal/storage/auth/sql"
	oplocksql "go.flipt.io/flipt/internal/storage/oplock/sql"
	"go.flipt.io/flipt/internal/storage/sql"
	"go.flipt.io/flipt/internal/storage/sql/mysql"
	"go.flipt.io/flipt/internal/storage/sql/postgres"
	"go.flipt.io/flipt/internal/storage/sql/sqlite"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	goredis_cache "github.com/go-redis/cache/v8"
	goredis "github.com/go-redis/redis/v8"
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
// (e.g. tracing, metrics, storage, migrations, caching and cleanup).
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

	db, driver, err := sql.Open(*cfg)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}

	if driver == sql.SQLite && cfg.Database.MaxOpenConn > 1 {
		logger.Warn("ignoring config.db.max_open_conn due to driver limitation (sqlite)", zap.Int("attempted_max_conn", cfg.Database.MaxOpenConn))
	}

	server.onShutdown(func(context.Context) error {
		return db.Close()
	})

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("pinging db: %w", err)
	}

	var store storage.Store

	switch driver {
	case sql.SQLite:
		store = sqlite.NewStore(db, logger)
	case sql.Postgres, sql.CockroachDB:
		store = postgres.NewStore(db, logger)
	case sql.MySQL:
		store = mysql.NewStore(db, logger)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}

	logger.Debug("store enabled", zap.Stringer("driver", driver))

	var tracingProvider = fliptotel.NewNoopProvider()

	if cfg.Tracing.Enabled {
		var exp tracesdk.SpanExporter

		switch cfg.Tracing.Exporter {
		case config.TracingJaeger:
			exp, err = jaeger.New(jaeger.WithAgentEndpoint(
				jaeger.WithAgentHost(cfg.Tracing.Jaeger.Host),
				jaeger.WithAgentPort(strconv.FormatInt(int64(cfg.Tracing.Jaeger.Port), 10)),
			))
		case config.TracingZipkin:
			exp, err = zipkin.New(cfg.Tracing.Zipkin.Endpoint)
		}

		if err != nil {
			return nil, err
		}

		tracingProvider = tracesdk.NewTracerProvider(
			tracesdk.WithBatcher(
				exp,
				tracesdk.WithBatchTimeout(1*time.Second),
			),
			tracesdk.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("flipt"),
				semconv.ServiceVersionKey.String(info.Version),
			)),
			tracesdk.WithSampler(tracesdk.AlwaysSample()),
		)

		logger.Debug("otel tracing enabled", zap.String("exporter", cfg.Tracing.Exporter.String()))
		server.onShutdown(func(ctx context.Context) error {
			return tracingProvider.Shutdown(ctx)
		})
	}

	otel.SetTracerProvider(tracingProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	var (
		sqlBuilder           = sql.BuilderFor(db, driver)
		authenticationStore  = authsql.NewStore(driver, sqlBuilder, logger)
		operationLockService = oplocksql.New(logger, driver, sqlBuilder)
	)

	register, authInterceptors, authShutdown, err := authenticationGRPC(
		ctx,
		logger,
		cfg.Authentication,
		authenticationStore,
		operationLockService,
	)
	if err != nil {
		return nil, err
	}

	server.onShutdown(authShutdown)

	// forward internal gRPC logging to zap
	grpcLogLevel, err := zapcore.ParseLevel(cfg.Log.GRPCLevel)
	if err != nil {
		return nil, fmt.Errorf("parsing grpc log level (%q): %w", cfg.Log.GRPCLevel, err)
	}

	grpc_zap.ReplaceGrpcLoggerV2(logger.WithOptions(zap.IncreaseLevel(grpcLogLevel)))

	// base observability inteceptors
	interceptors := append([]grpc.UnaryServerInterceptor{
		grpc_recovery.UnaryServerInterceptor(),
		grpc_ctxtags.UnaryServerInterceptor(),
		grpc_zap.UnaryServerInterceptor(logger),
		grpc_prometheus.UnaryServerInterceptor,
		otelgrpc.UnaryServerInterceptor(),
	},
		append(authInterceptors,
			middlewaregrpc.ErrorUnaryInterceptor,
			middlewaregrpc.ValidationUnaryInterceptor,
			middlewaregrpc.EvaluationUnaryInterceptor,
		)...,
	)

	if cfg.Cache.Enabled {
		var cacher cache.Cacher

		switch cfg.Cache.Backend {
		case config.CacheMemory:
			cacher = memory.NewCache(cfg.Cache)
		case config.CacheRedis:
			rdb := goredis.NewClient(&goredis.Options{
				Addr:     fmt.Sprintf("%s:%d", cfg.Cache.Redis.Host, cfg.Cache.Redis.Port),
				Password: cfg.Cache.Redis.Password,
				DB:       cfg.Cache.Redis.DB,
			})

			server.onShutdown(func(ctx context.Context) error {
				return rdb.Shutdown(ctx).Err()
			})

			status := rdb.Ping(ctx)
			if status == nil {
				return nil, errors.New("connecting to redis: no status")
			}

			if status.Err() != nil {
				return nil, fmt.Errorf("connecting to redis: %w", status.Err())
			}

			cacher = redis.NewCache(cfg.Cache, goredis_cache.New(&goredis_cache.Options{
				Redis: rdb,
			}))
		}

		interceptors = append(interceptors, middlewaregrpc.CacheUnaryInterceptor(cacher, logger))

		logger.Debug("cache enabled", zap.Stringer("backend", cacher))
	}

	grpcOpts := []grpc.ServerOption{grpc_middleware.WithUnaryServerChain(interceptors...)}

	if cfg.Server.Protocol == config.HTTPS {
		creds, err := credentials.NewServerTLSFromFile(cfg.Server.CertFile, cfg.Server.CertKey)
		if err != nil {
			return nil, fmt.Errorf("loading TLS credentials: %w", err)
		}

		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}

	// initialize server
	register.Add(fliptserver.New(logger, store))
	register.Add(metadata.NewServer(cfg, info))

	// initialize grpc server
	server.Server = grpc.NewServer(grpcOpts...)

	// register grpcServer graceful stop on shutdown
	server.onShutdown(func(context.Context) error {
		server.Server.GracefulStop()
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

	return s.Server.Serve(s.ln)
}

// Shutdown tearsdown the entire gRPC stack including dependencies.
func (s *GRPCServer) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down GRPC server...")

	// call in reverse order to emulate pop semantics of a stack
	for i := len(s.shutdownFuncs) - 1; i >= 0; i-- {
		if err := s.shutdownFuncs[i](ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *GRPCServer) onShutdown(fn func(context.Context) error) {
	s.shutdownFuncs = append(s.shutdownFuncs, fn)
}
