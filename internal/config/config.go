package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"go.flipt.io/flipt/internal/storage/fs/object"
	"gocloud.dev/blob"
	"golang.org/x/exp/constraints"
)

const (
	Version   = "1.0"
	EnvPrefix = "FLIPT"
)

var (
	_        validator = (*Config)(nil)
	envsubst           = regexp.MustCompile(`^\${([a-zA-Z_]+[a-zA-Z0-9_]*)}$`)
)

var DecodeHooks = []mapstructure.DecodeHookFunc{
	stringToEnvsubstHookFunc(),
	mapstructure.StringToTimeDurationHookFunc(),
	stringToSliceHookFunc(),
	stringToEnumHookFunc(stringToCacheBackend),
	stringToEnumHookFunc(stringToTracingExporter),
	stringToEnumHookFunc(stringToScheme),
	stringToEnumHookFunc(stringToDatabaseProtocol),
	stringToEnumHookFunc(stringToAuthMethod),
}

// Config contains all of Flipts configuration needs.
//
// The root of this structure contains a collection of sub-configuration categories.
//
// Each sub-configuration (e.g. LogConfig) optionally implements either or both of
// the defaulter or validator interfaces.
// Given the sub-config implements a `setDefaults(*viper.Viper) []string` method
// then this will be called with the viper context before unmarshalling.
// This allows the sub-configuration to set any appropriate defaults.
// Given the sub-config implements a `validate() error` method
// then this will be called after unmarshalling, such that the function can emit
// any errors derived from the resulting state of the configuration.
type Config struct {
	Version        string               `json:"version,omitempty" mapstructure:"version,omitempty" yaml:"version,omitempty"`
	Audit          AuditConfig          `json:"audit,omitempty" mapstructure:"audit" yaml:"audit,omitempty"`
	Authentication AuthenticationConfig `json:"authentication,omitempty" mapstructure:"authentication" yaml:"authentication,omitempty"`
	Authorization  AuthorizationConfig  `json:"authorization,omitempty" mapstructure:"authorization" yaml:"authorization,omitempty"`
	Cache          CacheConfig          `json:"cache,omitempty" mapstructure:"cache" yaml:"cache,omitempty"`
	Cloud          CloudConfig          `json:"cloud,omitempty" mapstructure:"cloud" yaml:"cloud,omitempty" experiment:"cloud"`
	Cors           CorsConfig           `json:"cors,omitempty" mapstructure:"cors" yaml:"cors,omitempty"`
	Database       DatabaseConfig       `json:"db,omitempty" mapstructure:"db" yaml:"db,omitempty"`
	Diagnostics    DiagnosticConfig     `json:"diagnostics,omitempty" mapstructure:"diagnostics" yaml:"diagnostics,omitempty"`
	Experimental   ExperimentalConfig   `json:"experimental,omitempty" mapstructure:"experimental" yaml:"experimental,omitempty"`
	Log            LogConfig            `json:"log,omitempty" mapstructure:"log" yaml:"log,omitempty"`
	Meta           MetaConfig           `json:"meta,omitempty" mapstructure:"meta" yaml:"meta,omitempty"`
	Analytics      AnalyticsConfig      `json:"analytics,omitempty" mapstructure:"analytics" yaml:"analytics,omitempty"`
	Server         ServerConfig         `json:"server,omitempty" mapstructure:"server" yaml:"server,omitempty"`
	Storage        StorageConfig        `json:"storage,omitempty" mapstructure:"storage" yaml:"storage,omitempty"`
	Metrics        MetricsConfig        `json:"metrics,omitempty" mapstructure:"metrics" yaml:"metrics,omitempty"`
	Tracing        TracingConfig        `json:"tracing,omitempty" mapstructure:"tracing" yaml:"tracing,omitempty"`
	UI             UIConfig             `json:"ui,omitempty" mapstructure:"ui" yaml:"ui,omitempty"`
}

type Result struct {
	Config   *Config
	Warnings []string
}

// Dir returns the default root directory for Flipt configuration
func Dir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("getting user config dir: %w", err)
	}

	return filepath.Join(configDir, "flipt"), nil
}

