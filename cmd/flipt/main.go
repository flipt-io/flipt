package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"text/template"
	"time"

	"github.com/blang/semver/v4"
	"github.com/fatih/color"
	"github.com/google/go-github/v32/github"
	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/cmd"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/info"
	"go.flipt.io/flipt/internal/storage/sql"
	"go.flipt.io/flipt/internal/telemetry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const devVersion = "dev"

var (
	cfg *config.Config

	cfgPath      string
	forceMigrate bool
	version      = devVersion
	commit       string
	date         string
	goVersion    = runtime.Version()
	analyticsKey string
	banner       string
)

func main() {
	var (
		once         sync.Once
		loggerConfig = zap.Config{
			Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
			Development: false,
			Encoding:    "console",
			EncoderConfig: zapcore.EncoderConfig{
				// Keys can be anything except the empty string.
				TimeKey:        "T",
				LevelKey:       "L",
				NameKey:        "N",
				CallerKey:      zapcore.OmitKey,
				FunctionKey:    zapcore.OmitKey,
				MessageKey:     "M",
				StacktraceKey:  zapcore.OmitKey,
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.CapitalColorLevelEncoder,
				EncodeTime:     zapcore.RFC3339TimeEncoder,
				EncodeDuration: zapcore.StringDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			},
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}
		l      *zap.Logger
		logger = func() *zap.Logger {
			once.Do(func() { l = zap.Must(loggerConfig.Build()) })
			return l
		}
	)

	defer func() {
		_ = logger().Sync()
	}()

	var (
		rootCmd = &cobra.Command{
			Use:     "flipt",
			Short:   "Flipt is a modern feature flag solution",
			Version: version,
			Run: func(cmd *cobra.Command, _ []string) {
				if err := run(cmd.Context(), logger()); err != nil {
					logger().Fatal("flipt", zap.Error(err))
				}
			},
			CompletionOptions: cobra.CompletionOptions{
				DisableDefaultCmd: true,
			},
		}

		exportCmd = &cobra.Command{
			Use:   "export",
			Short: "Export flags/segments/rules to file/stdout",
			Run: func(cmd *cobra.Command, _ []string) {
				if err := runExport(cmd.Context(), logger()); err != nil {
					logger().Fatal("export", zap.Error(err))
				}
			},
		}

		importCmd = &cobra.Command{
			Use:   "import",
			Short: "Import flags/segments/rules from file",
			Run: func(cmd *cobra.Command, args []string) {
				if err := runImport(cmd.Context(), logger(), args); err != nil {
					logger().Fatal("import", zap.Error(err))
				}
			},
		}

		migrateCmd = &cobra.Command{
			Use:   "migrate",
			Short: "Run pending database migrations",
			Run: func(cmd *cobra.Command, _ []string) {
				migrator, err := sql.NewMigrator(*cfg, logger())
				if err != nil {
					logger().Fatal("initializing migrator", zap.Error(err))
				}

				defer migrator.Close()

				if err := migrator.Up(true); err != nil {
					logger().Fatal("running migrator", zap.Error(err))
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
		logger().Fatal("executing template", zap.Error(err))
	}

	banner = buf.String()

	cobra.OnInitialize(func() {
		var err error

		// read in config
		cfg, err = config.Load(cfgPath)
		if err != nil {
			logger().Fatal("loading configuration", zap.Error(err))
		}

		// log to file if enabled
		if cfg.Log.File != "" {
			loggerConfig.OutputPaths = []string{cfg.Log.File}
		}

		// parse/set log level
		loggerConfig.Level, err = zap.ParseAtomicLevel(cfg.Log.Level)
		if err != nil {
			logger().Fatal("parsing log level", zap.String("level", cfg.Log.Level), zap.Error(err))
		}

		if cfg.Log.Encoding > config.LogEncodingConsole {
			loggerConfig.Encoding = cfg.Log.Encoding.String()

			// don't encode with colors if not using console log output
			loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		}
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
		logger().Fatal("execute", zap.Error(err))
	}
}

func run(ctx context.Context, logger *zap.Logger) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	var (
		isRelease = isRelease()
		isConsole = cfg.Log.Encoding == config.LogEncodingConsole

		updateAvailable bool
		cv, lv          semver.Version
	)

	if isConsole {
		color.Cyan("%s\n", banner)
	} else {
		logger.Info("flipt starting", zap.String("version", version), zap.String("commit", commit), zap.String("date", date), zap.String("go_version", goVersion))
	}

	if isRelease {
		var err error
		cv, err = semver.ParseTolerant(version)
		if err != nil {
			return fmt.Errorf("parsing version: %w", err)
		}
	}

	// print out any warnings from config parsing
	for _, warning := range cfg.Warnings {
		logger.Warn("configuration warning", zap.String("message", warning))
	}

	if cfg.Meta.CheckForUpdates && isRelease {
		logger.Debug("checking for updates")

		release, err := getLatestRelease(ctx)
		if err != nil {
			logger.Warn("getting latest release", zap.Error(err))
		}

		if release != nil {
			var err error
			lv, err = semver.ParseTolerant(release.GetTagName())
			if err != nil {
				return fmt.Errorf("parsing latest version: %w", err)
			}

			logger.Debug("version info", zap.Stringer("current_version", cv), zap.Stringer("latest_version", lv))

			switch cv.Compare(lv) {
			case 0:
				if isConsole {
					color.Green("You are currently running the latest version of Flipt [%s]!", cv)
				} else {
					logger.Info("running latest version", zap.Stringer("version", cv))
				}
			case -1:
				updateAvailable = true
				if isConsole {
					color.Yellow("A newer version of Flipt exists at %s, \nplease consider updating to the latest version.", release.GetHTMLURL())
				} else {
					logger.Info("newer version available", zap.Stringer("version", lv), zap.String("url", release.GetHTMLURL()))
				}
			}
		}
	}

	info := info.Flipt{
		Commit:          commit,
		BuildDate:       date,
		GoVersion:       goVersion,
		Version:         cv.String(),
		LatestVersion:   lv.String(),
		IsRelease:       isRelease,
		UpdateAvailable: updateAvailable,
	}

	if os.Getenv("CI") == "true" || os.Getenv("CI") == "1" {
		logger.Debug("CI detected, disabling telemetry")
		cfg.Meta.TelemetryEnabled = false
	}

	g, ctx := errgroup.WithContext(ctx)

	if err := initLocalState(); err != nil {
		logger.Debug("disabling telemetry, state directory not accessible", zap.String("path", cfg.Meta.StateDirectory), zap.Error(err))
		cfg.Meta.TelemetryEnabled = false
	} else {
		logger.Debug("local state directory exists", zap.String("path", cfg.Meta.StateDirectory))
	}

	if cfg.Meta.TelemetryEnabled && isRelease {
		logger := logger.With(zap.String("component", "telemetry"))

		g.Go(func() error {
			reporter, err := telemetry.NewReporter(*cfg, logger, analyticsKey, info)
			if err != nil {
				logger.Debug("initializing telemetry reporter", zap.Error(err))
				return nil
			}

			defer func() {
				_ = reporter.Shutdown()
			}()

			reporter.Run(ctx)
			return nil
		})
	}

	grpcServer, err := cmd.NewGRPCServer(ctx, logger, forceMigrate, cfg)
	if err != nil {
		return err
	}

	// starts grpc server
	g.Go(grpcServer.Run)

	// retrieve client connection to associated running gRPC server.
	conn, err := clientConn(ctx, cfg)
	if err != nil {
		return err
	}

	httpServer, err := cmd.NewHTTPServer(ctx, logger, cfg, conn, info)
	if err != nil {
		return err
	}

	// starts REST http(s) server
	g.Go(httpServer.Run)

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

	_ = httpServer.Shutdown(shutdownCtx)
	_ = grpcServer.Shutdown(shutdownCtx)

	return g.Wait()
}

func getLatestRelease(ctx context.Context) (*github.RepositoryRelease, error) {
	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetLatestRelease(ctx, "flipt-io", "flipt")
	if err != nil {
		return nil, fmt.Errorf("checking for latest version: %w", err)
	}

	return release, nil
}

func isRelease() bool {
	if version == "" || version == devVersion {
		return false
	}
	if strings.HasSuffix(version, "-snapshot") {
		return false
	}
	return true
}

// check if state directory already exists, create it if not
func initLocalState() error {
	if cfg.Meta.StateDirectory == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return fmt.Errorf("getting user config dir: %w", err)
		}
		cfg.Meta.StateDirectory = filepath.Join(configDir, "flipt")
	}

	fp, err := os.Stat(cfg.Meta.StateDirectory)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// state directory doesnt exist, so try to create it
			return os.MkdirAll(cfg.Meta.StateDirectory, 0700)
		}
		return fmt.Errorf("checking state directory: %w", err)
	}

	if fp != nil && !fp.IsDir() {
		return fmt.Errorf("state directory is not a directory")
	}

	// assume state directory exists and is a directory
	return nil
}

// clientConn constructs and configures a client connection to the underlying gRPC server.
func clientConn(ctx context.Context, cfg *config.Config) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{grpc.WithBlock()}
	switch cfg.Server.Protocol {
	case config.HTTPS:
		creds, err := credentials.NewClientTLSFromFile(cfg.Server.CertFile, "")
		if err != nil {
			return nil, fmt.Errorf("loading TLS credentials: %w", err)
		}

		opts = append(opts, grpc.WithTransportCredentials(creds))
	case config.HTTP:
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	dialCtx, dialCancel := context.WithTimeout(ctx, 5*time.Second)
	defer dialCancel()

	return grpc.DialContext(dialCtx,
		fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GRPCPort), opts...)
}
