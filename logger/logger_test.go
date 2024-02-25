package logger_test

import (
	"testing"

	"bbox/logger"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestInitializeLogger(t *testing.T) {
	testCases := []struct {
		name          string
		inputLevel    string
		expectedLevel logrus.Level
	}{
		{"DebugLevel", "debug", logrus.DebugLevel},
		{"InfoLevel", "info", logrus.InfoLevel},
		{"WarnLevel", "warn", logrus.WarnLevel},
		{"ErrorLevel", "error", logrus.ErrorLevel},
		{"FatalLevel", "fatal", logrus.FatalLevel},
		{"PanicLevel", "panic", logrus.PanicLevel},
		{"DefaultLevel", "unknown", logrus.WarnLevel},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger.InitializeLogger(tc.inputLevel)
			assert.Equal(t, tc.expectedLevel, logrus.GetLevel())
		})
	}
}