func Load(ctx context.Context, path string) (*Result, error) {
	v := viper.New()
	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg *Config

	if path == "" {
		cfg = Default()
	} else {
		cfg = &Config{}
		file, err := getConfigFile(ctx, path)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		stat, err := file.Stat()
		if err != nil {
			return nil, err
		}

		// reimplement logic from v.ReadInConfig()
		v.SetConfigFile(stat.Name())
		ext := filepath.Ext(stat.Name())
		if len(ext) > 1 {
			ext = ext[1:]
		}
		if !slices.Contains(viper.SupportedExts, ext) {
			return nil, viper.UnsupportedConfigError(ext)
		}
		if err := v.ReadConfig(file); err != nil {
			return nil, err
		}
	}

	var (
		result      = &Result{Config: cfg}
		deprecators []deprecator
		defaulters  []defaulter
		validators  []validator
	)

	f := func(field any) {
		// for-each deprecator implementing field we collect
		// them up and return them to be run before unmarshalling and before setting defaults.
		if deprecator, ok := field.(deprecator); ok {
			deprecators = append(deprecators, deprecator)
		}

		// for-each defaulter implementing fields we invoke
		// setting any defaults during this prepare stage
		// on the supplied viper.
		if defaulter, ok := field.(defaulter); ok {
			defaulters = append(defaulters, defaulter)
		}

		// for-each validator implementing field we collect
		// them up and return them to be validated after
		// unmarshalling.
		if validator, ok := field.(validator); ok {
			validators = append(validators, validator)
		}
	}

	// invoke the field visitor on the root config firsts
	root := reflect.ValueOf(cfg).Interface()
	f(root)

	// these are reflected config top-level types for fields where
	// they have been marked as experimental and their associated
	// flag has enabled set to false.
	var skippedTypes []reflect.Type

	val := reflect.ValueOf(cfg).Elem()
	for i := 0; i < val.NumField(); i++ {
		// search for all expected env vars since Viper cannot
		// infer when doing Unmarshal + AutomaticEnv.
		// see: https://github.com/spf13/viper/issues/761
		structField := val.Type().Field(i)
		if exp := structField.Tag.Get("experiment"); exp != "" {
			// TODO(georgemac): register target for skipping
			if !v.GetBool(fmt.Sprintf("experimental.%s.enabled", exp)) {
				skippedTypes = append(skippedTypes, structField.Type)
			}
		}

		key := fieldKey(structField)
		bindEnvVars(v, getFliptEnvs(), []string{key}, structField.Type)

		field := val.Field(i).Addr().Interface()
		f(field)
	}

	// run any deprecations checks
	for _, deprecator := range deprecators {
		warnings := deprecator.deprecations(v)
		for _, warning := range warnings {
			result.Warnings = append(result.Warnings, warning.Message())
		}
	}

	// run any defaulters
	for _, defaulter := range defaulters {
		if err := defaulter.setDefaults(v); err != nil {
			return nil, err
		}
	}

	if err := v.Unmarshal(cfg, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			append(DecodeHooks, experimentalFieldSkipHookFunc(skippedTypes...))...,
		),
	)); err != nil {
		return nil, err
	}

	// run any validation steps
	for _, validator := range validators {
		if err := validator.validate(); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// getConfigFile provides a file from different types of storage.
func getConfigFile(ctx context.Context, path string) (fs.File, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	if slices.Contains(object.SupportedSchemes(), u.Scheme) {
		key := strings.TrimPrefix(u.Path, "/")
		u.Path = ""
		bucket, err := object.OpenBucket(ctx, u)
		if err != nil {
			return nil, err
		}
		defer bucket.Close()
		bucket.SetIOFSCallback(func() (context.Context, *blob.ReaderOptions) { return ctx, nil })
		return bucket.Open(key)
	}

	// assumes that the local file is used
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}

type defaulter interface {
	setDefaults(v *viper.Viper) error
}

type validator interface {
	validate() error
}

type deprecator interface {
	deprecations(v *viper.Viper) []deprecated
}

// fieldKey returns the name to be used when deriving a fields env var key.
// If marked as squash the key will be the empty string.
// Otherwise, it is derived from the lowercase name of the field.
func fieldKey(field reflect.StructField) string {
	if tag := field.Tag.Get("mapstructure"); tag != "" {
		tag, attr, ok := strings.Cut(tag, ",")
		if !ok || attr == "squash" || attr == "omitempty" {
			return tag
		}
	}

	return strings.ToLower(field.Name)
}

type envBinder interface {
	MustBindEnv(...string)
}

// bindEnvVars descends into the provided struct field binding any expected
// environment variable keys it finds reflecting struct and field tags.
func bindEnvVars(v envBinder, env, prefixes []string, typ reflect.Type) {
	// descend through pointers
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	switch typ.Kind() {
	case reflect.Map:
		// recurse into bindEnvVars while signifying that the last
		// key was unbound using the wildcard "*".
		bindEnvVars(v, env, append(prefixes, wildcard), typ.Elem())

		return
	case reflect.Struct:
		for i := 0; i < typ.NumField(); i++ {
			var (
				structField = typ.Field(i)
				key         = fieldKey(structField)
			)

			bind(env, prefixes, key, func(prefixes []string) {
				bindEnvVars(v, env, prefixes, structField.Type)
			})
		}

		return
	}

	bind(env, prefixes, "", func(prefixes []string) {
		v.MustBindEnv(strings.Join(prefixes, "."))
	})
}

const wildcard = "*"

func appendIfNotEmpty(s []string, v ...string) []string {
	for _, vs := range v {
		if vs != "" {
			s = append(s, vs)
		}
	}

	return s
}

// bind invokes the supplied function "fn" with each possible set of
// prefixes for the next prefix ("next").
// If the last prefix is "*" then we must search the current environment
// for matching env vars to obtain the potential keys which populate
// the unbound map keys.
func bind(env, prefixes []string, next string, fn func([]string)) {
	// given the previous entry is non-existent or not the wildcard
	if len(prefixes) < 1 || prefixes[len(prefixes)-1] != wildcard {
		fn(appendIfNotEmpty(prefixes, next))
		return
	}

	// drop the wildcard and derive all the possible keys from
	// existing environment variables.
	p := make([]string, len(prefixes)-1)
	copy(p, prefixes[:len(prefixes)-1])

	var (
		// makezero linter doesn't take note of subsequent copy
		// nolint https://github.com/ashanbrown/makezero/issues/12
		prefix = strings.ToUpper(strings.Join(append(p, ""), "_"))
		keys   = strippedKeys(env, prefix, strings.ToUpper(next))
	)

	for _, key := range keys {
		fn(appendIfNotEmpty(p, strings.ToLower(key), next))
	}
}

// strippedKeys returns a set of keys derived from a list of env var keys.
// It starts by filtering and stripping each key with a matching prefix.
// Given a child delimiter string is supplied it also trims the delimeter string
// and any remaining characters after this suffix.
//
// e.g strippedKeys(["A_B_C_D", "A_B_F_D", "A_B_E_D_G"], "A_B", "D")
// returns ["c", "f", "e"]
//
// It's purpose is to extract the parts of env vars which are likely
// keys in an arbitrary map type.
func strippedKeys(envs []string, prefix, delim string) (keys []string) {
	for _, env := range envs {
		if strings.HasPrefix(env, prefix) {
			env = env[len(prefix):]
			if env == "" {
				continue
			}

			if delim == "" {
				keys = append(keys, env)
				continue
			}

			// cut the string on the child key and take the left hand component
			if left, _, ok := strings.Cut(env, "_"+delim); ok {
				keys = append(keys, left)
			}
		}
	}
	return
}

// getFliptEnvs returns all environment variables which have FLIPT_
// as a prefix. It also strips this prefix before appending them to the
// resulting set.
func getFliptEnvs() (envs []string) {
	const prefix = "FLIPT_"
	for _, e := range os.Environ() {
		key, _, ok := strings.Cut(e, "=")
		if ok && strings.HasPrefix(key, prefix) {
			// strip FLIPT_ off env vars for convenience
			envs = append(envs, key[len(prefix):])
		}
	}
	return envs
}

func (c *Config) validate() (err error) {
	if c.Version != "" {
		if strings.TrimSpace(c.Version) != Version {
			return fmt.Errorf("invalid version: %s", c.Version)
		}
	}

	if c.Authorization.Required && !c.Authentication.Required {
		return fmt.Errorf("authorization requires authentication also be required")
	}

	return nil
}

func (c *Config) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		out []byte
		err error
	)

	if r.Header.Get("Accept") == "application/json+pretty" {
		out, err = json.MarshalIndent(c, "", "  ")
	} else {
		out, err = json.Marshal(c)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(out); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// stringToEnumHookFunc returns a DecodeHookFunc that converts strings to a target enum
func stringToEnumHookFunc[T constraints.Integer](mappings map[string]T) mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(T(0)) {
			return data, nil
		}

		enum := mappings[data.(string)]

		return enum, nil
	}
}

