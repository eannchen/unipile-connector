package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

var (
	// Env environment
	Env  Environment
	once sync.Once
)

// InitEnvironment init env
func InitEnvironment(confPath string) error {
	var err error
	once.Do(func() {
		Env, err = loadEnvironment(confPath)
	})
	return err
}

// Load loads configuration from .env file and environment variables
func Load() (*Config, error) {
	if err := InitEnvironment(""); err != nil {
		return nil, err
	}

	// Convert Environment to Config for backward compatibility
	config := &Config{
		Server: ServerConfig{
			Port: Env.Server.Port,
			Host: Env.Server.Host,
		},
		Database: DatabaseConfig{
			Host:     Env.Database.Host,
			Port:     Env.Database.Port,
			User:     Env.Database.User,
			Password: Env.Database.Password,
			DBName:   Env.Database.DBName,
			SSLMode:  Env.Database.SSLMode,
		},
		Unipile: UnipileConfig{
			BaseURL: Env.Unipile.BaseURL,
			APIKey:  Env.Unipile.APIKey,
		},
		Redis: RedisConfig{
			Host:     Env.Redis.Host,
			Port:     Env.Redis.Port,
			Password: Env.Redis.Password,
			DB:       Env.Redis.DB,
		},
	}

	return config, nil
}

type Environment struct {
	Server   sectionServer
	Database sectionDatabase
	Unipile  sectionUnipile
	Redis    sectionRedis
}

type sectionServer struct {
	Host string
	Port string
}

type sectionDatabase struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type sectionUnipile struct {
	BaseURL string
	APIKey  string
}

type sectionRedis struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func loadEnvironment(path string) (Environment, error) {
	var env Environment

	viper.AutomaticEnv()
	viper.SetConfigType("env")

	if path != "" {
		content, err := os.ReadFile(path)
		if err != nil {
			return Environment{}, err
		}
		if err := viper.ReadConfig(bytes.NewBuffer(content)); err != nil {
			return Environment{}, err
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
	env.Server.Host = viper.GetString("server_host")
	env.Server.Port = viper.GetString("server_port")
	if len(env.Server.Port) == 0 {
		env.Server.Port = "8080"
	}
	if len(env.Server.Host) == 0 {
		env.Server.Host = "0.0.0.0"
	}

	// database
	env.Database.Host = viper.GetString("db_host")
	env.Database.Port = viper.GetInt("db_port")
	env.Database.User = viper.GetString("db_user")
	env.Database.Password = viper.GetString("db_password")
	env.Database.DBName = viper.GetString("db_name")
	env.Database.SSLMode = viper.GetString("db_sslmode")
	if env.Database.Host == "" {
		env.Database.Host = "localhost"
	}
	if env.Database.Port == 0 {
		env.Database.Port = 5432
	}
	if env.Database.User == "" {
		env.Database.User = "postgres"
	}
	if env.Database.Password == "" {
		env.Database.Password = "password"
	}
	if env.Database.DBName == "" {
		env.Database.DBName = "unipile_connector"
	}
	if env.Database.SSLMode == "" {
		env.Database.SSLMode = "disable"
	}

	// unipile
	env.Unipile.BaseURL = viper.GetString("unipile_base_url")
	env.Unipile.APIKey = viper.GetString("unipile_api_key")
	if env.Unipile.BaseURL == "" {
		env.Unipile.BaseURL = "https://api.unipile.com"
	}

	// redis
	env.Redis.Host = viper.GetString("redis_host")
	env.Redis.Port = viper.GetInt("redis_port")
	env.Redis.Password = viper.GetString("redis_password")
	env.Redis.DB = viper.GetInt("redis_db")
	if env.Redis.Host == "" {
		env.Redis.Host = "localhost"
	}
	if env.Redis.Port == 0 {
		env.Redis.Port = 6379
	}

	return env, nil
}

// Config holds all configuration for the application (for backward compatibility)
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Unipile  UnipileConfig
	Redis    RedisConfig
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
