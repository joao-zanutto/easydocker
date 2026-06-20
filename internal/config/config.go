package config

type LoggingConfig struct {
	Enable bool   `yaml:"enable" mapstructure:"enable"`
	Level  string `yaml:"level" mapstructure:"level"`
	Path   string `yaml:"path" mapstructure:"path"`
}

type Config struct {
	Logging LoggingConfig `yaml:"log" mapstructure:"log"`
}

func Default() Config {
	return Config{
		Logging: LoggingConfig{
			Enable: false,
			Level:  "warn",
			Path:   "/var/log/easydocker.log",
		},
	}
}
