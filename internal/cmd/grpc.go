package cmd

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	"go.flipt.io/flipt/internal/cache"
	"go.flipt.io/flipt/internal/cache/memory"
	"go.flipt.io/flipt/internal/cache/redis"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/info"
	fliptserver "go.flipt.io/flipt/internal/server"
	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/internal/server/audit/logfile"
	"go.flipt.io/flipt/internal/server/audit/webhook"
	"go.flipt.io/flipt/internal/server/auth"
	"go.flipt.io/flipt/internal/server/evaluation"
	"go.flipt.io/flipt/internal/server/metadata"
	middlewaregrpc "go.flipt.io/flipt/internal/server/middleware/grpc"
	"go.flipt.io/flipt/internal/storage"
	storagecache "go.flipt.io/flipt/internal/storage/cache"
	"go.flipt.io/flipt/internal/storage/fs"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
	"go.flipt.io/flipt/internal/storage/sql/mysql"
	"go.flipt.io/flipt/internal/storage/sql/postgres"
	"go.flipt.io/flipt/internal/storage/sql/sqlite"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"go.flipt.io/flipt/internal/storage/fs/git"
	"go.flipt.io/flipt/internal/storage/fs/local"
	"go.flipt.io/flipt/internal/storage/fs/s3"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	goredis_cache "github.com/go-redis/cache/v9"
	goredis "github.com/redis/go-redis/v9"
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
	forceMigrate bool,
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

	switch cfg.Storage.Type {
	case "", config.DatabaseStorageType:
		db, builder, driver, dbShutdown, err := getDB(ctx, logger, cfg, forceMigrate)
		if err != nil {
			return nil, err
		}

		server.onShutdown(dbShutdown)

		switch driver {
		case fliptsql.SQLite:
			store = sqlite.NewStore(db, builder, logger)
		case fliptsql.Postgres, fliptsql.CockroachDB:
			store = postgres.NewStore(db, builder, logger)
		case fliptsql.MySQL:
			store = mysql.NewStore(db, builder, logger)
		default:
			return nil, fmt.Errorf("unsupported driver: %s", driver)
		}

		logger.Debug("database driver configured", zap.Stringer("driver", driver))
	case config.GitStorageType:
		opts := []containers.Option[git.Source]{
			git.WithRef(cfg.Storage.Git.Ref),
			git.WithPollInterval(cfg.Storage.Git.PollInterval),
		}

		auth := cfg.Storage.Git.Authentication
		switch {
		case auth.BasicAuth != nil:
			opts = append(opts, git.WithAuth(&http.BasicAuth{
				Username: auth.BasicAuth.Username,
				Password: auth.BasicAuth.Password,
			}))
		case auth.TokenAuth != nil:
			opts = append(opts, git.WithAuth(&http.TokenAuth{
				Token: auth.TokenAuth.AccessToken,
			}))
		}

		source, err := git.NewSource(logger, cfg.Storage.Git.Repository, opts...)
		if err != nil {
			return nil, err
		}

		store, err = fs.NewStore(logger, source)
		if err != nil {
			return nil, err
		}
	case config.LocalStorageType:
		source, err := local.NewSource(logger, cfg.Storage.Local.Path)
		if err != nil {
			return nil, err
		}

		store, err = fs.NewStore(logger, source)
		if err != nil {
			return nil, err
		}
	case config.ObjectStorageType:
		store, err = NewObjectStore(cfg, logger)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unexpected storage type: %q", cfg.Storage.Type)
	}

	logger.Debug("store enabled", zap.Stringer("type", store))

	// Initialize tracingProvider regardless of configuration. No extraordinary resources
	// are consumed, or goroutines initialized until a SpanProcessor is registered.
	var tracingProvider = tracesdk.NewTracerProvider(
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("flipt"),
			semconv.ServiceVersionKey.String(info.Version),
		)),
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
	)

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
		case config.TracingOTLP:
			// TODO: support additional configuration options
			client := otlptracegrpc.NewClient(
				otlptracegrpc.WithEndpoint(cfg.Tracing.OTLP.Endpoint),
				// TODO: support TLS
				otlptracegrpc.WithInsecure())
			exp, err = otlptrace.New(ctx, client)
		}

		if err != nil {
			return nil, fmt.Errorf("creating exporter: %w", err)
		}

		tracingProvider.RegisterSpanProcessor(tracesdk.NewBatchSpanProcessor(exp, tracesdk.WithBatchTimeout(1*time.Second)))

		logger.Debug("otel tracing enabled", zap.String("exporter", cfg.Tracing.Exporter.String()))
	}

	// base observability inteceptors
	interceptors := []grpc.UnaryServerInterceptor{
		grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			logger.Error("panic recovered", zap.Any("panic", p))
			return status.Errorf(codes.Internal, "%v", p)
		})),
		grpc_ctxtags.UnaryServerInterceptor(),
		grpc_zap.UnaryServerInterceptor(logger),
		grpc_prometheus.UnaryServerInterceptor,
		otelgrpc.UnaryServerInterceptor(),
	}

	var cacher cache.Cacher
	if cfg.Cache.Enabled {
		var (
			cacheShutdown errFunc
			err           error
		)
		cacher, cacheShutdown, err = getCache(ctx, cfg)

		if err != nil {
			return nil, err
		}

		server.onShutdown(cacheShutdown)

		store = storagecache.NewStore(store, cacher, logger)

		logger.Debug("cache enabled", zap.Stringer("backend", cacher))
	}

	var (
		fliptsrv           = fliptserver.New(logger, store)
		metasrv            = metadata.NewServer(cfg, info)
		evalsrv            = evaluation.New(logger, store)
		authOpts           = []containers.Option[auth.InterceptorOptions]{}
		skipAuthIfExcluded = func(server any, excluded bool) {
			if excluded {
				authOpts = append(authOpts, auth.WithServerSkipsAuthentication(server))
			}
		}
	)

	skipAuthIfExcluded(fliptsrv, cfg.Authentication.Exclude.Management)
	skipAuthIfExcluded(metasrv, cfg.Authentication.Exclude.Metadata)
	skipAuthIfExcluded(evalsrv, cfg.Authentication.Exclude.Evaluation)

	register, authInterceptors, authShutdown, err := authenticationGRPC(
		ctx,
		logger,
		cfg,
		forceMigrate,
		authOpts...,
	)
	if err != nil {
		return nil, err
	}

	server.onShutdown(authShutdown)

	// initialize server
	register.Add(fliptsrv)
	register.Add(metasrv)
	register.Add(evalsrv)

	// forward internal gRPC logging to zap
	grpcLogLevel, err := zapcore.ParseLevel(cfg.Log.GRPCLevel)
	if err != nil {
		return nil, fmt.Errorf("parsing grpc log level (%q): %w", cfg.Log.GRPCLevel, err)
	}

	grpc_zap.ReplaceGrpcLoggerV2(logger.WithOptions(zap.IncreaseLevel(grpcLogLevel)))

	// add auth interceptors to the server
	interceptors = append(interceptors,
		append(authInterceptors,
			middlewaregrpc.ErrorUnaryInterceptor,
			middlewaregrpc.ValidationUnaryInterceptor,
			middlewaregrpc.EvaluationUnaryInterceptor,
		)...,
	)

	// cache must come after auth interceptors
	if cfg.Cache.Enabled && cacher != nil {
		interceptors = append(interceptors, middlewaregrpc.CacheUnaryInterceptor(cacher, logger))
	}

	// audit sinks configuration
	sinks := make([]audit.Sink, 0)

	if cfg.Audit.Sinks.LogFile.Enabled {
		logFileSink, err := logfile.NewSink(logger, cfg.Audit.Sinks.LogFile.File)
		if err != nil {
			return nil, fmt.Errorf("opening file at path: %s", cfg.Audit.Sinks.LogFile.File)
		}

		sinks = append(sinks, logFileSink)
	}

	if cfg.Audit.Sinks.Webhook.Enabled {
		opts := []webhook.ClientOption{}
		if cfg.Audit.Sinks.Webhook.MaxBackoffDuration != 0 {
			opts = append(opts, webhook.WithMaxBackoffDuration(cfg.Audit.Sinks.Webhook.MaxBackoffDuration))
		}

		webhookSink := webhook.NewSink(logger, webhook.NewHTTPClient(logger, cfg.Audit.Sinks.Webhook.URL, cfg.Audit.Sinks.Webhook.SigningSecret, opts...))

		sinks = append(sinks, webhookSink)
	}

	// based on audit sink configuration from the user, provision the audit sinks and add them to a slice,
	// and if the slice has a non-zero length, add the audit sink interceptor
	if len(sinks) > 0 {
		checker, err := audit.NewChecker(cfg.Audit.Events)
		if err != nil {
			return nil, err
		}

		sse := audit.NewSinkSpanExporter(logger, sinks)
		tracingProvider.RegisterSpanProcessor(tracesdk.NewBatchSpanProcessor(sse, tracesdk.WithBatchTimeout(cfg.Audit.Buffer.FlushPeriod), tracesdk.WithMaxExportBatchSize(cfg.Audit.Buffer.Capacity)))

		interceptors = append(interceptors, middlewaregrpc.AuditUnaryInterceptor(logger, checker))
		logger.Debug("audit sinks enabled",
			zap.Stringers("sinks", sinks),
			zap.Int("buffer capacity", cfg.Audit.Buffer.Capacity),
			zap.String("flush period", cfg.Audit.Buffer.FlushPeriod.String()),
			zap.Strings("events", checker.Events()),
		)

		server.onShutdown(func(ctx context.Context) error {
			return sse.Shutdown(ctx)
		})
	}

	server.onShutdown(func(ctx context.Context) error {
		return tracingProvider.Shutdown(ctx)
	})

	otel.SetTracerProvider(tracingProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	grpcOpts := []grpc.ServerOption{grpc.ChainUnaryInterceptor(interceptors...)}

	if cfg.Server.Protocol == config.HTTPS {
		creds, err := credentials.NewServerTLSFromFile(cfg.Server.CertFile, cfg.Server.CertKey)
		if err != nil {
			return nil, fmt.Errorf("loading TLS credentials: %w", err)
		}

		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}

	// initialize grpc server
	server.Server = grpc.NewServer(grpcOpts...)

	// register grpcServer graceful stop on shutdown
	server.onShutdown(func(context.Context) error {
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

// NewObjectStore create a new storate.Store from the object config
func NewObjectStore(cfg *config.Config, logger *zap.Logger) (storage.Store, error) {
	objectCfg := cfg.Storage.Object
	var store storage.Store
	// keep this as a case statement in anticipation of
	// more object types in the future
	// nolint:gocritic
	switch objectCfg.Type {
	case config.S3ObjectSubStorageType:
		opts := []containers.Option[s3.Source]{
			s3.WithPollInterval(objectCfg.S3.PollInterval),
		}
		if objectCfg.S3.Endpoint != "" {
			opts = append(opts, s3.WithEndpoint(objectCfg.S3.Endpoint))
		}
		if objectCfg.S3.Region != "" {
			opts = append(opts, s3.WithRegion(objectCfg.S3.Region))
		}
		source, err := s3.NewSource(logger, objectCfg.S3.Bucket, opts...)
		if err != nil {
			return nil, err
		}
		store, err = fs.NewStore(logger, source)
		if err != nil {
			return nil, err
		}
	}
	return store, nil
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
	cacheOnce sync.Once
	cacher    cache.Cacher
	cacheFunc errFunc = func(context.Context) error { return nil }
	cacheErr  error
)

func getCache(ctx context.Context, cfg *config.Config) (cache.Cacher, errFunc, error) {
	cacheOnce.Do(func() {
		switch cfg.Cache.Backend {
		case config.CacheMemory:
			cacher = memory.NewCache(cfg.Cache)
		case config.CacheRedis:
			var tlsConfig *tls.Config
			if cfg.Cache.Redis.RequireTLS {
				tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12}
			}

			rdb := goredis.NewClient(&goredis.Options{
				Addr:            fmt.Sprintf("%s:%d", cfg.Cache.Redis.Host, cfg.Cache.Redis.Port),
				TLSConfig:       tlsConfig,
				Password:        cfg.Cache.Redis.Password,
				DB:              cfg.Cache.Redis.DB,
				PoolSize:        cfg.Cache.Redis.PoolSize,
				MinIdleConns:    cfg.Cache.Redis.MinIdleConn,
				ConnMaxIdleTime: cfg.Cache.Redis.ConnMaxIdleTime,
				DialTimeout:     cfg.Cache.Redis.NetTimeout,
				ReadTimeout:     cfg.Cache.Redis.NetTimeout * 2,
				WriteTimeout:    cfg.Cache.Redis.NetTimeout * 2,
				PoolTimeout:     cfg.Cache.Redis.NetTimeout * 2,
			})

			cacheFunc = func(ctx context.Context) error {
				return rdb.Shutdown(ctx).Err()
			}

			status := rdb.Ping(ctx)
			if status == nil {
				cacheErr = errors.New("connecting to redis: no status")
				return
			}

			if status.Err() != nil {
				cacheErr = fmt.Errorf("connecting to redis: %w", status.Err())
				return
			}

			cacher = redis.NewCache(cfg.Cache, goredis_cache.New(&goredis_cache.Options{
				Redis: rdb,
			}))
		}
	})

	return cacher, cacheFunc, cacheErr
}

var (
	dbOnce  sync.Once
	db      *sql.DB
	builder sq.StatementBuilderType
	driver  fliptsql.Driver
	dbFunc  errFunc = func(context.Context) error { return nil }
	dbErr   error
)

func getDB(ctx context.Context, logger *zap.Logger, cfg *config.Config, forceMigrate bool) (*sql.DB, sq.StatementBuilderType, fliptsql.Driver, errFunc, error) {
	dbOnce.Do(func() {
		migrator, err := fliptsql.NewMigrator(*cfg, logger)
		if err != nil {
			dbErr = err
			return
		}

		if err := migrator.Up(forceMigrate); err != nil {
			migrator.Close()
			dbErr = err
			return
		}

		migrator.Close()

		db, driver, err = fliptsql.Open(*cfg)
		if err != nil {
			dbErr = fmt.Errorf("opening db: %w", err)
			return
		}

		logger.Debug("constructing builder", zap.Bool("prepared_statements", cfg.Database.PreparedStatementsEnabled))

		builder = fliptsql.BuilderFor(db, driver, cfg.Database.PreparedStatementsEnabled)

		dbFunc = func(context.Context) error {
			return db.Close()
		}

		if driver == fliptsql.SQLite && cfg.Database.MaxOpenConn > 1 {
			logger.Warn("ignoring config.db.max_open_conn due to driver limitation (sqlite)", zap.Int("attempted_max_conn", cfg.Database.MaxOpenConn))
		}

		if err := db.PingContext(ctx); err != nil {
			dbErr = fmt.Errorf("pinging db: %w", err)
		}
	})

	return db, builder, driver, dbFunc, dbErr
}
