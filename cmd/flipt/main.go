package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"text/template"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/fatih/color"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/cmd"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/info"
	"go.flipt.io/flipt/internal/release"
	"go.flipt.io/flipt/internal/telemetry"
	"go.flipt.io/reverst/client"
	"go.flipt.io/reverst/pkg/protocol"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	providedConfigFile string
	forceMigrate       bool
	version            = "dev"
	commit             string
	date               string
	goVersion          = runtime.Version()
	goOS               = runtime.GOOS
	goArch             = runtime.GOARCH
	analyticsKey       string
	analyticsEndpoint  string
	banner             string
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

func main() {
	if err := exec(); err != nil {
		os.Exit(1)
	}
}

func exec() error {
	var (
		rootCmd = &cobra.Command{
			Use:   "flipt <command> <subcommand> [flags]",
			Short: "Flipt is a modern, self-hosted, feature flag solution",
			Example: heredoc.Doc(`
				$ flipt
				$ flipt config init
				$ flipt --config /path/to/config.yml migrate
			`),
			Version: version,
			RunE: func(cmd *cobra.Command, _ []string) error {
				ctx := cmd.Context()
				logger, cfg, err := buildConfig(ctx)
				if err != nil {
					return err
				}

				defer func() {
					_ = logger.Sync()
				}()

				return run(ctx, logger, cfg)
			},
			CompletionOptions: cobra.CompletionOptions{
				DisableDefaultCmd: true,
			},
			SilenceUsage: true,
		}

		t   = template.Must(template.New("banner").Parse(bannerTmpl))
		buf = new(bytes.Buffer)
	)

	if err := t.Execute(buf, &bannerOpts{
		Version:   version,
		Commit:    commit,
		Date:      date,
		GoVersion: goVersion,
		GoOS:      goOS,
		GoArch:    goArch,
	}); err != nil {
		return fmt.Errorf("executing template %w", err)
	}

	banner = buf.String()

	rootCmd.SetVersionTemplate(banner)
	rootCmd.Flags().StringVar(&providedConfigFile, "config", "", "path to config file")
	rootCmd.Flags().BoolVar(&forceMigrate, "force-migrate", false, "force migrations before running")
	_ = rootCmd.Flags().MarkHidden("force-migrate")

	rootCmd.AddCommand(newMigrateCommand())
	rootCmd.AddCommand(newExportCommand())
	rootCmd.AddCommand(newImportCommand())
	rootCmd.AddCommand(newValidateCommand())
	rootCmd.AddCommand(newConfigCommand())
	rootCmd.AddCommand(newCompletionCommand())
	rootCmd.AddCommand(newDocCommand())
	rootCmd.AddCommand(newBundleCommand())
	rootCmd.AddCommand(newEvaluateCommand())
	rootCmd.AddCommand(newCloudCommand())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-interrupt
		cancel()
	}()

	return rootCmd.ExecuteContext(ctx)
}

// determineConfig will figure out which file to use for Flipt configuration.
func determineConfig(configFile string) (string, bool) {
	// if config file is provided, use it
	if configFile != "" {
		return configFile, true
	}

	// otherwise, check if user config file exists on filesystem
	_, err := os.Stat(userConfigFile)
	if err == nil {
		return userConfigFile, true
	}

	if !errors.Is(err, fs.ErrNotExist) {
		defaultLogger.Warn("unexpected error checking configuration file", zap.String("config_file", userConfigFile), zap.Error(err))
	}

	// finally, check if default config file exists on filesystem
	_, err = os.Stat(defaultConfigFile)
	if err == nil {
		return defaultConfigFile, true
	}

	if !errors.Is(err, fs.ErrNotExist) {
		defaultLogger.Warn("unexpected error checking configuration file", zap.String("config_file", defaultConfigFile), zap.Error(err))
	}

	return "", false
}

