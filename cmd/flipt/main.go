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

	"github.com/MakeNowJust/heredoc"
	"github.com/fatih/color"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.flipt.io/flipt/internal/cmd"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/info"
	"go.flipt.io/flipt/internal/release"
	"go.flipt.io/flipt/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

// Global variables for configs and logging
var (
	version           = "dev" // These values can be set at build time
	commit            = ""    // via -ldflags "-X main.version=1.0.0 -X main.commit=abcdef"
	date              = ""
	goVersion         = runtime.Version()
	goOS              = runtime.GOOS
	goArch            = runtime.GOARCH
	analyticsKey      = ""
	analyticsEndpoint = ""

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
	defaultLogger    = zap.Must(loggerConfig(defaultEncoding).Build())
	userConfigDir, _ = config.Dir()
	userConfigFile   = filepath.Join(userConfigDir, "config.yml")
)

func loggerConfig(encoding zapcore.EncoderConfig) zap.Config {
	level, err := zap.ParseAtomicLevel(os.Getenv(config.EnvPrefix + "_LOG_LEVEL"))
	if err != nil {
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	return zap.Config{
		Level:            level,
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    encoding,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

// rootCommand holds all configuration for the root command
type rootCommand struct {
	configFile   string
	forceMigrate bool

	// Banner template
	banner string

	configManager *configManager
}

func main() {
	if err := exec(); err != nil {
		os.Exit(1)
	}
}

func exec() error {
	var (
		root = &rootCommand{}
		cmd  = root.newCommand()
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-interrupt
		cancel()
	}()

	// Parse the flags first to get the config file
	if err := cmd.ParseFlags(os.Args[1:]); err != nil {
		// Ignore flag parsing errors, they will be handled by cobra during ExecuteContext
		if !errors.Is(err, pflag.ErrHelp) {
			defaultLogger.Debug("parsing flags", zap.Error(err))
		}
	}

	// Update the config file from flags
	if cfgFlag, err := cmd.Flags().GetString("config"); err == nil && cfgFlag != "" {
		root.configFile = cfgFlag
	}

	// Make sure all subcommands also use the updated config manager
	cmd = root.newCommand()

	return cmd.ExecuteContext(ctx)
}

// newCommand creates a new root cobra command
func (r *rootCommand) newCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flipt <command> <subcommand> [flags]",
		Short: "Flipt is a modern, self-hosted, feature flag solution",
		Example: heredoc.Doc(`
			$ flipt
			$ flipt config init
			$ flipt --config /path/to/config.yml migrate
		`),
		Version: version,
		RunE:    r.run,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		SilenceUsage: true,
	}

	// Generate banner from template
	t := template.Must(template.New("banner").Parse(bannerTmpl))
	buf := new(bytes.Buffer)

	if err := t.Execute(buf, &bannerOpts{
		Version:   version,
		Commit:    commit,
		Date:      date,
		GoVersion: goVersion,
		GoOS:      goOS,
		GoArch:    goArch,
	}); err != nil {
		defaultLogger.Error("executing template", zap.Error(err))
	}

	r.banner = buf.String()
	cmd.SetVersionTemplate(r.banner)

	// Set up flags
	cmd.Flags().StringVar(&r.configFile, "config", "", "path to config file")
	cmd.PersistentFlags().StringVar(&r.configFile, "config", "", "path to config file")

	cmd.Flags().BoolVar(&r.forceMigrate, "force-migrate", false, "force migrations before running")
	_ = cmd.Flags().MarkHidden("force-migrate")

	// Initialize config manager with empty config file for now
	// It will be updated in exec() after flag parsing
	configManager := newConfigManager(r.configFile)
	r.configManager = configManager

	// Add subcommands with root command reference
	cmd.AddCommand(newMigrateCommand(configManager))
	cmd.AddCommand(newExportCommand(configManager))
	cmd.AddCommand(newImportCommand(configManager))
	cmd.AddCommand(newValidateCommand(configManager))
	cmd.AddCommand(newConfigCommand(configManager))
	cmd.AddCommand(newCompletionCommand())
	cmd.AddCommand(newDocCommand())
	cmd.AddCommand(newBundleCommand(configManager))
	cmd.AddCommand(newEvaluateCommand())

	return cmd
}

// run implements the main command logic
func (r *rootCommand) run(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	// Get the config file from the command line flag
	if cfgFlag, err := cmd.Flags().GetString("config"); err == nil && cfgFlag != "" {
		// Update the config manager's config file if needed
		r.configManager.configFile = cfgFlag
	}

	logger, cfg, err := r.configManager.build(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = logger.Sync()
	}()

	return r.runServer(ctx, logger, cfg)
}

const (
	dntVar = "DO_NOT_TRACK"
	ciVar  = "CI"
)

func isSet(env string) bool {
	return os.Getenv(env) == "true" || os.Getenv(env) == "1"
}

func (r *rootCommand) runServer(ctx context.Context, logger *zap.Logger, cfg *config.Config) error {
	isConsole := cfg.Log.Encoding == config.LogEncodingConsole

	if isConsole {
		color.Cyan("%s\n", r.banner)
	} else {
		logger.Info("flipt starting",
			zap.String("version", version),
			zap.String("commit", commit),
			zap.String("date", date),
			zap.String("go_version", goVersion))
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
	if isSet(dntVar) && cfg.Meta.TelemetryEnabled {
		logger.Debug("DO_NOT_TRACK environment variable set, disabling telemetry")
		cfg.Meta.TelemetryEnabled = false
	}

	if isSet(ciVar) && cfg.Meta.TelemetryEnabled {
		logger.Debug("CI detected, disabling telemetry")
		cfg.Meta.TelemetryEnabled = false
	}

	if !isRelease && cfg.Meta.TelemetryEnabled {
		logger.Debug("not a release version, disabling telemetry")
		cfg.Meta.TelemetryEnabled = false
	}

	g, ctx := errgroup.WithContext(ctx)

	if err := initMetaStateDir(cfg); err != nil {
		logger.Debug("disabling telemetry, state directory not accessible", zap.String("path", cfg.Meta.StateDirectory), zap.Error(err))
		cfg.Meta.TelemetryEnabled = false
	} else {
		logger.Debug("local state directory exists", zap.String("path", cfg.Meta.StateDirectory))
	}

	serverInfo := info.New(
		info.WithBuild(commit, date, goVersion, version, isRelease),
		info.WithLatestRelease(releaseInfo),
		info.WithOS(goOS, goArch),
		info.WithConfig(cfg),
	)

	if cfg.Meta.TelemetryEnabled {
		logger := logger.With(zap.String("component", "telemetry"))

		g.Go(func() error {
			reporter, err := telemetry.NewReporter(*cfg, logger, analyticsKey, analyticsEndpoint, serverInfo)
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

	// in-process client connection for grpc services
	var ipch = &inprocgrpc.Channel{}
	ipch = ipch.WithServerUnaryInterceptor(grpc_middleware.ChainUnaryServer(
		//nolint:staticcheck // Deprecated but inprocgrpc does not support stats handlers
		otelgrpc.UnaryServerInterceptor(),
	))

	// initialize grpc server
	grpcServer, err := cmd.NewGRPCServer(ctx, logger, cfg, ipch, serverInfo, r.forceMigrate)
	if err != nil {
		return err
	}

	// starts grpc server
	g.Go(grpcServer.Run)

	httpServer, err := cmd.NewHTTPServer(ctx, logger, cfg, ipch, serverInfo)
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

func ensureDir(path string) error {
	fp, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// directory doesnt exist, so try to create it
			return os.MkdirAll(path, 0700)
		}
		return fmt.Errorf("checking directory: %w", err)
	}

	if fp != nil && !fp.IsDir() {
		return fmt.Errorf("not a directory")
	}

	return nil
}

// check if state directory already exists, create it if not
func initMetaStateDir(cfg *config.Config) error {
	if cfg.Meta.StateDirectory == "" {
		var err error
		cfg.Meta.StateDirectory, err = config.Dir()
		if err != nil {
			return err
		}
	}

	return ensureDir(cfg.Meta.StateDirectory)
}
