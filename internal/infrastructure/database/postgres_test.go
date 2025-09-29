package database

import (
	"errors"
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type fakeDialector struct{}

func (f fakeDialector) Name() string {
	return "fake"
}

func (f fakeDialector) Initialize(db *gorm.DB) error {
	return errors.New("init error")
}

func TestConnect_InvalidDSN(t *testing.T) {
	_, err := Connect(Config{Host: "localhost", Port: 0, User: "", Password: "", DBName: "", SSLMode: "disable"})
	require.Error(t, err)
}

func TestRunMigrations(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:migrations?mode=memory&cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	require.Error(t, RunMigrations(db, logger))
}
