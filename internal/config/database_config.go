package config

import (
	"time"

	"github.com/spf13/viper"
)

type DatabaseConfig struct {
	Host         string        `mapstructure:"host"`
	Port         string        `mapstructure:"port"`
	Username     string        `mapstructure:"username"`
	Password     string        `mapstructure:"password"`
	Database     string        `mapstructure:"database"`
	SSLMode      string        `mapstructure:"ssl_mode"`
	MaxOpenConns int           `mapstructure:"max_open_conns"`
	MaxIdleConns int           `mapstructure:"max_idle_conns"`
	MaxLifetime  time.Duration `mapstructure:"max_lifetime"`
}

func DatabaseDefaults(v *viper.Viper) {
	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", "5432")
	v.SetDefault("database.username", "product-service")
	v.SetDefault("database.password", "admin")
	v.SetDefault("database.database", "product-service")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 25)
	v.SetDefault("database.max_lifetime", 5*time.Minute)
}
