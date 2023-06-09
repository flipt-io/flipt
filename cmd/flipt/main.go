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
	"syscall"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/cmd"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/info"
	"go.flipt.io/flipt/internal/release"
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

const (
	defaultCfgPath = "/etc/flipt/config/default.yml"
)

var (
	cfgPath      string
	forceMigrate bool
	version      = "dev"
	commit       string
	date         string
	goVersion    = runtime.Version()
	analyticsKey string
	banner       string
)

var (
	defaultEncoding = zapcore.EncoderConfig{
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
	}
	defaultLogger    = zap.Must(defaultConfig(defaultEncoding).Build())
	userConfigDir, _ = os.UserConfigDir()
	fliptConfigFile  = filepath.Join(userConfigDir, "flipt", "config.yml")
)

func defaultConfig(encoding zapcore.EncoderConfig) zap.Config {
	return zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    encoding,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

func main() {
	var (
		rootCmd = &cobra.Command{
			Use:     "flipt",
			Short:   "Flipt is a modern feature flag solution",
			Version: version,
			Run: func(cmd *cobra.Command, _ []string) {
				logger, cfg := buildConfig()
				defer func() {
					_ = logger.Sync()
				}()

				if err := run(cmd.Context(), logger, cfg); err != nil {
					logger.Fatal("flipt", zap.Error(err))
				}
			},
			CompletionOptions: cobra.CompletionOptions{
				DisableDefaultCmd: true,
			},
		}

		migrateCmd = &cobra.Command{
			Use:   "migrate",
			Short: "Run pending database migrations",
			Run: func(cmd *cobra.Command, _ []string) {
				logger, cfg := buildConfig()
				defer func() {
					_ = logger.Sync()
				}()

				migrator, err := sql.NewMigrator(*cfg, logger)
				if err != nil {
					logger.Fatal("initializing migrator", zap.Error(err))
				}

				defer migrator.Close()

				if err := migrator.Up(true); err != nil {
					logger.Fatal("running migrator", zap.Error(err))
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
		defaultLogger.Fatal("executing template", zap.Error(err))
	}

	banner = buf.String()

	rootCmd.SetVersionTemplate(banner)
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "path to config file")
	rootCmd.Flags().BoolVar(&forceMigrate, "force-migrate", false, "force migrations before running")
	_ = rootCmd.Flags().MarkHidden("force-migrate")

	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(newExportCommand())
	rootCmd.AddCommand(newImportCommand())
	rootCmd.AddCommand(newValidateCommand())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-interrupt
		cancel()
	}()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		defaultLogger.Fatal("execute", zap.Error(err))
	}
}

// determinePath will figure out which path to use for Flipt configuration.
func determinePath(cfgPath string) string {
	if cfgPath != "" {
		return cfgPath
	}

	_, err := os.Stat(fliptConfigFile)
	if err == nil {
		return fliptConfigFile
	}

	if !errors.Is(err, fs.ErrNotExist) {
		defaultLogger.Warn("unexpected error checking configuration path", zap.String("config_path", fliptConfigFile), zap.Error(err))
	}

	return defaultCfgPath
}

func buildConfig() (*zap.Logger, *config.Config) {
	path := determinePath(cfgPath)

	// read in config
	res, err := config.Load(path)
	if err != nil {
		defaultLogger.Fatal("loading configuration", zap.Error(err), zap.String("config_path", path))
	}

	cfg := res.Config

	encoding := defaultEncoding
	encoding.TimeKey = cfg.Log.Keys.Time
	encoding.LevelKey = cfg.Log.Keys.Level
	encoding.MessageKey = cfg.Log.Keys.Message

	loggerConfig := defaultConfig(encoding)

	// log to file if enabled
	if cfg.Log.File != "" {
		loggerConfig.OutputPaths = []string{cfg.Log.File}
	}

	// parse/set log level
	loggerConfig.Level, err = zap.ParseAtomicLevel(cfg.Log.Level)
	if err != nil {
		defaultLogger.Fatal("parsing log level", zap.String("level", cfg.Log.Level), zap.Error(err))
	}

	if cfg.Log.Encoding > config.LogEncodingConsole {
		loggerConfig.Encoding = cfg.Log.Encoding.String()

		// don't encode with colors if not using console log output
		loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	logger := zap.Must(loggerConfig.Build())

	// print out any warnings from config parsing
	for _, warning := range res.Warnings {
		logger.Warn("configuration warning", zap.String("message", warning))
	}

	logger.Debug("configuration source", zap.String("path", path))

	return logger, cfg
}

func run(ctx context.Context, logger *zap.Logger, cfg *config.Config) error {
	isConsole := cfg.Log.Encoding == config.LogEncodingConsole

	if isConsole {
		color.Cyan("%s\n", banner)
	} else {
		logger.Info("flipt starting", zap.String("version", version), zap.String("commit", commit), zap.String("date", date), zap.String("go_version", goVersion))
	}

	var (
		isRelease   = release.Is(version)
		releaseInfo release.Info
		err         error
	)

	if cfg.Meta.CheckForUpdates && isRelease {
		logger.Debug("checking for updates")

		releaseInfo, err = release.Check(ctx, version)
		if err != nil {
			logger.Warn("checking for updates", zap.Error(err))
		}

		logger.Debug("version info", zap.String("current_version", releaseInfo.CurrentVersion), zap.String("latest_version", releaseInfo.LatestVersion))

		if isConsole {
			if releaseInfo.UpdateAvailable {
				color.Yellow("A newer version of Flipt exists at %s, \nplease consider updating to the latest version.", releaseInfo.LatestVersionURL)
			} else {
				color.Green("You are currently running the latest version of Flipt [%s]!", releaseInfo.CurrentVersion)
			}
		} else {
			if releaseInfo.UpdateAvailable {
				logger.Info("newer version available", zap.String("version", releaseInfo.LatestVersion), zap.String("url", releaseInfo.LatestVersionURL))
			} else {
				logger.Info("running latest version", zap.String("version", releaseInfo.CurrentVersion))
			}
		}
	}

	// see: https://consoledonottrack.com/
	if (os.Getenv("DO_NOT_TRACK") == "true" || os.Getenv("DO_NOT_TRACK") == "1") && cfg.Meta.TelemetryEnabled {
		logger.Debug("DO_NOT_TRACK environment variable set, disabling telemetry")
		cfg.Meta.TelemetryEnabled = false
	}

	if (os.Getenv("CI") == "true" || os.Getenv("CI") == "1") && cfg.Meta.TelemetryEnabled {
		logger.Debug("CI detected, disabling telemetry")
		cfg.Meta.TelemetryEnabled = false
	}

	if !isRelease && cfg.Meta.TelemetryEnabled {
		logger.Debug("not a release version, disabling telemetry")
		cfg.Meta.TelemetryEnabled = false
	}

	g, ctx := errgroup.WithContext(ctx)

	if err := initLocalState(cfg); err != nil {
		logger.Debug("disabling telemetry, state directory not accessible", zap.String("path", cfg.Meta.StateDirectory), zap.Error(err))
		cfg.Meta.TelemetryEnabled = false
	} else {
		logger.Debug("local state directory exists", zap.String("path", cfg.Meta.StateDirectory))
	}

	info := info.Flipt{
		Commit:           commit,
		BuildDate:        date,
		GoVersion:        goVersion,
		Version:          version,
		LatestVersion:    releaseInfo.LatestVersion,
		LatestVersionURL: releaseInfo.LatestVersionURL,
		IsRelease:        isRelease,
		UpdateAvailable:  releaseInfo.UpdateAvailable,
	}

	if cfg.Meta.TelemetryEnabled {
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

	grpcServer, err := cmd.NewGRPCServer(ctx, logger, cfg, info, forceMigrate)
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

	// block until root context is cancelled
	// and shutdown has been signalled
	<-ctx.Done()

	logger.Info("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	_ = httpServer.Shutdown(shutdownCtx)
	_ = grpcServer.Shutdown(shutdownCtx)

	return g.Wait()
}

// check if state directory already exists, create it if not
func initLocalState(cfg *config.Config) error {
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
