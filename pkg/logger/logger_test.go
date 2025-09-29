package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	log := NewLogger()
	require.NotNil(t, log)
	require.NotNil(t, log.Out)
	require.NotNil(t, log.Formatter)
}
