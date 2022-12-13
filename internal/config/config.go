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

var decodeHooks = mapstructure.ComposeDecodeHookFunc(
	mapstructure.StringToTimeDurationHookFunc(),
	stringToSliceHookFunc(),
	stringToEnumHookFunc(stringToLogEncoding),
	stringToEnumHookFunc(stringToCacheBackend),
	stringToEnumHookFunc(stringToScheme),
	stringToEnumHookFunc(stringToDatabaseProtocol),
	stringToEnumHookFunc(stringToAuthMethod),
)

// Config contains all of Flipts configuration needs.
//
// The root of this structure contains a collection of sub-configuration categories,
// along with a set of warnings derived once the configuration has been loaded.
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
	Log            LogConfig            `json:"log,omitempty" mapstructure:"log"`
	UI             UIConfig             `json:"ui,omitempty" mapstructure:"ui"`
	Cors           CorsConfig           `json:"cors,omitempty" mapstructure:"cors"`
	Cache          CacheConfig          `json:"cache,omitempty" mapstructure:"cache"`
	Server         ServerConfig         `json:"server,omitempty" mapstructure:"server"`
	Tracing        TracingConfig        `json:"tracing,omitempty" mapstructure:"tracing"`
	Database       DatabaseConfig       `json:"db,omitempty" mapstructure:"db"`
	Meta           MetaConfig           `json:"meta,omitempty" mapstructure:"meta"`
	Authentication AuthenticationConfig `json:"authentication,omitempty" mapstructure:"authentication"`
	Warnings       []string             `json:"warnings,omitempty"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("FLIPT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("loading configuration: %w", err)
	}

	var (
		cfg        = &Config{}
		validators = cfg.prepare(v)
	)

	if err := v.Unmarshal(cfg, viper.DecodeHook(decodeHooks)); err != nil {
		return nil, err
	}

	// run any validation steps
	for _, validator := range validators {
		if err := validator.validate(); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

type defaulter interface {
	setDefaults(v *viper.Viper) []string
}

type validator interface {
	validate() error
}

func (c *Config) prepare(v *viper.Viper) (validators []validator) {
	val := reflect.ValueOf(c).Elem()
	for i := 0; i < val.NumField(); i++ {
		// search for all expected env vars since Viper cannot
		// infer when doing Unmarshal + AutomaticEnv.
		// see: https://github.com/spf13/viper/issues/761
		structField := val.Type().Field(i)
		bindEnvVars(v, getEnvVarTrie(), fieldKey(structField), structField.Type)

		field := val.Field(i).Addr().Interface()

		// for-each defaulter implementing fields we invoke
		// setting any defaults during this prepare stage
		// on the supplied viper.
		if defaulter, ok := field.(defaulter); ok {
			c.Warnings = append(c.Warnings, defaulter.setDefaults(v)...)
		}

		// for-each validator implementing field we collect
		// them up and return them to be validated after
		// unmarshalling.
		if validator, ok := field.(validator); ok {
			validators = append(validators, validator)
		}
	}

	return
}
func fieldKey(field reflect.StructField) string {
	if tag := field.Tag.Get("mapstructure"); tag != "" {
		return tag
	}

	return strings.ToLower(field.Name)
}

// bindEnvVars descends into the provided struct field binding any expected
// environment variable keys it finds reflecting struct and field tags.
func bindEnvVars(v *viper.Viper, prefixes trie, prefix string, typ reflect.Type) {
	// descend through pointers
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	// descend into struct fields
	switch typ.Kind() {
	case reflect.Map:
		// use the environment to pre-emptively bind env vars for map keys
		parts := strings.Split(strings.ReplaceAll(strings.ToUpper(prefix), "_", "."), ".")
		if trie, ok := prefixes.getPrefix(parts...); ok {
			for k := range trie {
				prefix = prefix + "." + strings.ToLower(k)
				bindEnvVars(v, prefixes, prefix, typ.Elem())
			}
		}

		return
	case reflect.Struct:
		for i := 0; i < typ.NumField(); i++ {
			structField := typ.Field(i)

			// key becomes prefix for sub-fields
			prefix = prefix + "." + fieldKey(structField)
			bindEnvVars(v, prefixes, prefix, structField.Type)
		}

		return
	}

	v.MustBindEnv(prefix)
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

// trie is a prefix trie used to create a map of potential
// key paths from the current environment.
// This is used to support pre-emptively binding env vars
// for the keys of map[string]T which we don't know aheaad of
// time.
type trie map[string]trie

func (t trie) getPrefix(parts ...string) (trie, bool) {
	if len(parts) == 0 {
		return t, true
	}

	return t[parts[0]].getPrefix(parts[1:]...)
}

func getEnvVarTrie() trie {
	const envPrefix = "FLIPT_"
	root := trie{}

	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, envPrefix) {
			continue
		}

		env = strings.SplitN(env[len(envPrefix):], "=", 2)[0]

		node := root
		for _, part := range strings.Split(strings.ToUpper(env), "_") {
			if n, ok := node[part]; ok {
				node = n
				continue
			}

			node[part] = trie{}
			node = node[part]
		}
	}

	return root
}
