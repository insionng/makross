package logger_test

import (
	"github.com/insionng/makross"
	"github.com/insionng/makross/logger"
	"testing"
)

func TestLogger(t *testing.T) {
	// Note: Just for the test coverage, not a real test.
	e := makross.New()
	e.Use(logger.LoggerWithConfig(logger.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	go e.Run(":9000")
}
