package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"text/template"
	"time"

	"github.com/blang/semver/v4"
	"github.com/fatih/color"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/gobuffalo/packr"
	"github.com/google/go-github/v32/github"
	grpc_gateway "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/markphelps/flipt/config"
	pb "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/server"
	"github.com/markphelps/flipt/storage"
	"github.com/markphelps/flipt/storage/cache"
	"github.com/markphelps/flipt/storage/db"
	"github.com/markphelps/flipt/storage/db/mysql"
	"github.com/markphelps/flipt/storage/db/postgres"
	"github.com/markphelps/flipt/storage/db/sqlite"
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

const defaultVersion = "dev"

var (
	l   = logrus.New()
	cfg *config.Config

	cfgPath      string
	forceMigrate bool

	version   = defaultVersion
	commit    string
	date      = time.Now().UTC().Format(time.RFC3339)
	goVersion = runtime.Version()

	banner string
)

func main() {
	var (
		rootCmd = &cobra.Command{
			Use:     "flipt",
			Short:   "Flipt is a modern feature flag solution",
			Version: version,
			Run: func(cmd *cobra.Command, args []string) {
				if err := run(args); err != nil {
					fmt.Println("error: ", err)
					logrus.Exit(1)
				}
			},
		}

		exportCmd = &cobra.Command{
			Use:   "export",
			Short: "Export flags/segments/rules to file/stdout",
			Run: func(cmd *cobra.Command, args []string) {
				if err := runExport(args); err != nil {
					fmt.Println("error: ", err)
					logrus.Exit(1)
				}
			},
		}

		importCmd = &cobra.Command{
			Use:   "import",
			Short: "Import flags/segments/rules from file",
			Run: func(cmd *cobra.Command, args []string) {
				if err := runImport(args); err != nil {
					fmt.Println("error: ", err)
					logrus.Exit(1)
				}
			},
		}

		migrateCmd = &cobra.Command{
			Use:   "migrate",
			Short: "Run pending database migrations",
			Run: func(cmd *cobra.Command, args []string) {
				migrator, err := db.NewMigrator(cfg, l)
				if err != nil {
					fmt.Println("error: ", err)
					logrus.Exit(1)
				}

				defer migrator.Close()

				if err := migrator.Run(true); err != nil {
					fmt.Println("error: ", err)
					logrus.Exit(1)
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
		fmt.Printf("error: executing template: %v", err)
		logrus.Exit(1)
	}

	banner = buf.String()

	cobra.OnInitialize(func() {
		var err error

		// read in config
		cfg, err = config.Load(cfgPath)
		if err != nil {
			fmt.Println("error: ", err)
			logrus.Exit(1)
		}

		l.SetOutput(os.Stdout)

		// log to file if enabled
		if cfg.Log.File != "" {
			logFile, err := os.OpenFile(cfg.Log.File, os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				fmt.Printf("error: opening log file: %s %v\n", cfg.Log.File, err)
				logrus.Exit(1)
			}

			l.SetOutput(logFile)
			logrus.RegisterExitHandler(func() {
				if logFile != nil {
					_ = logFile.Close()
				}
			})
		}

		// parse/set log level
		lvl, err := logrus.ParseLevel(cfg.Log.Level)
		if err != nil {
			fmt.Printf("error: parsing log level: %s %v\n", cfg.Log.Level, err)
			logrus.Exit(1)
		}

		l.SetLevel(lvl)
	})

	rootCmd.SetVersionTemplate(banner)
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "/etc/flipt/config/default.yml", "path to config file")
	rootCmd.Flags().BoolVar(&forceMigrate, "force-migrate", false, "force migrations before running")
	_ = rootCmd.Flags().MarkHidden("force-migrate")

	exportCmd.Flags().StringVarP(&exportFilename, "output", "o", "", "export to filename (default STDOUT)")
	importCmd.Flags().BoolVar(&dropBeforeImport, "drop", false, "drop database before import")
	importCmd.Flags().BoolVar(&importStdin, "stdin", false, "import from STDIN")

	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		logrus.Exit(1)
	}

	logrus.Exit(0)
}

func run(_ []string) error {
	color.Cyan(banner)
	fmt.Println()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	defer signal.Stop(interrupt)

	if cfg.Meta.CheckForUpdates && version != defaultVersion {
		checkForUpdates(ctx)
	}

	g, ctx := errgroup.WithContext(ctx)

	var (
		grpcServer *grpc.Server
		httpServer *http.Server
	)

	g.Go(func() error {
		logger := l.WithField("server", "grpc")
		logger.Debugf("connecting to database: %s", cfg.Database.URL)

		migrator, err := db.NewMigrator(cfg, l)
		if err != nil {
			return err
		}

		defer migrator.Close()

		if err := migrator.Run(forceMigrate); err != nil {
			return err
		}

		migrator.Close()

		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GRPCPort))
		if err != nil {
			return fmt.Errorf("creating grpc listener: %w", err)
		}

		defer func() {
			_ = lis.Close()
		}()

		var (
			grpcOpts []grpc.ServerOption
			srv      *server.Server
		)

		sql, driver, err := db.Open(*cfg)
		if err != nil {
			return fmt.Errorf("opening db: %w", err)
		}

		defer sql.Close()

		var store storage.Store

		switch driver {
		case db.SQLite:
			store = sqlite.NewStore(sql)
		case db.Postgres:
			store = postgres.NewStore(sql)
		case db.MySQL:
			store = mysql.NewStore(sql)
		}

		if cfg.Cache.Memory.Enabled {
			cacher := cache.NewInMemoryCache(cfg.Cache.Memory.Expiration, cfg.Cache.Memory.EvictionInterval, logger)
			if cfg.Cache.Memory.Expiration > 0 {
				logger.Infof("in-memory cache enabled [expiration: %v, evictionInterval: %v]", cfg.Cache.Memory.Expiration, cfg.Cache.Memory.EvictionInterval)
			} else {
				logger.Info("in-memory cache enabled with no expiration")
			}

			store = cache.NewStore(logger, cacher, store)
		}

		logger = logger.WithField("store", store.String())

		srv = server.New(logger, store)

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
		grpc_prometheus.EnableHandlingTimeHistogram()
		grpc_prometheus.Register(grpcServer)

		logger.Debug("starting grpc server")
		return grpcServer.Serve(lis)
	})

	g.Go(func() error {
		logger := l.WithField("server", cfg.Server.Protocol.String())

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
			logger.Infof("CORS enabled with allowed origins: %v", cfg.Cors.AllowedOrigins)
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

	l.Info("shutting down...")

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

func checkForUpdates(ctx context.Context) {
	l.Debug("checking for updates...")

	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetLatestRelease(ctx, "markphelps", "flipt")
	if err != nil {
		l.Warnf("error: checking for latest version: %v", err)
		return
	}

	var (
		releaseTag                    = release.GetTagName()
		latestVersion, currentVersion semver.Version
	)

	latestVersion, err = semver.ParseTolerant(releaseTag)
	if err != nil {
		l.Warnf("error: parsing latest version: %v", err)
		return
	}

	currentVersion, err = semver.ParseTolerant(version)
	if err != nil {
		l.Warnf("error: parsing current version: %v", err)
		return
	}

	l.Debugf("current version: %s; latest version: %s", currentVersion.String(), latestVersion.String())

	switch currentVersion.Compare(latestVersion) {
	case 0:
		l.Info("currently running the latest version of Flipt")
	case -1:
		l.Warnf("a newer version of Flipt exists at %s, please consider updating to the latest version", release.GetHTMLURL())
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
}
