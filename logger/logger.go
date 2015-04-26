package logger

import "github.com/Sirupsen/logrus"

type Logger struct {
	logrus.Logger
}

var loggerIntance *Logger = nil

func New() *Logger {
	if loggerIntance == nil {
		loggerIntance = &Logger{*logrus.New()}
	}
	return loggerIntance
}

// Implements internface resouce.Logger . Logrus only need mapping for the Criticalf

func (l *Logger) Criticalf(format string, args ...interface{}) {
	l.Fatalf(format, args)
}
