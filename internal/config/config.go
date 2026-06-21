package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Logging LoggingConfig `yaml:"log" mapstructure:"log"`
	Viewer  ViewerConfig  `yaml:"viewer" mapstructure:"viewer"`
}

func Default() Config {
	return Config{
		Logging: LoggingConfig{
			Enable: false,
			Level:  "warn",
			Path:   defaultLogPath(),
		},
		Viewer: ViewerConfig{
			Log: ViewerLogConfig{
				Lines: 2000,
			},
		},
	}
}

type fieldInfo struct {
	configKey    string
	flagName     string
	flagUsage    string
	defaultValue any
	fieldPath    []int
}

func discoverFields() []fieldInfo {
	var info []fieldInfo
	d := Default()
	walkFields(reflect.ValueOf(d), reflect.TypeOf(d), nil, "", &info)
	return info
}

func walkFields(val reflect.Value, typ reflect.Type, path []int, prefix string, info *[]fieldInfo) {
	for i := range typ.NumField() {
		field := typ.Field(i)
		tag := field.Tag.Get("mapstructure")
		if tag == "" || tag == "-" {
			continue
		}

		fullKey := tag
		if prefix != "" {
			fullKey = prefix + "." + tag
		}

		fieldPath := append(append([]int{}, path...), i)
		fieldVal := val.Field(i)

		if fieldVal.Kind() == reflect.Struct {
			walkFields(fieldVal, field.Type, fieldPath, fullKey, info)
		} else {
			flagUsage := field.Tag.Get("flag")
			flagName := strings.ReplaceAll(fullKey, ".", "-")
			*info = append(*info, fieldInfo{
				configKey:    fullKey,
				flagName:     flagName,
				flagUsage:    flagUsage,
				defaultValue: fieldVal.Interface(),
				fieldPath:    fieldPath,
			})
		}
	}
}

func SetViperDefaults(v *viper.Viper) {
	for _, f := range discoverFields() {
		v.SetDefault(f.configKey, f.defaultValue)
	}
}

func RegisterFlags(cmd *cobra.Command) {
	for _, f := range discoverFields() {
		if f.flagUsage == "" {
			continue
		}
		switch v := f.defaultValue.(type) {
		case bool:
			cmd.Flags().Bool(f.flagName, v, f.flagUsage)
		case string:
			cmd.Flags().String(f.flagName, v, f.flagUsage)
		case int:
			cmd.Flags().Int(f.flagName, v, f.flagUsage)
		}
	}
}

func BindFlags(v *viper.Viper, cmd *cobra.Command) {
	for _, f := range discoverFields() {
		if f.flagUsage == "" {
			continue
		}
		_ = v.BindPFlag(f.configKey, cmd.Flags().Lookup(f.flagName))
	}
}

func ReadConfig(v *viper.Viper) Config {
	cfg := Default()
	cfgVal := reflect.ValueOf(&cfg).Elem()
	for _, f := range discoverFields() {
		raw := v.Get(f.configKey)
		if raw == nil {
			continue
		}
		field := cfgVal.FieldByIndex(f.fieldPath)
		if !field.IsValid() || !field.CanSet() {
			continue
		}
		switch field.Kind() {
		case reflect.Bool:
			if b, ok := raw.(bool); ok {
				field.SetBool(b)
			}
		case reflect.String:
			if s, ok := raw.(string); ok {
				field.SetString(s)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			switch r := raw.(type) {
			case int:
				field.SetInt(int64(r))
			case int64:
				field.SetInt(r)
			case float64:
				field.SetInt(int64(r))
			}
		}
	}
	return cfg
}

func (c Config) DisplayLines(cfgPath string) []string {
	cfgVal := reflect.ValueOf(c)
	var lines []string
	for _, f := range discoverFields() {
		val := cfgVal.FieldByIndex(f.fieldPath)
		lines = append(lines, fmt.Sprintf("%-20s %v", f.configKey+":", val.Interface()))
	}
	if cfgPath == "" {
		lines = append(lines, "", "config:            (none)")
	} else {
		lines = append(lines, "", "config:            "+cfgPath)
	}
	return lines
}