func experimentalFieldSkipHookFunc(types ...reflect.Type) mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if len(types) == 0 {
			return data, nil
		}

		if t.Kind() != reflect.Struct {
			return data, nil
		}

		// skip any types that match a type in the provided set
		for _, typ := range types {
			if t == typ {
				return reflect.New(typ).Interface(), nil
			}
		}

		return data, nil
	}
}

// stringToEnvsubstHookFunc returns a DecodeHookFunc that substitutes
// `${VARIABLE}` strings with their matching environment variables.
func stringToEnvsubstHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || f != reflect.TypeOf("") {
			return data, nil
		}
		str := data.(string)
		if !envsubst.MatchString(str) {
			return data, nil
		}
		key := envsubst.ReplaceAllString(str, `$1`)
		return os.Getenv(key), nil
	}
}

// stringToSliceHookFunc returns a DecodeHookFunc that converts
// string to []string by splitting using strings.Fields().
func stringToSliceHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{}) (interface{}, error) {
		if f != reflect.String || t != reflect.Slice {
			return data, nil
		}

		raw := data.(string)
		if raw == "" {
			return []string{}, nil
		}

		return strings.Fields(raw), nil
	}
}

// Default is the base config used when no configuration is explicit provided.
func Default() *Config {
	dbRoot, err := defaultDatabaseRoot()
	if err != nil {
		panic(err)
	}

	dbPath := filepath.ToSlash(filepath.Join(dbRoot, "flipt.db"))

	return &Config{
		Log: LogConfig{
			Level:     "INFO",
			Encoding:  LogEncodingConsole,
			GRPCLevel: "ERROR",
			Keys: LogKeys{
				Time:    "T",
				Level:   "L",
				Message: "M",
			},
		},

		UI: UIConfig{
			DefaultTheme: SystemUITheme,
		},

		Cors: CorsConfig{
			Enabled:        false,
			AllowedOrigins: []string{"*"},
			AllowedHeaders: []string{
				"Accept",
				"Authorization",
				"Content-Type",
				"X-CSRF-Token",
				"X-Fern-Language",
				"X-Fern-SDK-Name",
				"X-Fern-SDK-Version",
				"X-Flipt-Namespace",
				"X-Flipt-Accept-Server-Version",
			},
		},

		Cache: CacheConfig{
			Enabled: false,
			Backend: CacheMemory,
			TTL:     1 * time.Minute,
			Memory: MemoryCacheConfig{
				EvictionInterval: 5 * time.Minute,
			},
			Redis: RedisCacheConfig{
				Host:            "localhost",
				Port:            6379,
				RequireTLS:      false,
				Password:        "",
				DB:              0,
				PoolSize:        0,
				MinIdleConn:     0,
				ConnMaxIdleTime: 0,
				NetTimeout:      0,
			},
		},

		Diagnostics: DiagnosticConfig{
			Profiling: ProfilingDiagnosticConfig{
				Enabled: true,
			},
		},

		Server: ServerConfig{
			Host:      "0.0.0.0",
			Protocol:  HTTP,
			HTTPPort:  8080,
			HTTPSPort: 443,
			GRPCPort:  9000,
		},

		Metrics: MetricsConfig{
			Enabled:  true,
			Exporter: MetricsPrometheus,
		},

		Tracing: TracingConfig{
			Enabled:       false,
			Exporter:      TracingJaeger,
			SamplingRatio: 1,
			Propagators: []TracingPropagator{
				TracingPropagatorTraceContext,
				TracingPropagatorBaggage,
			},
			Jaeger: JaegerTracingConfig{
				Host: "localhost",
				Port: 6831,
			},
			Zipkin: ZipkinTracingConfig{
				Endpoint: "http://localhost:9411/api/v2/spans",
			},
			OTLP: OTLPTracingConfig{
				Endpoint: "localhost:4317",
			},
		},

		Database: DatabaseConfig{
			URL:                       "file:" + dbPath,
			MaxIdleConn:               2,
			PreparedStatementsEnabled: true,
		},

		Storage: StorageConfig{
			Type: DatabaseStorageType,
		},

		Meta: MetaConfig{
			CheckForUpdates:  true,
			TelemetryEnabled: true,
			StateDirectory:   "",
		},

		Authentication: AuthenticationConfig{
			Session: AuthenticationSession{
				TokenLifetime: 24 * time.Hour,
				StateLifetime: 10 * time.Minute,
			},
		},

		Audit: AuditConfig{
			Sinks: SinksConfig{
				Log: LogSinkConfig{
					Enabled: false,
					File:    "",
				},
				Kafka: KafkaSinkConfig{
					Encoding: "protobuf",
				},
			},
			Buffer: BufferConfig{
				Capacity:    2,
				FlushPeriod: 2 * time.Minute,
			},
			Events: []string{"*:*"},
		},

		Analytics: AnalyticsConfig{
			Buffer: BufferConfig{
				FlushPeriod: 10 * time.Second,
			},
		},

		Authorization: AuthorizationConfig{},
	}
}
