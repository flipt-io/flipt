package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"golang.org/x/exp/constraints"
)

var decodeHooks = []mapstructure.DecodeHookFunc{
	mapstructure.StringToTimeDurationHookFunc(),
	stringToSliceHookFunc(),
	stringToEnumHookFunc(stringToLogEncoding),
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
	Version        string               `json:"version,omitempty"`
	Experimental   ExperimentalConfig   `json:"experimental,omitempty" mapstructure:"experimental"`
	Log            LogConfig            `json:"log,omitempty" mapstructure:"log"`
	UI             UIConfig             `json:"ui,omitempty" mapstructure:"ui"`
	Cors           CorsConfig           `json:"cors,omitempty" mapstructure:"cors"`
	Cache          CacheConfig          `json:"cache,omitempty" mapstructure:"cache"`
	Server         ServerConfig         `json:"server,omitempty" mapstructure:"server"`
	Storage        StorageConfig        `json:"storage,omitempty" mapstructure:"storage" experiment:"filesystem_storage"`
	Tracing        TracingConfig        `json:"tracing,omitempty" mapstructure:"tracing"`
	Database       DatabaseConfig       `json:"db,omitempty" mapstructure:"db"`
	Meta           MetaConfig           `json:"meta,omitempty" mapstructure:"meta"`
	Authentication AuthenticationConfig `json:"authentication,omitempty" mapstructure:"authentication"`
	Audit          AuditConfig          `json:"audit,omitempty" mapstructure:"audit"`
}

type Result struct {
	Config   *Config
	Warnings []string
}

func Load(path string) (*Result, error) {
	v := viper.New()
	v.SetEnvPrefix("FLIPT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("loading configuration: %w", err)
	}

	var (
		cfg         = &Config{}
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
			result.Warnings = append(result.Warnings, warning.String())
		}
	}

	// run any defaulters
	for _, defaulter := range defaulters {
		defaulter.setDefaults(v)
	}

	if err := v.Unmarshal(cfg, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			append(decodeHooks, experimentalFieldSkipHookFunc(skippedTypes...))...,
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

type defaulter interface {
	setDefaults(v *viper.Viper)
}

type validator interface {
	validate() error
}

type deprecator interface {
	deprecations(v *viper.Viper) []deprecation
}

// fieldKey returns the name to be used when deriving a fields env var key.
// If marked as squash the key will be the empty string.
// Otherwise, it is derived from the lowercase name of the field.
func fieldKey(field reflect.StructField) string {
	if tag := field.Tag.Get("mapstructure"); tag != "" {
		tag, attr, ok := strings.Cut(tag, ",")
		if !ok || attr == "squash" {
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
		if strings.TrimSpace(c.Version) != "1.0" {
			return fmt.Errorf("invalid version: %s", c.Version)
		}
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
