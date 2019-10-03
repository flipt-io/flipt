package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	lru "github.com/hashicorp/golang-lru"

	"github.com/fatih/color"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	_ "github.com/mattn/go-sqlite3"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	sq "github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/golang-migrate/migrate/database/sqlite3"
	grpc_gateway "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/markphelps/flipt/config"
	pb "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/server"
	"github.com/markphelps/flipt/storage"
	"github.com/markphelps/flipt/swagger"
	"github.com/markphelps/flipt/ui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
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

	dbMigrationVersion uint = 1
)

var (
	logger  = logrus.New()
	cfg     *config.Config
	logFile *os.File

	cfgPath      string
	printVersion bool

	version   = "dev"
	commit    = ""
	date      = time.Now().UTC().Format(time.RFC3339)
	goVersion = runtime.Version()
)

func main() {
	var (
		rootCmd = &cobra.Command{
			Use:   "flipt",
			Short: "Flipt is a self contained feature flag solution",
			Run: func(cmd *cobra.Command, args []string) {
				if err := execute(); err != nil {
					logger.Fatal(err)
				}
			},
		}

		migrateCmd = &cobra.Command{
			Use:   "migrate",
			Short: "Run pending database migrations",
			Run: func(cmd *cobra.Command, args []string) {
				if err := runMigrations(); err != nil {
					logger.Fatal(err)
				}
			},
		}
	)

	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "/etc/flipt/config/default.yml", "path to config file")
	rootCmd.Flags().BoolVar(&printVersion, "version", false, "print version info and exit")

	rootCmd.AddCommand(migrateCmd)
	cobra.OnInitialize(initConfig)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err)
	}

	logrus.Exit(0)
}

func initConfig() {
	// Do not load the config if the user just wants the version
	if !printVersion {
		var err error
		cfg, err = config.Load(cfgPath)
		if err != nil {
			logger.Fatal(errors.Wrap(err, "loading configuration"))
		}

		if err = setupLogger(cfg); err != nil {
			logger.Fatal(err)
		}
	}
}

func printVersionHeader() {
	color.Cyan("%s\nVersion: %s\nCommit: %s\nBuild Date: %s\nGo Version: %s\n", banner, version, commit, date, goVersion)
}

func runMigrations() error {
	db, driver, err := storage.Open(cfg.Database.URL)
	if err != nil {
		return errors.Wrap(err, "opening db")
	}

	defer db.Close()

	var dr database.Driver

	switch driver {
	case storage.SQLite:
		dr, err = sqlite3.WithInstance(db, &sqlite3.Config{})
	case storage.Postgres:
		dr, err = postgres.WithInstance(db, &postgres.Config{})
	}

	if err != nil {
		return errors.Wrapf(err, "getting db driver for: %s", driver)
	}

	f := filepath.Clean(fmt.Sprintf("%s/%s", cfg.Database.MigrationsPath, driver))

	mm, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", f), driver.String(), dr)
	if err != nil {
		return errors.Wrap(err, "opening migrations")
	}

	logger.Info("running migrations...")

	if err := mm.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	logger.Info("finished migrations")

	return nil
}

