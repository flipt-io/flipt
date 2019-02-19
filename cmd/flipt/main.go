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
	logLevel         string
	apiPort          int
	backendPort      int
	host             string
	dbName           string
	dbPath           string
	dbMigrationsPath string
	dbAutoMigrate    bool
}

func defaultConfig() *config {
	return &config{
		logLevel: "INFO",

		apiPort:     8080,
		backendPort: 9000,
		host:        "0.0.0.0",

		dbName:           "flipt",
		dbPath:           "/var/opt/flipt",
		dbMigrationsPath: "/etc/flipt/config/migrations",
		dbAutoMigrate:    true,
	}
}

const (
	cfgHost             = "host"
	cfgLogLevel         = "log.level"
	cfgAPIPort          = "api.port"
	cfgBackendPort      = "backend.port"
	cfgDBName           = "db.name"
	cfgDBPath           = "db.path"
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

	if viper.IsSet(cfgHost) {
		cfg.host = viper.GetString(cfgHost)
	}

	if viper.IsSet(cfgLogLevel) {
		cfg.logLevel = viper.GetString(cfgLogLevel)
	}

	if viper.IsSet(cfgAPIPort) {
		cfg.apiPort = viper.GetInt(cfgAPIPort)
	}

	if viper.IsSet(cfgBackendPort) {
		cfg.backendPort = viper.GetInt(cfgBackendPort)
	}

	if viper.IsSet(cfgDBName) {
		cfg.dbName = viper.GetString(cfgDBName)
	}

	if viper.IsSet(cfgDBPath) {
		cfg.dbPath = viper.GetString(cfgDBPath)
	}

	if viper.IsSet(cfgDBMigrationsPath) {
		cfg.dbMigrationsPath = viper.GetString(cfgDBMigrationsPath)
	}

	if viper.IsSet(cfgDBAutoMigrate) {
		cfg.dbAutoMigrate = viper.GetBool(cfgDBAutoMigrate)
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
		backendServer *grpc.Server
		apiServer     *http.Server
	)

	printHeader()

	if cfg.backendPort > 0 {
		g.Go(func() error {
			logger := logger.WithField("system", "backend")

			path := filepath.Clean(cfg.dbPath)
			db, err := sql.Open("sqlite3", fmt.Sprintf("%s/%s.db?%s", path, cfg.dbName, dbConnOpts))
			if err != nil {
				return errors.Wrap(err, "opening db")
			}

			defer db.Close()

			if cfg.dbAutoMigrate {
				logger.Info("running migrations...")

				driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
				if err != nil {
					return errors.Wrap(err, "getting db instance")
				}

				path := filepath.Clean(cfg.dbMigrationsPath)
				mm, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", path), "sqlite3", driver)
				if err != nil {
					return errors.Wrap(err, "opening migrations")
				}

				if err := mm.Up(); err != nil && err != migrate.ErrNoChange {
					return errors.Wrap(err, "running migrations")
				}

				logger.Info("finished migrations")
			}

			lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.host, cfg.backendPort))
			if err != nil {
				return errors.Wrap(err, "creating listener")
			}
			defer func() {
				_ = lis.Close()
			}()

			var opts []grpc.ServerOption

			opts = append(opts, grpc_middleware.WithUnaryServerChain(
				grpc_ctxtags.UnaryServerInterceptor(),
				grpc_logrus.UnaryServerInterceptor(logger),
				server.ErrorInterceptor,
				grpc_recovery.UnaryServerInterceptor(),
			))

			backendServer = grpc.NewServer(opts...)

			srv, err := server.New(logger, db)
			if err != nil {
				return errors.Wrap(err, "initializing server")
			}

			pb.RegisterFliptServer(backendServer, srv)

			logger.Infof("serving backend at: %s:%d", cfg.host, cfg.backendPort)
			return backendServer.Serve(lis)
		})
	}

	if cfg.apiPort > 0 {
		g.Go(func() error {
			logger := logger.WithField("system", "api")

			var (
				r    = chi.NewRouter()
				n    = negroni.New()
				api  = runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: false}))
				opts = []grpc.DialOption{grpc.WithInsecure()}
			)

			err := pb.RegisterFliptHandlerFromEndpoint(ctx, api, fmt.Sprintf("%s:%d", cfg.host, cfg.backendPort), opts)
			if err != nil {
				return errors.Wrap(err, "connecting to backend")
			}

			r.Handle("/docs/*", http.StripPrefix("/docs/", http.FileServer(swagger.Assets)))
			r.Handle("/api/v1/*", api)
			r.Handle("/*", http.FileServer(ui.Assets))

			n.Use(gzip.Gzip(gzip.DefaultCompression))

			n.Use(negroni.NewRecovery())
			n.UseHandler(r)

			apiServer = &http.Server{
				Addr:           fmt.Sprintf("%s:%d", cfg.host, cfg.apiPort),
				Handler:        n,
				ReadTimeout:    10 * time.Second,
				WriteTimeout:   10 * time.Second,
				MaxHeaderBytes: 1 << 20,
			}

			logger.Infof("serving api at: http://%s:%d", cfg.host, cfg.apiPort)
			return apiServer.ListenAndServe()
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
	if apiServer != nil {
		_ = apiServer.Shutdown(shutdownCtx)
	}

	if backendServer != nil {
		backendServer.Stop()
	}

	return g.Wait()
}
