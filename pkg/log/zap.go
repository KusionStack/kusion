package log

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"kusionstack.io/kusion/pkg/util/kfile"
)

var (
	Folder    = "logs"
	dirConfig Dir
)

type Dir struct {
	DefaultLogDir string
	ErrorLogDir   string
	DebugLogDir   string
}

type zapLogger struct {
	sugaredLogger *zap.SugaredLogger
	debugLevel    zap.AtomicLevel
	defaultLevel  zap.AtomicLevel
	errorLevel    zap.AtomicLevel
}

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func newZapCore(logDir string, level zapcore.Level) (zapcore.Core, zap.AtomicLevel) {
	atom := zap.NewAtomicLevel()
	atom.SetLevel(level)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:  logDir,
		MaxSize:   10,
		Compress:  false,
		LocalTime: true,
		MaxAge:    28,
	})
	return zapcore.NewCore(encoder, writer, atom), atom
}

func newZapLogger() (Logger, error) {
	kusionDataDir, _ := kfile.KusionDataFolder()
	if v := os.Getenv("LOG_DIR"); v != "" {
		kusionDataDir = v
	}
	dirConfig = Dir{
		DefaultLogDir: filepath.Join(kusionDataDir, Folder, "kusion.log"),
		ErrorLogDir:   filepath.Join(kusionDataDir, Folder, "kusion_error.log"),
		DebugLogDir:   filepath.Join(kusionDataDir, Folder, "kusion_debug.log"),
	}
	debugCore, debugAtom := newZapCore(dirConfig.DebugLogDir, zapcore.DebugLevel)
	defaultCore, defaultAtom := newZapCore(dirConfig.DefaultLogDir, zapcore.InfoLevel)
	errorCore, errorAtom := newZapCore(dirConfig.ErrorLogDir, zapcore.ErrorLevel)
	combinedCore := zapcore.NewTee(defaultCore, errorCore, debugCore)
	// AddCallerSkip skips 2 number of callers since the file that gets
	// logged will always be the wrapped file.
	logger := zap.New(combinedCore, zap.AddCallerSkip(2), zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).Sugar()
	return &zapLogger{sugaredLogger: logger, defaultLevel: defaultAtom, debugLevel: debugAtom, errorLevel: errorAtom}, nil
}

func (l *zapLogger) Debug(args ...interface{}) {
	l.sugaredLogger.Debug(args...)
}

func (l *zapLogger) Debugf(format string, args ...interface{}) {
	l.sugaredLogger.Debugf(format, args...)
}

func (l *zapLogger) Infof(format string, args ...interface{}) {
	l.sugaredLogger.Infof(format, args...)
}

func (l *zapLogger) Info(args ...interface{}) {
	l.sugaredLogger.Info(args...)
}

func (l *zapLogger) Warnf(format string, args ...interface{}) {
	l.sugaredLogger.Warnf(format, args...)
}

func (l *zapLogger) Warn(args ...interface{}) {
	l.sugaredLogger.Warn(args...)
}

func (l *zapLogger) Errorf(format string, args ...interface{}) {
	l.sugaredLogger.Errorf(format, args...)
}

func (l *zapLogger) Error(args ...interface{}) {
	l.sugaredLogger.Error(args...)
}

func (l *zapLogger) Panicf(format string, args ...interface{}) {
	l.sugaredLogger.Panicf(format, args...)
}

func (l *zapLogger) Panic(args ...interface{}) {
	l.sugaredLogger.Panic(args...)
}

func (l *zapLogger) Fatalf(format string, args ...interface{}) {
	l.sugaredLogger.Fatalf(format, args...)
}

func (l *zapLogger) Fatal(args ...interface{}) {
	l.sugaredLogger.Fatal(args...)
}

func (l *zapLogger) SetLevel(level Level) {
	userLevel := getZapLevel(level)
	l.debugLevel.SetLevel(userLevel)

	if userLevel > zapcore.InfoLevel {
		l.defaultLevel.SetLevel(userLevel)
	}
	if userLevel > zapcore.ErrorLevel {
		l.defaultLevel.SetLevel(userLevel)
	}
}

func (l *zapLogger) GetLogDir() Dir {
	return dirConfig
}

func (l *zapLogger) With(args ...interface{}) Logger {
	curLogger := l.sugaredLogger.With(args...)
	return &zapLogger{sugaredLogger: curLogger, defaultLevel: l.defaultLevel, debugLevel: l.debugLevel, errorLevel: l.errorLevel}
}

func getZapLevel(level Level) zapcore.Level {
	switch level {
	case INFO:
		return zapcore.InfoLevel
	case WARN:
		return zapcore.WarnLevel
	case DEBUG:
		return zapcore.DebugLevel
	case ERROR:
		return zapcore.ErrorLevel
	case FATAL:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}