func execute() error {
	if printVersion {
		printVersionHeader()
		return nil
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	g, ctx := errgroup.WithContext(ctx)

	var (
		info = info{
			Version:   version,
			Commit:    commit,
			BuildDate: date,
			GoVersion: goVersion,
		}

		grpcServer *grpc.Server
		httpServer *http.Server
	)

	printVersionHeader()

	g.Go(func() error {
		logger := logger.WithField("server", "grpc")
		logger.Infof("connecting to database: %s", cfg.Database.URL)

		db, driver, err := storage.Open(cfg.Database.URL)
		if err != nil {
			return errors.Wrap(err, "opening db")
		}

		defer db.Close()

		var (
			builder sq.StatementBuilderType
			dr      database.Driver
		)

		switch driver {
		case storage.SQLite:
			builder = sq.StatementBuilder.RunWith(sq.NewStmtCacher(db))
			dr, err = sqlite3.WithInstance(db, &sqlite3.Config{})
		case storage.Postgres:
			builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(sq.NewStmtCacher(db))
			dr, err = postgres.WithInstance(db, &postgres.Config{})
		}

		if err != nil {
			return errors.Wrapf(err, "getting db driver for: %s", driver)
		}

		f := filepath.Clean(fmt.Sprintf("%s/%s", cfg.Database.MigrationsPath, driver))

		mm, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", f), driver.String(), dr)
		if err != nil {
			return errors.Wrap(err, "opening migrations")
		}

		v, _, err := mm.Version()
		if err != nil && err != migrate.ErrNilVersion {
			return errors.Wrap(err, "getting current migrations version")
		}

		// if first run, go ahead and run all migrations
		// otherwise exit and inform user to run manually
		if err == migrate.ErrNilVersion {
			logger.Debug("no previous migrations run; running now")
			if err := runMigrations(); err != nil {
				return errors.Wrap(err, "running migrations")
			}
		} else if v < dbMigrationVersion {
			logger.Debugf("migrations pending: current=%d, want=%d", v, dbMigrationVersion)
			return errors.New("migrations pending, please backup your database and run `flipt migrate`")
		}

		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GRPCPort))
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

		if cfg.Cache.Memory.Enabled {
			cache, err := lru.New(cfg.Cache.Memory.Items)
			if err != nil {
				return errors.Wrap(err, "creating in-memory cache")
			}

			logger.Infof("in-memory cache enabled with size: %d", cfg.Cache.Memory.Items)
			serverOpts = append(serverOpts, server.WithCache(cache))
		}

		srv = server.New(logger, builder, db, serverOpts...)

		grpcOpts = append(grpcOpts, grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_logrus.UnaryServerInterceptor(logger),
			grpc_prometheus.UnaryServerInterceptor,
			srv.ErrorUnaryInterceptor,
			grpc_recovery.UnaryServerInterceptor(),
		))

		if cfg.Server.Protocol == config.HTTPS {
			creds, err := credentials.NewServerTLSFromFile(cfg.Server.CertFile, cfg.Server.CertKey)
			if err != nil {
				return errors.Wrap(err, "loading TLS credentials")
			}

			grpcOpts = append(grpcOpts, grpc.Creds(creds))
		}

		grpcServer = grpc.NewServer(grpcOpts...)
		pb.RegisterFliptServer(grpcServer, srv)
		grpc_prometheus.Register(grpcServer)

		return grpcServer.Serve(lis)
	})

	g.Go(func() error {
		logger := logger.WithField("server", cfg.Server.Protocol.String())

		var (
			r        = chi.NewRouter()
			api      = grpc_gateway.NewServeMux(grpc_gateway.WithMarshalerOption(grpc_gateway.MIMEWildcard, &grpc_gateway.JSONPb{OrigName: false}))
			opts     = []grpc.DialOption{grpc.WithBlock()}
			httpPort int
		)

		switch cfg.Server.Protocol {
		case config.HTTPS:
			creds, err := credentials.NewClientTLSFromFile(cfg.Server.CertFile, "")
			if err != nil {
				return errors.Wrap(err, "loading TLS credentials")
			}

			opts = append(opts, grpc.WithTransportCredentials(creds))
			httpPort = cfg.Server.HTTPSPort
		case config.HTTP:
			opts = append(opts, grpc.WithInsecure())
			httpPort = cfg.Server.HTTPPort
		}

		dialCtx, dialCancel := context.WithTimeout(ctx, 5*time.Second)
		defer dialCancel()

		conn, err := grpc.DialContext(dialCtx, fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GRPCPort), opts...)
		if err != nil {
			return errors.Wrap(err, "connecting to grpc server")
		}

		if err := pb.RegisterFliptHandler(ctx, api, conn); err != nil {
			return errors.Wrap(err, "registering grpc gateway")
		}

		if cfg.Cors.Enabled {
			cors := cors.New(cors.Options{
				AllowedOrigins:   cfg.Cors.AllowedOrigins,
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
				ExposedHeaders:   []string{"Link"},
				AllowCredentials: true,
				MaxAge:           300,
			})

			r.Use(cors.Handler)
			logger.Debugf("CORS enabled with allowed origins: %v", cfg.Cors.AllowedOrigins)
		}

		r.Use(middleware.RequestID)
		r.Use(middleware.RealIP)
		r.Use(middleware.Compress(gzip.DefaultCompression))
		r.Use(middleware.Heartbeat("/health"))
		r.Use(middleware.Recoverer)
		r.Mount("/metrics", promhttp.Handler())
		r.Mount("/api/v1", api)
		r.Mount("/debug", middleware.Profiler())

		r.Route("/meta", func(r chi.Router) {
			r.Use(middleware.SetHeader("Content-Type", "application/json"))
			r.Handle("/info", info)
			r.Handle("/config", cfg)
		})

		if cfg.UI.Enabled {
			r.Mount("/docs", http.StripPrefix("/docs/", http.FileServer(swagger.Assets)))
			r.Mount("/", http.FileServer(ui.Assets))
		}

		httpServer = &http.Server{
			Addr:           fmt.Sprintf("%s:%d", cfg.Server.Host, httpPort),
			Handler:        r,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		logger.Infof("api server running at: %s://%s:%d/api/v1", cfg.Server.Protocol, cfg.Server.Host, httpPort)

		if cfg.UI.Enabled {
			logger.Infof("ui available at: %s://%s:%d", cfg.Server.Protocol, cfg.Server.Host, httpPort)
		}

		if cfg.Server.Protocol == config.HTTPS {
			httpServer.TLSConfig = &tls.Config{
				MinVersion:               tls.VersionTLS12,
				PreferServerCipherSuites: true,
				CipherSuites: []uint16{
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				},
			}

			httpServer.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))

			err = httpServer.ListenAndServeTLS(cfg.Server.CertFile, cfg.Server.CertKey)
		} else {
			err = httpServer.ListenAndServe()
		}

		if err != http.ErrServerClosed {
			return errors.Wrap(err, "http server")
		}

		logger.Info("server shutdown gracefully")
		return nil
	})

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

func setupLogger(cfg *config.Config) error {
	if err := setLogOutput(cfg); err != nil {
		return err
	}

	if err := setLogLevel(cfg); err != nil {
		return err
	}

	return nil
}

func setLogOutput(cfg *config.Config) error {
	if cfg.Log.File != "" {
		logFile, err := os.OpenFile(cfg.Log.File, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}

		logger.SetOutput(logFile)
		logrus.RegisterExitHandler(closeLogFile)
	} else {
		logger.SetOutput(os.Stdout)
	}

	return nil
}

func setLogLevel(cfg *config.Config) error {
	lvl, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		return err
	}

	logger.SetLevel(lvl)
	return nil
}

func closeLogFile() {
	if logFile != nil {
		logFile.Close()
	}
}

type info struct {
	Version   string `json:"version,omitempty"`
	Commit    string `json:"commit,omitempty"`
	BuildDate string `json:"buildDate,omitempty"`
	GoVersion string `json:"goVersion,omitempty"`
}

func (i info) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	out, err := json.Marshal(i)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(out); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
