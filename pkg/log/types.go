package log

type Level uint8

const (
	FATAL Level = iota
	ERROR
	WARN
	INFO
	DEBUG
)

type Logger interface {
	Debugf(format string, args ...interface{})
	Debug(args ...interface{})
	Infof(format string, args ...interface{})
	Info(args ...interface{})
	Warnf(format string, args ...interface{})
	Warn(args ...interface{})
	Errorf(format string, args ...interface{})
	Error(args ...interface{})
	Panicf(format string, args ...interface{})
	Panic(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatal(args ...interface{})
	SetLevel(level Level)
	GetLogDir() LogDir
	With(args ...interface{}) Logger
}
