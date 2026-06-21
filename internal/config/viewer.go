package config

type ViewerLogConfig struct {
	Lines int `yaml:"lines" mapstructure:"lines" flag:"initial log lines to fetch for a container"`
}

type ViewerConfig struct {
	Log ViewerLogConfig `yaml:"log" mapstructure:"log"`
}
