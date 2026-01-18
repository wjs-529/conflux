package logger

import (
	"fmt"
	"os"

	"github.com/veil-net/conflux/anchor"
	"github.com/hashicorp/go-hclog"
)

// CustomLogger wraps hclog.CustomLogger to provide a zap-like interface
type CustomLogger struct {
	hclog.Logger
}

// Sugar provides a zap-like sugar interface
type Sugar struct {
	logger hclog.Logger
}

func (s *Sugar) Infof(format string, args ...interface{}) {
	s.logger.Info(fmt.Sprintf(format, args...))
}

func (s *Sugar) Errorf(format string, args ...interface{}) {
	s.logger.Error(fmt.Sprintf(format, args...))
}

func (s *Sugar) Warnf(format string, args ...interface{}) {
	s.logger.Warn(fmt.Sprintf(format, args...))
}

func (s *Sugar) Debugf(format string, args ...interface{}) {
	s.logger.Debug(fmt.Sprintf(format, args...))
}

func (s *Sugar) Fatalf(format string, args ...interface{}) {
	s.logger.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

func (l *CustomLogger) Sugar() *Sugar {
	return &Sugar{logger: l.Logger}
}

// Global logger instance
var Logger *CustomLogger

// init initializes the hclogger and sets up the global logger
func init() {
	Logger = &CustomLogger{Logger: anchor.HCLogger}
}
