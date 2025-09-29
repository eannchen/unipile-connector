package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadWithDefaults(t *testing.T) {
	config, err := Load("")
	require.NoError(t, err)
	require.Equal(t, "0.0.0.0", config.Server.Host)
	require.Equal(t, "8080", config.Server.Port)
	require.Equal(t, "warn", config.Log.Level)
	require.Equal(t, "localhost", config.Database.Host)
	require.Equal(t, 5432, config.Database.Port)
	require.Equal(t, "postgres", config.Database.User)
	require.Equal(t, "password", config.Database.Password)
	require.Equal(t, "unipile_connector", config.Database.DBName)
	require.Equal(t, "disable", config.Database.SSLMode)
	require.Equal(t, "https://api.unipile.com", config.Unipile.BaseURL)
	require.Equal(t, "localhost", config.Redis.Host)
	require.Equal(t, 6379, config.Redis.Port)
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".env")
	envContent := `SERVER_HOST=127.0.0.1
SERVER_PORT=3000
LOG_LEVEL=info
DB_HOST=db.example.com
DB_PORT=6543
DB_USER=dbuser
DB_PASSWORD=secret
DB_NAME=custom_db
DB_SSLMODE=require
UNIPILE_BASE_URL=https://custom.unipile
UNIPILE_API_KEY=apikey
REDIS_HOST=redis.example.com
REDIS_PORT=6380
REDIS_PASSWORD=redispass
REDIS_DB=2
JWT_SECRET_KEY=supersecret
JWT_ISSUER=test-issuer
`
	require.NoError(t, os.WriteFile(configPath, []byte(envContent), 0o600))

	config, err := Load(configPath)
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1", config.Server.Host)
	require.Equal(t, "3000", config.Server.Port)
	require.Equal(t, "info", config.Log.Level)
	require.Equal(t, "db.example.com", config.Database.Host)
	require.Equal(t, 6543, config.Database.Port)
	require.Equal(t, "dbuser", config.Database.User)
	require.Equal(t, "secret", config.Database.Password)
	require.Equal(t, "custom_db", config.Database.DBName)
	require.Equal(t, "require", config.Database.SSLMode)
	require.Equal(t, "https://custom.unipile", config.Unipile.BaseURL)
	require.Equal(t, "apikey", config.Unipile.APIKey)
	require.Equal(t, "redis.example.com", config.Redis.Host)
	require.Equal(t, 6380, config.Redis.Port)
	require.Equal(t, "redispass", config.Redis.Password)
	require.Equal(t, 2, config.Redis.DB)
	require.Equal(t, "supersecret", config.JWT.SecretKey)
	require.Equal(t, "test-issuer", config.JWT.Issuer)
}
