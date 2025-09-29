package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application (for backward compatibility)
type Config struct {
	Server   ServerConfig
	Log      LogConfig
	Database DatabaseConfig
	Unipile  UnipileConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Host string
}

// LogConfig holds log configuration
type LogConfig struct {
	Level string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// UnipileConfig holds Unipile API configuration
type UnipileConfig struct {
	BaseURL string
	APIKey  string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey string
	Issuer    string
}

// Load loads configuration from .env file and environment variables
func Load(path string) (*Config, error) {
	var config Config

	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigType("env")

	if path != "" {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if err := v.ReadConfig(bytes.NewBuffer(content)); err != nil {
			return nil, err
		}
	} else {
		// look for config in the working directory
		v.AddConfigPath(".")
		v.SetConfigFile(".env")

		// If a config file is found, read it in.
		if err := v.ReadInConfig(); err == nil {
			fmt.Println("Using config file:", v.ConfigFileUsed())
		}
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// server
	config.Server.Host = v.GetString("server_host")
	config.Server.Port = v.GetString("server_port")
	if len(config.Server.Port) == 0 {
		config.Server.Port = "8080"
	}
	if len(config.Server.Host) == 0 {
		config.Server.Host = "0.0.0.0"
	}

	// log
	config.Log.Level = v.GetString("log_level")
	if config.Log.Level == "" {
		config.Log.Level = "warn"
	}

	// database
	config.Database.Host = v.GetString("db_host")
	config.Database.Port = v.GetInt("db_port")
	config.Database.User = v.GetString("db_user")
	config.Database.Password = v.GetString("db_password")
	config.Database.DBName = v.GetString("db_name")
	config.Database.SSLMode = v.GetString("db_sslmode")
	if config.Database.Host == "" {
		config.Database.Host = "localhost"
	}
	if config.Database.Port == 0 {
		config.Database.Port = 5432
	}
	if config.Database.User == "" {
		config.Database.User = "postgres"
	}
	if config.Database.Password == "" {
		config.Database.Password = "password"
	}
	if config.Database.DBName == "" {
		config.Database.DBName = "unipile_connector"
	}
	if config.Database.SSLMode == "" {
		config.Database.SSLMode = "disable"
	}

	// unipile
	config.Unipile.BaseURL = v.GetString("unipile_base_url")
	config.Unipile.APIKey = v.GetString("unipile_api_key")
	if config.Unipile.BaseURL == "" {
		config.Unipile.BaseURL = "https://api.unipile.com"
	}

	// redis
	config.Redis.Host = v.GetString("redis_host")
	config.Redis.Port = v.GetInt("redis_port")
	config.Redis.Password = v.GetString("redis_password")
	config.Redis.DB = v.GetInt("redis_db")
	if config.Redis.Host == "" {
		config.Redis.Host = "localhost"
	}
	if config.Redis.Port == 0 {
		config.Redis.Port = 6379
	}

	// jwt
	config.JWT.SecretKey = v.GetString("jwt_secret_key")
	config.JWT.Issuer = v.GetString("jwt_issuer")

	return &config, nil
}
