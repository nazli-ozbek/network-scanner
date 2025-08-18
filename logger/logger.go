package logger

import "github.com/sirupsen/logrus"

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Warn(args ...interface{})
}

type LogrusAdapter struct{}

func NewLogger() Logger {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.DebugLevel)
	return &LogrusAdapter{}
}

func (l *LogrusAdapter) Info(args ...interface{})  { logrus.Info(args...) }
func (l *LogrusAdapter) Error(args ...interface{}) { logrus.Error(args...) }
func (l *LogrusAdapter) Debug(args ...interface{}) { logrus.Debug(args...) }
func (l *LogrusAdapter) Warn(args ...interface{})  { logrus.Warn(args...) }
