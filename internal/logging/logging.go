package logging

import (
	"fmt"
	"go.uber.org/zap"
)

var logger *zap.Logger

// GetLogger is a global helper to get the logger object
func GetLogger() *zap.Logger {
	return logger
}

// SetupLogging creates the logger object relevant
// to the current application environment
func SetupLogging(production bool) error {
	var l *zap.Logger
	var err error
	if production {
		if l, err = zap.NewProduction(); err != nil {
			return fmt.Errorf("could not setup production logger: %w", err)
		}
	} else {
		if l, err = zap.NewDevelopment(); err != nil {
			return fmt.Errorf("could not setup development logger: %w", err)
		}
	}

	logger = l

	return nil
}