func buildConfig(ctx context.Context) (*zap.Logger, *config.Config, error) {
	path, found := determineConfig(providedConfigFile)

	// read in config if it exists
	// otherwise, use defaults
	res, err := config.Load(ctx, path)
	if err != nil {
		return nil, nil, fmt.Errorf("loading configuration: %w", err)
	}

	if !found {
		defaultLogger.Info("no configuration file found, using defaults")
	}

	cfg := res.Config

	encoding := defaultEncoding
	encoding.TimeKey = cfg.Log.Keys.Time
	encoding.LevelKey = cfg.Log.Keys.Level
	encoding.MessageKey = cfg.Log.Keys.Message

	loggerConfig := loggerConfig(encoding)

	// log to file if enabled
	if cfg.Log.File != "" {
		loggerConfig.OutputPaths = []string{cfg.Log.File}
	}

	// parse/set log level
	loggerConfig.Level, err = zap.ParseAtomicLevel(cfg.Log.Level)
	if err != nil {
		defaultLogger.Warn("parsing log level, defaulting to INFO", zap.String("level", cfg.Log.Level), zap.Error(err))
		// default to info level
		loggerConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
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

	if found {
		logger.Debug("configuration source", zap.String("path", path))
	}

	return logger, cfg, nil
}

const (
	dntVar = "DO_NOT_TRACK"
	ciVar  = "CI"
)

func isSet(env string) bool {
	return os.Getenv(env) == "true" || os.Getenv(env) == "1"
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

	info := info.Flipt{
		Commit:           commit,
		BuildDate:        date,
		GoVersion:        goVersion,
		Version:          version,
		LatestVersion:    releaseInfo.LatestVersion,
		LatestVersionURL: releaseInfo.LatestVersionURL,
		IsRelease:        isRelease,
		UpdateAvailable:  releaseInfo.UpdateAvailable,
		OS:               goOS,
		Arch:             goArch,
	}

	if cfg.Meta.TelemetryEnabled {
		logger := logger.With(zap.String("component", "telemetry"))

		g.Go(func() error {
			reporter, err := telemetry.NewReporter(*cfg, logger, analyticsKey, analyticsEndpoint, info)
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

	if cfg.Experimental.Cloud.Enabled && cfg.Server.Cloud.Enabled {
		// starts QUIC tunnel server to connect to Cloud

		g.Go(func() error {
			var (
				orgHost       string
				tunnel        string
				authenticator client.Authenticator
			)

			// prefer API key over local token
			if cfg.Cloud.Authentication.ApiKey != "" {
				authenticator = client.BearerAuthenticator(cfg.Cloud.Authentication.ApiKey)

				if cfg.Cloud.Organization == "" || cfg.Cloud.Gateway == "" {
					return errors.New("missing cloud.organization or cloud.gateway")
				}

				orgHost = fmt.Sprintf("%s.%s", cfg.Cloud.Organization, cfg.Cloud.Host)
				tunnel = fmt.Sprintf("%s-%s", cfg.Cloud.Gateway, orgHost)
			} else {
				// read in local token from file
				cloudAuthFile := filepath.Join(userConfigDir, "cloud.json")
				cloudAuthBytes, err := os.ReadFile(cloudAuthFile)
				if err != nil {
					return fmt.Errorf("reading cloud auth token: %w", err)
				}

				var auth cloudAuth

				if err := json.Unmarshal(cloudAuthBytes, &auth); err != nil {
					return fmt.Errorf("unmarshalling cloud auth token: %w", err)
				}

				authenticator = client.BearerAuthenticator(auth.Token, client.WithScheme("JWT"))

				// use gateway and organization from local token
				if auth.Tunnel == nil || auth.Tunnel.Organization == "" || auth.Tunnel.Gateway == "" {
					return errors.New("missing cloud.organization or cloud.gateway")
				}

				orgHost = fmt.Sprintf("%s.%s", auth.Tunnel.Organization, cfg.Cloud.Host)
				tunnel = fmt.Sprintf("%s-%s", auth.Tunnel.Gateway, orgHost)
			}

			sl := slog.New(zapslog.NewHandler(logger.Core(), nil))

			tunnelServer := &client.Server{
				TunnelGroup:   tunnel,
				Handler:       httpServer.Handler,
				Logger:        sl,
				Authenticator: authenticator,
				TLSConfig: &tls.Config{
					MinVersion: tls.VersionTLS13,
					NextProtos: []string{protocol.Name},
					ServerName: orgHost,
				},
			}

			addr := fmt.Sprintf("%s:%d", tunnel, cfg.Server.Cloud.Port)

			logger.Info("cloud tunnel established", zap.String("address", fmt.Sprintf("https://%s", tunnel)))

			if err := tunnelServer.DialAndServe(ctx, addr); err != nil &&
				!errors.Is(err, http.ErrServerClosed) &&
				!errors.Is(err, context.Canceled) {
				return fmt.Errorf("creating cloud tunnel server: %w", err)
			}

			return nil
		})
	}

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
