package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"golang.org/x/exp/constraints"
)

var decodeHooks = mapstructure.ComposeDecodeHookFunc(
	mapstructure.StringToTimeDurationHookFunc(),
	mapstructure.StringToSliceHookFunc(","),
	StringToEnumHookFunc(stringToLogEncoding),
	StringToEnumHookFunc(stringToCacheBackend),
	StringToEnumHookFunc(stringToScheme),
	StringToEnumHookFunc(stringToDatabaseProtocol),
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
	Log      LogConfig      `json:"log,omitempty" mapstructure:"log"`
	UI       UIConfig       `json:"ui,omitempty" mapstructure:"ui"`
	Cors     CorsConfig     `json:"cors,omitempty" mapstructure:"cors"`
	Cache    CacheConfig    `json:"cache,omitempty" mapstructure:"cache"`
	Server   ServerConfig   `json:"server,omitempty" mapstructure:"server"`
	Tracing  TracingConfig  `json:"tracing,omitempty" mapstructure:"tracing"`
	Database DatabaseConfig `json:"database,omitempty" mapstructure:"db"`
	Meta     MetaConfig     `json:"meta,omitempty" mapstructure:"meta"`
	Warnings []string       `json:"warnings,omitempty"`
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
		cfg    = &Config{}
		fields = cfg.fields()
	)

	// set viper defaults per field
	for _, defaulter := range fields.defaulters {
		cfg.Warnings = append(cfg.Warnings, defaulter.setDefaults(v)...)
	}

	if err := v.Unmarshal(cfg, viper.DecodeHook(decodeHooks)); err != nil {
		return nil, err
	}

	// run any validation steps
	for _, validator := range fields.validators {
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

type fields struct {
	defaulters []defaulter
	validators []validator
}

func (c *Config) fields() (fields fields) {
	structVal := reflect.ValueOf(c).Elem()
	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i).Addr().Interface()

		if defaulter, ok := field.(defaulter); ok {
			fields.defaulters = append(fields.defaulters, defaulter)
		}

		if validator, ok := field.(validator); ok {
			fields.validators = append(fields.validators, validator)
		}
	}

	return
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

// StringToEnumHookFunc returns a DecodeHookFunc that converts strings to a target enum
func StringToEnumHookFunc[T constraints.Integer](mappings map[string]T) mapstructure.DecodeHookFunc {
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

		enum, _ := mappings[data.(string)]

		return enum, nil
	}
}
