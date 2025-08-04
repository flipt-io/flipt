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
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/fatih/color"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/cmd"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/coss/license"
	_ "go.flipt.io/flipt/internal/coss/secrets/vault" // Register vault provider for Pro features
	"go.flipt.io/flipt/internal/info"
	"go.flipt.io/flipt/internal/otel"
	"go.flipt.io/flipt/internal/otel/logs"
	"go.flipt.io/flipt/internal/product"
	"go.flipt.io/flipt/internal/release"
	"go.flipt.io/flipt/internal/secrets"
	"go.flipt.io/flipt/internal/telemetry"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

var (
	providedConfigFile string
	forceMigrate       bool
	version            = "v2.0.0-dev"
	commit             string
	date               string
	goVersion          = runtime.Version()
	goOS               = runtime.GOOS
	goArch             = runtime.GOARCH
	analyticsKey       string
	analyticsEndpoint  string
	banner             string
	keygenVerifyKey    string
	keygenAccountID    string
	keygenProductID    string
	// Secret reference pattern for config resolution
	secretReference = regexp.MustCompile(`^\${secret:([a-zA-Z0-9_:/-]+)}$`)
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
	if err := execute(); err != nil {
		os.Exit(1)
	}
}

func execute() error {
	rootCmd := &cobra.Command{
		Use:   "flipt <command> <subcommand> [flags]",
		Short: "Flipt is a cloud-native, self-hosted, feature flag solution that manages feature flags in your Git repositories",
		Example: heredoc.Doc(`
			$ flipt server
			$ flipt config init
			$ flipt --config /path/to/config.yml migrate
		`),
		Version: version,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		SilenceUsage: true,
	}

	var (
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

	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Run the Flipt server",
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
	}

	serverCmd.Flags().StringVar(&providedConfigFile, "config", "", "path to config file")
	serverCmd.Flags().BoolVar(&forceMigrate, "force-migrate", false, "force migrations before running")
	_ = serverCmd.Flags().MarkHidden("force-migrate")

	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(newMigrateCommand())
	rootCmd.AddCommand(newValidateCommand())
	rootCmd.AddCommand(newConfigCommand())
	rootCmd.AddCommand(newCompletionCommand())
	rootCmd.AddCommand(newDocCommand())
	rootCmd.AddCommand(newQuickstartCommand())
	rootCmd.AddCommand(newEvaluateCommand())

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

	// Load config normally
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
	var err error

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if os.Getenv("OTEL_LOGS_EXPORTER") != "" {
		otelResource, err := otel.NewResource(ctx, version)
		if err != nil {
			return fmt.Errorf("creating otel resource: %w", err)
		}

		logsExp, logsExpShutdown, err := logs.GetExporter(ctx)
		if err != nil {
			return fmt.Errorf("creating otel log exporter: %w", err)
		}

		defer func() {
			_ = logsExpShutdown(shutdownCtx)
		}()

		logProcessor := log.NewBatchProcessor(logsExp)
		defer func() {
			_ = logProcessor.Shutdown(shutdownCtx)
		}()

		loggerProvider := log.NewLoggerProvider(
			log.WithResource(otelResource),
			log.WithProcessor(logProcessor),
		)

		defer func() {
			_ = loggerProvider.Shutdown(shutdownCtx)
		}()

		global.SetLoggerProvider(loggerProvider)
		lcore := zapcore.NewTee(logger.Core(), otelzap.NewCore("flipt", otelzap.WithLoggerProvider(loggerProvider)))
		logger = zap.New(lcore)
	}

	isConsole := cfg.Log.Encoding == config.LogEncodingConsole

	if isConsole {
		color.Cyan("%s\n", banner)
	} else {
		logger.Info("flipt starting", zap.String("version", version), zap.String("commit", commit), zap.String("date", date), zap.String("go_version", goVersion))
	}

	var (
		isRelease   = release.Is(version)
		releaseInfo release.Info
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

	// initialize secrets manager and resolve secrets in config before creating any components
	secretsManager, err := secrets.NewManager(logger, cfg)
	if err != nil {
		return fmt.Errorf("initializing secrets manager: %w", err)
	}

	defer func() {
		if secretsManager != nil {
			_ = secretsManager.Close()
		}
	}()

	// resolve secrets in the config if any secret references exist
	if err := resolveSecretsInConfig(ctx, cfg, secretsManager); err != nil {
		return fmt.Errorf("resolving secrets in config: %w", err)
	}

	logger.Debug("secrets manager initialized and config processed")

	licenseManagerOpts := []license.LicenseManagerOption{}

	// Enable pro features in development mode
	if !isRelease {
		licenseManagerOpts = append(licenseManagerOpts, license.WithProduct(product.Pro))
	}

	if keygenVerifyKey != "" {
		licenseManagerOpts = append(licenseManagerOpts, license.WithVerificationKey(keygenVerifyKey))
	}

	licenseManager, licenseManagerShutdown := license.NewManager(ctx, logger, keygenAccountID, keygenProductID, &cfg.License, licenseManagerOpts...)

	defer func() {
		// Use a dedicated timeout context for deactivation to avoid competing with other shutdown operations
		deactivateCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		_ = licenseManagerShutdown(deactivateCtx)
		cancel()
	}()

	// Validate license requirements for secrets providers
	if cfg.Secrets.Providers.Vault != nil && cfg.Secrets.Providers.Vault.Enabled {
		if licenseManager.Product() == product.OSS {
			return fmt.Errorf("vault secrets provider requires a paid license")
		}
	}

	info := info.New(
		info.WithBuild(commit, date, goVersion, version, isRelease),
		info.WithLatestRelease(releaseInfo),
		info.WithConfig(cfg),
		info.WithLicenseManager(licenseManager),
	)

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

	// in-process client connection for grpc services
	ipch := &inprocgrpc.Channel{}
	ipch = ipch.WithServerUnaryInterceptor(grpc_middleware.ChainUnaryServer(
		//nolint:staticcheck // Deprecated but inprocgrpc does not support stats handlers
		otelgrpc.UnaryServerInterceptor(),
	))

	var grpcOptions []cmd.GRPCServerOption
	if forceMigrate {
		grpcOptions = append(grpcOptions, cmd.WithForceMigrate())
	}

	grpcServer, err := cmd.NewGRPCServer(ctx, logger, cfg, ipch, info, licenseManager, secretsManager, grpcOptions...)
	if err != nil {
		return err
	}

	// starts grpc server
	g.Go(grpcServer.Run)

	httpServer, err := cmd.NewHTTPServer(ctx, logger, cfg, ipch, info)
	if err != nil {
		return err
	}

	// starts REST http(s) server
	g.Go(httpServer.Run)

	// block until root context is cancelled
	// and shutdown has been signalled
	<-ctx.Done()

	logger.Info("shutting down...")

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

// resolveSecretsInConfig walks the config structure and resolves any secret references in-place
func resolveSecretsInConfig(ctx context.Context, cfg *config.Config, secretsManager secrets.Manager) error {
	return walkConfigForSecrets(ctx, reflect.ValueOf(cfg).Elem(), secretsManager)
}

// walkConfigForSecrets recursively walks the config struct and resolves secret references
func walkConfigForSecrets(ctx context.Context, v reflect.Value, secretsManager secrets.Manager) error {
	switch v.Kind() {
	case reflect.String:
		if v.CanSet() {
			str := v.String()
			// Check if this is a secret reference using the regex
			if secretReference.MatchString(str) {
				// Extract the reference: ${secret:reference} -> reference
				reference := secretReference.ReplaceAllString(str, `$1`)

				// Parse the reference format: either "key" or "provider:key"
				parts := strings.Split(reference, ":")

				var secretRef secrets.Reference
				switch {
				case len(parts) == 1:
					// Simple format: "key-name" - use default provider
					secretRef = secrets.Reference{
						Provider: "", // Use default provider
						Path:     parts[0],
						Key:      parts[0],
					}
				case len(parts) == 2:
					// Simple format: "provider:key"
					secretRef = secrets.Reference{
						Provider: parts[0],
						Path:     parts[1],
						Key:      parts[1],
					}
				default:
					return fmt.Errorf("invalid secret reference format %q, expected 'key' or 'provider:key'", reference)
				}

				// Resolve the secret
				secretValue, err := secretsManager.GetSecretValue(ctx, secretRef)
				if err != nil {
					return fmt.Errorf("resolving secret reference %q: %w", str, err)
				}

				// Replace the secret reference with the resolved value
				v.SetString(string(secretValue))
			}
		}

	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if err := walkConfigForSecrets(ctx, v.Field(i), secretsManager); err != nil {
				return err
			}
		}

	case reflect.Ptr:
		if !v.IsNil() {
			if err := walkConfigForSecrets(ctx, v.Elem(), secretsManager); err != nil {
				return err
			}
		}

	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if err := walkConfigForSecrets(ctx, v.Index(i), secretsManager); err != nil {
				return err
			}
		}

	case reflect.Map:
		for _, key := range v.MapKeys() {
			value := v.MapIndex(key)
			if value.Kind() == reflect.Interface && !value.IsNil() {
				value = value.Elem()
			}
			if err := walkConfigForSecrets(ctx, value, secretsManager); err != nil {
				return err
			}
		}
	}

	return nil
}
