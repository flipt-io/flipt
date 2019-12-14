package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"text/template"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/fatih/color"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/gobuffalo/packr"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/golang-migrate/migrate/database/sqlite3"
	grpc_gateway "github.com/grpc-ecosystem/grpc-gateway/runtime"
	lru "github.com/hashicorp/golang-lru"
	"github.com/markphelps/flipt/config"
	pb "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/server"
	"github.com/markphelps/flipt/storage/db"
	_ "github.com/mattn/go-sqlite3"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

const dbMigrationVersion uint = 2

var (
	logger = logrus.New()
	cfg    *config.Config

	cfgPath      string
	forceMigrate bool

	version   = "dev"
	commit    = ""
	date      = time.Now().UTC().Format(time.RFC3339)
	goVersion = runtime.Version()

	banner string
)

func main() {
	var (
		rootCmd = &cobra.Command{
			Use:     "flipt",
			Short:   "Flipt is a self contained feature flag solution",
			Version: version,
			Run: func(cmd *cobra.Command, args []string) {
				if err := run(); err != nil {
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

	var (
		t   = template.Must(template.New("banner").Parse(bannerTmpl))
		buf = new(bytes.Buffer)
	)

	if err := t.Execute(buf, &bannerOpts{
		Version:   version,
		Commit:    commit,
		Date:      date,
		GoVersion: goVersion,
	}); err != nil {
		logger.Fatal(fmt.Errorf("executing template: %w", err))
	}

	banner = buf.String()

	cobra.OnInitialize(initialize)

	rootCmd.SetVersionTemplate(banner)
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "/etc/flipt/config/default.yml", "path to config file")
	rootCmd.Flags().BoolVar(&forceMigrate, "force-migrate", false, "force migrations before running")
	_ = rootCmd.Flags().MarkHidden("force-migrate")

	rootCmd.AddCommand(migrateCmd)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err)
	}

	logrus.Exit(0)
}

func initialize() {
	var err error

	cfg, err = config.Load(cfgPath)
	if err != nil {
		logger.Fatal(fmt.Errorf("loading configuration: %w", err))
	}

	if err = setupLogger(cfg); err != nil {
		logger.Fatal(fmt.Errorf("setting up logger: %w", err))
	}
}

func runMigrations() error {
	sql, driver, err := db.Open(cfg.Database.URL)
	if err != nil {
		return fmt.Errorf("opening db: %w", err)
	}

	defer sql.Close()

	var dr database.Driver

	switch driver {
	case db.SQLite:
		dr, err = sqlite3.WithInstance(sql, &sqlite3.Config{})
	case db.Postgres:
		dr, err = postgres.WithInstance(sql, &postgres.Config{})
	}

	if err != nil {
		return fmt.Errorf("getting db driver for: %s: %w", driver, err)
	}

	f := filepath.Clean(fmt.Sprintf("%s/%s", cfg.Database.MigrationsPath, driver))

	mm, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", f), driver.String(), dr)
	if err != nil {
		return fmt.Errorf("opening migrations: %w", err)
	}

	logger.Info("running migrations...")

	if err := mm.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	logger.Info("finished migrations")

	return nil
}

func run() error {
	color.Cyan(banner)
	fmt.Println()

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

	g.Go(func() error {
		logger := logger.WithField("server", "grpc")
		logger.Debugf("connecting to database: %s", cfg.Database.URL)

		sql, driver, err := db.Open(cfg.Database.URL)
		if err != nil {
			return fmt.Errorf("opening db: %w", err)
		}

		defer sql.Close()

		var (
			builder    sq.StatementBuilderType
			dr         database.Driver
			stmtCacher = sq.NewStmtCacher(sql)
		)

		switch driver {
		case db.SQLite:
			builder = sq.StatementBuilder.RunWith(stmtCacher)
			dr, err = sqlite3.WithInstance(sql, &sqlite3.Config{})
		case db.Postgres:
			builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(stmtCacher)
			dr, err = postgres.WithInstance(sql, &postgres.Config{})
		}

		if err != nil {
			return fmt.Errorf("getting db driver for: %s: %w", driver, err)
		}

		f := filepath.Clean(fmt.Sprintf("%s/%s", cfg.Database.MigrationsPath, driver))

		mm, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", f), driver.String(), dr)
		if err != nil {
			return fmt.Errorf("opening migrations: %w", err)
		}

		v, _, err := mm.Version()
		if err != nil && err != migrate.ErrNilVersion {
			return fmt.Errorf("getting current migrations version: %w", err)
		}

		logger = logger.WithField("storage", driver.String())

		// if first run, go ahead and run all migrations
		// otherwise exit and inform user to run manually if migrations are pending
		if err == migrate.ErrNilVersion {
			logger.Debug("no previous migrations run; running now")
			if err := runMigrations(); err != nil {
				return fmt.Errorf("running migrations: %w", err)
			}
		} else if v < dbMigrationVersion {
			logger.Debugf("migrations pending: [current version=%d, want version=%d]", v, dbMigrationVersion)

			if forceMigrate {
				logger.Debugf("force-migrate set; running now")
				if err := runMigrations(); err != nil {
					return fmt.Errorf("running migrations: %w", err)
				}
			} else {
				return errors.New("migrations pending, please backup your database and run `flipt migrate`")
			}
		}

		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GRPCPort))
		if err != nil {
			return fmt.Errorf("creating grpc listener: %w", err)
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
				return fmt.Errorf("creating in-memory cache: %w", err)
			}

			logger.Debugf("in-memory cache enabled with size: %d", cfg.Cache.Memory.Items)
			serverOpts = append(serverOpts, server.WithCache(cache))
		}

		srv = server.New(logger, builder, sql, serverOpts...)

		grpcOpts = append(grpcOpts, grpc_middleware.WithUnaryServerChain(
			grpc_recovery.UnaryServerInterceptor(),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_logrus.UnaryServerInterceptor(logger),
			grpc_prometheus.UnaryServerInterceptor,
			srv.ErrorUnaryInterceptor,
			srv.ValidationUnaryInterceptor,
		))

		if cfg.Server.Protocol == config.HTTPS {
			creds, err := credentials.NewServerTLSFromFile(cfg.Server.CertFile, cfg.Server.CertKey)
			if err != nil {
				return fmt.Errorf("loading TLS credentials: %w", err)
			}

			grpcOpts = append(grpcOpts, grpc.Creds(creds))
		}

		grpcServer = grpc.NewServer(grpcOpts...)
		pb.RegisterFliptServer(grpcServer, srv)
		grpc_prometheus.Register(grpcServer)

		logger.Debug("starting grpc server")
		return grpcServer.Serve(lis)
	})

	g.Go(func() error {
		logger := logger.WithField("server", cfg.Server.Protocol.String())

		var (
			r        = chi.NewRouter()
			api      = grpc_gateway.NewServeMux(grpc_gateway.WithMarshalerOption(grpc_gateway.MIMEWildcard, &grpc_gateway.JSONPb{OrigName: false, EmitDefaults: true}))
			opts     = []grpc.DialOption{grpc.WithBlock()}
			httpPort int
		)

		switch cfg.Server.Protocol {
		case config.HTTPS:
			creds, err := credentials.NewClientTLSFromFile(cfg.Server.CertFile, "")
			if err != nil {
				return fmt.Errorf("loading TLS credentials: %w", err)
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
			return fmt.Errorf("connecting to grpc server: %w", err)
		}

		if err := pb.RegisterFliptHandler(ctx, api, conn); err != nil {
			return fmt.Errorf("registering grpc gateway: %w", err)
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
		r.Use(middleware.Heartbeat("/health"))
		r.Use(middleware.Compress(gzip.DefaultCompression))
		r.Use(middleware.Recoverer)
		r.Mount("/metrics", promhttp.Handler())
		r.Mount("/api/v1", api)
		r.Mount("/debug", middleware.Profiler())

		r.Route("/meta", func(r chi.Router) {
			r.Use(middleware.SetHeader("Content-Type", "application/json"))
			r.Handle("/info", info{
				Version:   version,
				Commit:    commit,
				BuildDate: date,
				GoVersion: goVersion,
			})
			r.Handle("/config", cfg)
		})

		if cfg.UI.Enabled {
			swagger := packr.NewBox("../../swagger")
			r.Mount("/docs", http.StripPrefix("/docs/", http.FileServer(swagger)))

			ui := packr.NewBox("../../ui/dist")
			r.Mount("/", http.FileServer(ui))
		}

		httpServer = &http.Server{
			Addr:           fmt.Sprintf("%s:%d", cfg.Server.Host, httpPort),
			Handler:        r,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   30 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		logger.Debug("starting http server")

		color.Green("\nAPI: %s://%s:%d/api/v1", cfg.Server.Protocol, cfg.Server.Host, httpPort)

		if cfg.UI.Enabled {
			color.Green("UI: %s://%s:%d", cfg.Server.Protocol, cfg.Server.Host, httpPort)
		}

		fmt.Println()

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
			return fmt.Errorf("http server: %w", err)
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
	logger.SetOutput(os.Stdout)

	if cfg.Log.File != "" {
		logFile, err := os.OpenFile(cfg.Log.File, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("opening log file: %s %w", cfg.Log.File, err)
		}

		logger.SetOutput(logFile)
		logrus.RegisterExitHandler(func() {
			if logFile != nil {
				_ = logFile.Close()
			}
		})
	}

	lvl, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		return fmt.Errorf("parsing log level: %s %w", cfg.Log.Level, err)
	}

	logger.SetLevel(lvl)
	return nil
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
}
