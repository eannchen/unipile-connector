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

	viper.AutomaticEnv()
	viper.SetConfigType("env")

	if path != "" {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if err := viper.ReadConfig(bytes.NewBuffer(content)); err != nil {
			return nil, err
		}
	} else {
		// look for config in the working directory
		viper.AddConfigPath(".")
		viper.SetConfigFile(".env")

		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err == nil {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// server
	config.Server.Host = viper.GetString("server_host")
	config.Server.Port = viper.GetString("server_port")
	if len(config.Server.Port) == 0 {
		config.Server.Port = "8080"
	}
	if len(config.Server.Host) == 0 {
		config.Server.Host = "0.0.0.0"
	}

	// database
	config.Database.Host = viper.GetString("db_host")
	config.Database.Port = viper.GetInt("db_port")
	config.Database.User = viper.GetString("db_user")
	config.Database.Password = viper.GetString("db_password")
	config.Database.DBName = viper.GetString("db_name")
	config.Database.SSLMode = viper.GetString("db_sslmode")
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
	config.Unipile.BaseURL = viper.GetString("unipile_base_url")
	config.Unipile.APIKey = viper.GetString("unipile_api_key")
	if config.Unipile.BaseURL == "" {
		config.Unipile.BaseURL = "https://api.unipile.com"
	}

	// redis
	config.Redis.Host = viper.GetString("redis_host")
	config.Redis.Port = viper.GetInt("redis_port")
	config.Redis.Password = viper.GetString("redis_password")
	config.Redis.DB = viper.GetInt("redis_db")
	if config.Redis.Host == "" {
		config.Redis.Host = "localhost"
	}
	if config.Redis.Port == 0 {
		config.Redis.Port = 6379
	}

	// jwt
	config.JWT.SecretKey = viper.GetString("jwt_secret_key")
	config.JWT.Issuer = viper.GetString("jwt_issuer")

	return &config, nil
}
