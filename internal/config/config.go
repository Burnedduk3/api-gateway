package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Environment string         `mapstructure:"environment"`
	Version     string         `mapstructure:"version"`
	LogLevel    string         `mapstructure:"loglevel"`
	Server      ServerConfig   `mapstructure:"server"`
	Database    DatabaseConfig `mapstructure:"database"`
	Security    SecurityConfig `mapstructure:"security"`
	Logging     LoggingConfig  `mapstructure:"logging"`
}

type ServerConfig struct {
	Port            string        `mapstructure:"port"`
	Host            string        `mapstructure:"host"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	CORS            CORSConfig    `mapstructure:"cors"`
}

type CORSConfig struct {
	AllowOrigins []string `mapstructure:"allow_origins"`
	AllowMethods []string `mapstructure:"allow_methods"`
	AllowHeaders []string `mapstructure:"allow_headers"`
}

type SecurityConfig struct {
	RateLimitRPS   int `mapstructure:"rate_limit_rps"`
	RateLimitBurst int `mapstructure:"rate_limit_burst"`
}

func Load(configFile, env string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Configure viper
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")
	v.AddConfigPath("/etc/product-service")

	// Use specific config file if provided
	if configFile != "" {
		v.SetConfigFile(configFile)
	}

	// Environment variables
	v.SetEnvPrefix("PRODUCT_SERVICE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	// Read config file
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	version := v.GetString("VERSION")

	// Override environment
	v.Set("environment", env)
	v.Set("version", version)

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("version", "0.0.1")
	v.SetDefault("loglevel", "info")
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.read_timeout", 15*time.Second)
	v.SetDefault("server.write_timeout", 30*time.Second)
	v.SetDefault("server.shutdown_timeout", 30*time.Second)
	v.SetDefault("server.cors.allow_origins", []string{"*"})
	v.SetDefault("server.cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("server.cors.allow_headers", []string{"*"})

	DatabaseDefaults(v)

	v.SetDefault("security.rate_limit_rps", 100)
	v.SetDefault("security.rate_limit_burst", 200)

	DefaultLogger(v)
}
