package log

import "github.com/sherifabdlnaby/bosun/config"

// Log defines a set of methods for writing application logs.
type Logger interface {
	Extend(name string) Logger
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Debugw(msg string, keysAndValues ...interface{})
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Panic(args ...interface{})
	Panicf(template string, args ...interface{})
	Panicw(msg string, keysAndValues ...interface{})
	Sync() error
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	WithFields(args ...interface{}) Logger
}

func Default() Logger {
	return NewZapLoggerImpl("root", config.Logging{
		Level:  "info",
		Format: "console",
		Debug:  true,
		Color:  false,
	})
}
