package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	lru "github.com/hashicorp/golang-lru"

	"github.com/fatih/color"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/pkg/errors"

	grpc_gateway "github.com/grpc-ecosystem/grpc-gateway/runtime"
	pb "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/server"
	"github.com/markphelps/flipt/storage"
	"github.com/markphelps/flipt/swagger"
	"github.com/markphelps/flipt/ui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
)

const (
	banner = `
    _____ _ _       _
   |  ___| (_)_ __ | |_
   | |_  | | | '_ \| __|
   |  _| | | | |_) | |_
   |_|   |_|_| .__/ \__|
             |_|
  `

	defaultMemoryCacheSize = 500
)

var (
	logger = logrus.New()

	cfgPath string
	print   bool

	version   = "dev"
	commit    = ""
	date      = time.Now().Format(time.RFC3339)
	goVersion = runtime.Version()
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "flipt",
		Short: "Flipt is a self contained feature flag solution",
		Run: func(cmd *cobra.Command, args []string) {
			if err := execute(); err != nil {
				logger.Fatal(err)
			}
		},
	}

	rootCmd.Flags().BoolVar(&print, "version", false, "print version info and exit")
	rootCmd.Flags().StringVar(&cfgPath, "config", "/etc/flipt/config/default.yml", "path to config file")

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err)
	}
}

type config struct {
	logLevel string
	ui       uiConfig
	cache    cacheConfig
	server   serverConfig
	database databaseConfig
}

type uiConfig struct {
	enabled bool
}

type memoryCacheConfig struct {
	enabled bool
}

type cacheConfig struct {
	memory memoryCacheConfig
}

type serverConfig struct {
	host     string
	httpPort int
	grpcPort int
}

type databaseConfig struct {
	autoMigrate    bool
	migrationsPath string
	url            string
}

func defaultConfig() *config {
	return &config{
		logLevel: "INFO",

		ui: uiConfig{
			enabled: true,
		},

		cache: cacheConfig{
			memory: memoryCacheConfig{
				enabled: false,
			},
		},

		server: serverConfig{
			host:     "0.0.0.0",
			httpPort: 8080,
			grpcPort: 9000,
		},

		database: databaseConfig{
			url:            "sqlite3:///var/opt/flipt/flipt.db",
			migrationsPath: "/etc/flipt/config/migrations",
			autoMigrate:    true,
		},
	}
}

const (
	// Logging
	cfgLogLevel = "log.level"

	// UI
	cfgUIEnabled = "ui.enabled"

	// Cache
	cfgCacheMemoryEnabled = "cache.memory.enabled"

	// Server
	cfgServerHost     = "server.host"
	cfgServerHTTPPort = "server.http_port"
	cfgServerGRPCPort = "server.grpc_port"

	// DB
	cfgDBURL            = "db.url"
	cfgDBMigrationsPath = "db.migrations.path"
	cfgDBAutoMigrate    = "db.migrations.auto"
)

func configure() (*config, error) {
	viper.SetEnvPrefix("FLIPT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetConfigFile(cfgPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "loading config")
	}

	cfg := defaultConfig()

	// Logging
	if viper.IsSet(cfgLogLevel) {
		cfg.logLevel = viper.GetString(cfgLogLevel)
	}

	// UI
	if viper.IsSet(cfgUIEnabled) {
		cfg.ui.enabled = viper.GetBool(cfgUIEnabled)
	}

	// Cache
	if viper.IsSet(cfgCacheMemoryEnabled) {
		cfg.cache.memory.enabled = viper.GetBool(cfgCacheMemoryEnabled)
	}

	// Server
	if viper.IsSet(cfgServerHost) {
		cfg.server.host = viper.GetString(cfgServerHost)
	}
	if viper.IsSet(cfgServerHTTPPort) {
		cfg.server.httpPort = viper.GetInt(cfgServerHTTPPort)
	}
	if viper.IsSet(cfgServerGRPCPort) {
		cfg.server.grpcPort = viper.GetInt(cfgServerGRPCPort)
	}

	// DB
	if viper.IsSet(cfgDBURL) {
		cfg.database.url = viper.GetString(cfgDBURL)
	}
	if viper.IsSet(cfgDBMigrationsPath) {
		cfg.database.migrationsPath = viper.GetString(cfgDBMigrationsPath)
	}
	if viper.IsSet(cfgDBAutoMigrate) {
		cfg.database.autoMigrate = viper.GetBool(cfgDBAutoMigrate)
	}

	return cfg, nil
}

func printHeader() {
	color.Cyan("%s\nVersion: %s\nCommit: %s\nBuild Date: %s\nGo Version: %s\n\n", banner, version, commit, date, goVersion)
}

