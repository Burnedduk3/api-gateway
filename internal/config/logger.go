package config

import "github.com/spf13/viper"

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func DefaultLogger(v *viper.Viper) {
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
}
