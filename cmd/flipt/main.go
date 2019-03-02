package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"database/sql"

	"github.com/fatih/color"
	"github.com/go-chi/chi"
	_ "github.com/mattn/go-sqlite3"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/pkg/errors"
	"github.com/urfave/negroni"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	pb "github.com/markphelps/flipt/proto"
	"github.com/markphelps/flipt/server"
	"github.com/markphelps/flipt/swagger"
	"github.com/markphelps/flipt/ui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	migrate "github.com/golang-migrate/migrate"
	sqlite3 "github.com/golang-migrate/migrate/database/sqlite3"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/mattn/go-sqlite3"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
)

const (
	dbConnOpts = "cache=shared&_fk=true"
	banner     = `
    _____ _ _       _
   |  ___| (_)_ __ | |_
   | |_  | | | '_ \| __|
   |  _| | | | |_) | |_
   |_|   |_|_| .__/ \__|
             |_|
  `
)

var (
	logger = logrus.New()

	cfgPath string
	print   bool

	version = "dev"
	commit  = ""
	date    = time.Now().Format(time.RFC3339)
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
	cache    cacheConfig
	server   serverConfig
	database databaseConfig
}

type cacheConfig struct {
	enabled bool
	size    int
}

type serverConfig struct {
	host     string
	httpPort int
	grpcPort int
}

type databaseConfig struct {
	name           string
	path           string
	migrationsPath string
	autoMigrate    bool
}

func defaultConfig() *config {
	return &config{
		logLevel: "INFO",

		cache: cacheConfig{
			enabled: false,
			size:    250,
		},

		server: serverConfig{
			host:     "0.0.0.0",
			httpPort: 8080,
			grpcPort: 9000,
		},

		database: databaseConfig{
			name:           "flipt",
			path:           "/var/opt/flipt",
			migrationsPath: "/etc/flipt/config/migrations",
			autoMigrate:    true,
		},
	}
}

const (
	// Logging
	cfgLogLevel = "log.level"

	// Cache
	cfgCacheEnabled = "cache.enabled"
	cfgCacheSize    = "cache.size"

	// Server
	cfgServerHost     = "server.host"
	cfgServerHTTPPort = "server.http_port"
	cfgServerGRPCPort = "server.grpc_port"

	// DB
	cfgDBName           = "db.name"
	cfgDBPath           = "db.path"
	cfgDBMigrationsPath = "db.migrations.path"
	cfgDBAutoMigrate    = "db.migrations.auto"

	// Aliases
	cfgAliasHost        = "host"
	cfgAliasAPIPort     = "api.port"
	cfgAliasBackendPort = "backend.port"
)

func configure() (*config, error) {
	viper.SetEnvPrefix("FLIPT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetConfigFile(cfgPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "loading config")
	}

	// TODO: remove in 1.0 release
	viper.RegisterAlias(cfgAliasHost, cfgServerHost)
	viper.RegisterAlias(cfgAliasAPIPort, cfgServerHTTPPort)
	viper.RegisterAlias(cfgAliasBackendPort, cfgServerGRPCPort)

	cfg := defaultConfig()

	// Logging
	if viper.IsSet(cfgLogLevel) {
		cfg.logLevel = viper.GetString(cfgLogLevel)
	}

	// Cache
	if viper.IsSet(cfgCacheEnabled) {
		cfg.cache.enabled = viper.GetBool(cfgCacheEnabled)
	}
	if viper.IsSet(cfgCacheSize) {
		cfg.cache.size = viper.GetInt(cfgCacheSize)
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
	if viper.IsSet(cfgDBName) {
		cfg.database.name = viper.GetString(cfgDBName)
	}
	if viper.IsSet(cfgDBPath) {
		cfg.database.path = viper.GetString(cfgDBPath)
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
	color.Cyan("%s\nVersion: %s\nCommit: %s\nBuilt: %s\n\n", banner, version, commit, date)
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

			path := filepath.Clean(cfg.database.path)
			db, err := sql.Open("sqlite3", fmt.Sprintf("%s/%s.db?%s", path, cfg.database.name, dbConnOpts))
			if err != nil {
				return errors.Wrap(err, "opening db")
			}

			defer db.Close()

			if cfg.database.autoMigrate {
				logger.Info("running migrations...")

				driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
				if err != nil {
					return errors.Wrap(err, "getting db instance")
				}

				path := filepath.Clean(cfg.database.migrationsPath)
				mm, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", path), "sqlite3", driver)
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

			if cfg.cache.enabled {
				logger.Infof("flag cache enabled with size: %d", cfg.cache.size)
				serverOpts = append(serverOpts, server.WithCacheSize(cfg.cache.size))
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
				n    = negroni.New()
				api  = runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: false}))
				opts = []grpc.DialOption{grpc.WithInsecure()}
			)

			if err := pb.RegisterFliptHandlerFromEndpoint(ctx, api, fmt.Sprintf("%s:%d", cfg.server.host, cfg.server.grpcPort), opts); err != nil {
				return errors.Wrap(err, "connecting to grpc server")
			}

			r.Handle("/docs/*", http.StripPrefix("/docs/", http.FileServer(swagger.Assets)))
			r.Handle("/api/v1/*", api)
			r.Handle("/*", http.FileServer(ui.Assets))

			n.Use(gzip.Gzip(gzip.DefaultCompression))

			n.Use(negroni.NewRecovery())
			n.UseHandler(r)

			httpServer = &http.Server{
				Addr:           fmt.Sprintf("%s:%d", cfg.server.host, cfg.server.httpPort),
				Handler:        n,
				ReadTimeout:    10 * time.Second,
				WriteTimeout:   10 * time.Second,
				MaxHeaderBytes: 1 << 20,
			}

			logger.Infof("http server running at: http://%s:%d", cfg.server.host, cfg.server.httpPort)

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
