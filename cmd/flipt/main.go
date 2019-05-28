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
)

var (
	logger = logrus.New()

	cfgPath      string
	printVersion bool

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

	rootCmd.Flags().BoolVar(&printVersion, "version", false, "print version info and exit")
	rootCmd.Flags().StringVar(&cfgPath, "config", "/etc/flipt/config/default.yml", "path to config file")

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err)
	}
}

func printVersionHeader() {
	color.Cyan("%s\nVersion: %s\nCommit: %s\nBuild Date: %s\nGo Version: %s\n\n", banner, version, commit, date, goVersion)
}

func infoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		if _, err = w.Write(out); err != nil {
			logger.WithError(err).Error("writing response")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func configHandler(cfg *config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal(cfg)
		if err != nil {
			logger.WithError(err).Error("getting config")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, err = w.Write(out); err != nil {
			logger.WithError(err).Error("writing response")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func execute() error {
	if printVersion {
		printVersionHeader()
		return nil
	}

	cfg, err := configure()
	if err != nil {
		return err
	}

	lvl, err := logrus.ParseLevel(cfg.LogLevel)
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

	printVersionHeader()

	if cfg.Server.GRPCPort > 0 {
		g.Go(func() error {
			logger := logger.WithField("server", "grpc")
			logger.Infof("connecting to database: %s", cfg.Database.URL)

			db, err := storage.Open(cfg.Database.URL)
			if err != nil {
				return errors.Wrap(err, "opening db")
			}

			defer db.Close()

			if cfg.Database.AutoMigrate {
				logger.Info("running migrations...")

				if err := db.Migrate(cfg.Database.MigrationsPath); err != nil {
					return errors.Wrap(err, "migrating database")
				}

				logger.Info("finished migrations")
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

			grpcOpts = append(grpcOpts, grpc_middleware.WithUnaryServerChain(
				grpc_ctxtags.UnaryServerInterceptor(),
				grpc_logrus.UnaryServerInterceptor(logger),
				srv.ErrorUnaryInterceptor,
				grpc_recovery.UnaryServerInterceptor(),
			))

			if cfg.Cache.Memory.Enabled {
				cache, err := lru.New(cfg.Cache.Memory.Items)
				if err != nil {
					return errors.Wrap(err, "creating in-memory cache")
				}

				logger.Infof("in-memory cache enabled with size: %d", cfg.Cache.Memory.Items)
				serverOpts = append(serverOpts, server.WithCache(cache))
			}

			srv = server.New(logger, db, serverOpts...)
			grpcServer = grpc.NewServer(grpcOpts...)

			pb.RegisterFliptServer(grpcServer, srv)

			logger.Infof("grpc server running at: %s:%d", cfg.Server.Host, cfg.Server.GRPCPort)
			return grpcServer.Serve(lis)
		})
	}

	if cfg.Server.HTTPPort > 0 {
		g.Go(func() error {
			logger := logger.WithField("server", "http")

			var (
				r    = chi.NewRouter()
				api  = grpc_gateway.NewServeMux(grpc_gateway.WithMarshalerOption(grpc_gateway.MIMEWildcard, &grpc_gateway.JSONPb{OrigName: false}))
				opts = []grpc.DialOption{grpc.WithInsecure()}
			)

			if err := pb.RegisterFliptHandlerFromEndpoint(ctx, api, fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GRPCPort), opts); err != nil {
				return errors.Wrap(err, "connecting to grpc server")
			}

			r.Use(middleware.RequestID)
			r.Use(middleware.RealIP)
			r.Use(middleware.Compress(gzip.DefaultCompression))
			r.Use(middleware.Heartbeat("/health"))
			r.Use(middleware.Recoverer)

			r.Mount("/api/v1", api)
			r.Mount("/debug", middleware.Profiler())

			r.Route("/meta", func(r chi.Router) {
				r.Use(middleware.SetHeader("Content-Type", "application/json"))
				r.Handle("/info", infoHandler())
				r.Handle("/config", configHandler(cfg))
			})

			if cfg.UI.Enabled {
				r.Mount("/docs", http.StripPrefix("/docs/", http.FileServer(swagger.Assets)))
				r.Mount("/", http.FileServer(ui.Assets))
			}

			httpServer = &http.Server{
				Addr:           fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.HTTPPort),
				Handler:        r,
				ReadTimeout:    10 * time.Second,
				WriteTimeout:   10 * time.Second,
				MaxHeaderBytes: 1 << 20,
			}

			logger.Infof("api server running at: http://%s:%d/api/v1", cfg.Server.Host, cfg.Server.HTTPPort)

			if cfg.UI.Enabled {
				logger.Infof("ui available at: http://%s:%d", cfg.Server.Host, cfg.Server.HTTPPort)
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
