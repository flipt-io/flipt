package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	lru "github.com/hashicorp/golang-lru"

	"database/sql"

	"github.com/fatih/color"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/pkg/errors"

	grpc_gateway "github.com/grpc-ecosystem/grpc-gateway/runtime"
	pb "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/server"
	"github.com/markphelps/flipt/swagger"
	"github.com/markphelps/flipt/ui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	migrate "github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	sqlite3 "github.com/golang-migrate/migrate/database/sqlite3"
	_ "github.com/golang-migrate/migrate/source/file"
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
	var rootCmd = &cobra.Command{
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

var (
	driverToString = map[databaseDriver]string{
		sqlite:   "sqlite3",
		postgres: "postgres",
	}

	stringToDriver = map[string]databaseDriver{
		"sqlite3":  sqlite,
		"postgres": postgres,
	}
)

type databaseDriver uint8

func (d databaseDriver) String() string {
	return driverToString[d]
}

const (
	_ databaseDriver = iota
	sqlite
	postgres
)

type databaseConfig struct {
	autoMigrate    bool
	driver         databaseDriver
	migrationsPath string
	uri            string
}

func defaultConfig() *config {
	return &config{
		logLevel: "INFO",

		ui: uiConfig{
			enabled: true,
		},

		cache: cacheConfig{
			memory: memoryCacheConfig{
				enabled: true,
			},
		},

		server: serverConfig{
			host:     "0.0.0.0",
			httpPort: 8080,
			grpcPort: 9000,
		},

		database: databaseConfig{
			driver:         sqlite,
			uri:            "/var/opt/flipt/flipt.db",
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
		driver, uri, err := parseDBURL(viper.GetString(cfgDBURL))
		if err != nil {
			return nil, err
		}

		cfg.database.driver = driver
		cfg.database.uri = uri

	}
	if viper.IsSet(cfgDBMigrationsPath) {
		cfg.database.migrationsPath = viper.GetString(cfgDBMigrationsPath)
	}
	if viper.IsSet(cfgDBAutoMigrate) {
		cfg.database.autoMigrate = viper.GetBool(cfgDBAutoMigrate)
	}

	return cfg, nil
}

func parseDBURL(url string) (databaseDriver, string, error) {
	parts := strings.SplitN(url, "://", 2)
	// TODO: check parts

	driver := stringToDriver[parts[0]]
	if driver == 0 {
		return 0, "", fmt.Errorf("unknown database driver: %s", parts[0])
	}

	uri := parts[1]

	switch driver {
	case sqlite:
		uri = fmt.Sprintf("%s?cache=shared&_fk=true", parts[1])
	case postgres:
		// TODO:
	}

	return driver, uri, nil
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

			logger.Infof("connecting to %s database: %s", cfg.database.driver, cfg.database.uri)

			db, err := sql.Open(cfg.database.driver.String(), cfg.database.uri)
			if err != nil {
				return errors.Wrap(err, "opening db")
			}

			defer db.Close()

			if cfg.database.autoMigrate {
				logger.Info("running migrations...")

				var driver database.Driver

				switch cfg.database.driver {
				case sqlite:
					driver, err = sqlite3.WithInstance(db, &sqlite3.Config{})
				case postgres:
					// TODO:
				}

				if err != nil {
					return errors.Wrap(err, "getting db driver for migrations")
				}

				path := filepath.Clean(fmt.Sprintf("%s/%s", cfg.database.migrationsPath, cfg.database.driver))

				// TODO: handle postgres too
				mm, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", path), cfg.database.driver.String(), driver)
				if err != nil {
					return errors.Wrap(err, "opening migrations")
				}

				if err := mm.Up(); err != nil && err != migrate.ErrNoChange {
					return errors.Wrap(err, "running migrations")
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

			srv = server.New(logger, db, serverOpts...)
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
