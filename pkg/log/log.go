package log

import (
	"fmt"
	"os"
	"strings"
)

var log Logger

func init() {
	logger, err := newZapLogger()
	log = logger
	if err != nil {
		fmt.Fprintf(os.Stderr, "error initate logger. %s\n", err)
		os.Exit(1)
	}
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

func Warn(args ...interface{}) {
	log.Warn(args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Panicf(format string, args ...interface{}) {
	log.Panicf(format, args...)
}

func Panic(args ...interface{}) {
	log.Panic(args...)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func SetLevel(level Level) {
	log.SetLevel(level)
}

func GetLevelFromStr(level string) Level {
	switch strings.ToUpper(level) {
	case "INFO":
		return INFO
	case "FATAL":
		return FATAL
	case "ERROR":
		return ERROR
	case "WARN":
		return WARN
	case "DEBUG":
		return DEBUG
	default:
		log.Info("user set log level is invalid, using default info level")
		return INFO
	}
}

func GetLogDir() LogDir {
	return log.GetLogDir()
}

func GetLogger() Logger {
	return log
}

func With(args ...interface{}) Logger {
	return log.With(args...)
}