func execute() error {
	if print {
		printHeader()
		return nil
	}

	cfg, err := configure()
	if err != nil {
		return err
	}

	lvl, err := logrus.ParseLevel(cfg.logLevel)
	if err != nil {
		return err
	}

	logger.SetLevel(lvl)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	g, ctx := errgroup.WithContext(ctx)

	var (
		grpcServer *grpc.Server
		httpServer *http.Server
	)

	printHeader()

	if cfg.server.grpcPort > 0 {
		g.Go(func() error {
			logger := logger.WithField("server", "grpc")
			logger.Infof("connecting to database: %s", cfg.database.url)

			store, err := storage.Open(cfg.database.url)
			if err != nil {
				return errors.Wrap(err, "opening db")
			}

			defer store.Close()

			if cfg.database.autoMigrate {
				logger.Info("running migrations...")

				if err := store.Migrate(cfg.database.migrationsPath); err != nil {
					return errors.Wrap(err, "migrating database")
				}

				logger.Info("finished migrations")
			}

			lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.server.host, cfg.server.grpcPort))
			if err != nil {
				return errors.Wrap(err, "creating grpc listener")
			}
			defer func() {
				_ = lis.Close()
			}()

			var (
				grpcOpts   []grpc.ServerOption
				serverOpts []server.Option
				srv        *server.Server
			)

			grpcOpts = append(grpcOpts, grpc_middleware.WithUnaryServerChain(
				grpc_ctxtags.UnaryServerInterceptor(),
				grpc_logrus.UnaryServerInterceptor(logger),
				srv.ErrorUnaryInterceptor,
				grpc_recovery.UnaryServerInterceptor(),
			))

			if cfg.cache.memory.enabled {
				logger.Infof("in-memory cache enabled")
				cache, err := lru.New(defaultMemoryCacheSize)
				if err != nil {
					return errors.Wrap(err, "creating in-memory cache")
				}

				serverOpts = append(serverOpts, server.WithCache(cache))
			}

			srv = server.New(logger, store, serverOpts...)
			grpcServer = grpc.NewServer(grpcOpts...)

			pb.RegisterFliptServer(grpcServer, srv)

			logger.Infof("grpc server running at: %s:%d", cfg.server.host, cfg.server.grpcPort)
			return grpcServer.Serve(lis)
		})
	}

	if cfg.server.httpPort > 0 {
		g.Go(func() error {
			logger := logger.WithField("server", "http")

			var (
				r    = chi.NewRouter()
				api  = grpc_gateway.NewServeMux(grpc_gateway.WithMarshalerOption(grpc_gateway.MIMEWildcard, &grpc_gateway.JSONPb{OrigName: false}))
				opts = []grpc.DialOption{grpc.WithInsecure()}
			)

			if err := pb.RegisterFliptHandlerFromEndpoint(ctx, api, fmt.Sprintf("%s:%d", cfg.server.host, cfg.server.grpcPort), opts); err != nil {
				return errors.Wrap(err, "connecting to grpc server")
			}

			r.Use(middleware.RequestID)
			r.Use(middleware.RealIP)
			r.Use(middleware.Compress(gzip.DefaultCompression))
			r.Use(middleware.Heartbeat("/health"))
			r.Use(middleware.Recoverer)

			r.Mount("/api/v1", api)
			r.Mount("/debug", middleware.Profiler())

			r.Handle("/meta/info", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				meta := struct {
					Version   string `json:"version,omitempty"`
					Commit    string `json:"commit,omitempty"`
					BuildDate string `json:"buildDate,omitempty"`
					GoVersion string `json:"goVersion,omitempty"`
				}{
					Version:   version,
					Commit:    commit,
					BuildDate: date,
					GoVersion: goVersion,
				}

				out, err := json.Marshal(meta)
				if err != nil {
					logger.WithError(err).Error("getting metadata")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")

				if _, err = w.Write(out); err != nil {
					logger.WithError(err).Error("writing response")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusOK)
			}))

			if cfg.ui.enabled {
				r.Mount("/docs", http.StripPrefix("/docs/", http.FileServer(swagger.Assets)))
				r.Mount("/", http.FileServer(ui.Assets))
			}

			httpServer = &http.Server{
				Addr:           fmt.Sprintf("%s:%d", cfg.server.host, cfg.server.httpPort),
				Handler:        r,
				ReadTimeout:    10 * time.Second,
				WriteTimeout:   10 * time.Second,
				MaxHeaderBytes: 1 << 20,
			}

			logger.Infof("api server running at: http://%s:%d/api/v1", cfg.server.host, cfg.server.httpPort)

			if cfg.ui.enabled {
				logger.Infof("ui available at: http://%s:%d", cfg.server.host, cfg.server.httpPort)
			}

			if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
				return err
			}

			return nil
		})
	}

	select {
	case <-interrupt:
		break
	case <-ctx.Done():
		break
	}

	logger.Info("shutting down...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if httpServer != nil {
		_ = httpServer.Shutdown(shutdownCtx)
	}

	if grpcServer != nil {
		grpcServer.GracefulStop()
	}

	return g.Wait()
}
