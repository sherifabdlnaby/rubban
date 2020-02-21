package log

import (
	"github.com/sherifabdlnaby/rubban/config"
	zaplogfmt "github.com/sykesm/zap-logfmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	doExtend bool
	l        *zap.SugaredLogger
}

//NewZapLoggerImpl Return Zap Instance
func NewZapLoggerImpl(name string, config config.Logging) Logger {

	// Base config
	zapConfig := zap.NewProductionConfig()
	if config.Debug {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// Level
	switch config.Level {
	case "debug":
		zapConfig.Level.SetLevel(zap.DebugLevel)
	case "info":
		zapConfig.Level.SetLevel(zap.InfoLevel)
	case "warn":
		zapConfig.Level.SetLevel(zap.WarnLevel)
	case "fatal":
		zapConfig.Level.SetLevel(zap.FatalLevel)
	case "panic":
		zapConfig.Level.SetLevel(zap.PanicLevel)
	}

	// Format
	switch config.Format {
	case console:
		zapConfig.Encoding = console
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	case json:
		zapConfig.Encoding = json
	case logfmt:
		err := zap.RegisterEncoder(logfmt, func(encoderConfig zapcore.EncoderConfig) (encoder zapcore.Encoder, err error) {
			return zaplogfmt.NewEncoder(encoderConfig), nil
		})
		if err != nil {
			panic(err)
		}
		zapConfig.Encoding = logfmt
	}

	// Color
	if config.Color && config.Format == console {
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logger, err := zapConfig.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}

	return &zapLogger{l: logger.Sugar().Named(name), doExtend: config.Format == json}
}

func (z *zapLogger) Extend(name string) Logger {
	if !z.doExtend {
		return &zapLogger{l: z.l.Named(name)}
	}
	return &zapLogger{l: z.l}
}

func (z *zapLogger) Debug(args ...interface{}) {
	z.l.Debug(args...)
}

func (z *zapLogger) Debugf(template string, args ...interface{}) {
	z.l.Debugf(template, args...)
}

func (z *zapLogger) Debugw(msg string, keysAndValues ...interface{}) {
	z.l.Debugw(msg, keysAndValues...)
}

func (z *zapLogger) Info(args ...interface{}) {
	z.l.Info(args...)
}

func (z *zapLogger) Infof(template string, args ...interface{}) {
	z.l.Infof(template, args...)
}

func (z *zapLogger) Infow(msg string, keysAndValues ...interface{}) {
	z.l.Infow(msg, keysAndValues...)
}

func (z *zapLogger) Warn(args ...interface{}) {
	z.l.Warn(args...)
}

func (z *zapLogger) Warnf(template string, args ...interface{}) {
	z.l.Warnf(template, args...)
}

func (z *zapLogger) Warnw(msg string, keysAndValues ...interface{}) {
	z.l.Warnw(msg, keysAndValues...)
}

func (z *zapLogger) Error(args ...interface{}) {
	z.l.Error(args...)
}

func (z *zapLogger) Errorf(template string, args ...interface{}) {
	z.l.Errorf(template, args...)
}

func (z *zapLogger) Errorw(msg string, keysAndValues ...interface{}) {
	z.l.Errorw(msg, keysAndValues...)
}

func (z *zapLogger) Fatal(args ...interface{}) {
	z.l.Fatal(args...)
}

func (z *zapLogger) Fatalf(template string, args ...interface{}) {
	z.l.Fatalf(template, args...)
}

func (z *zapLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	z.l.Fatalw(msg, keysAndValues...)
}

func (z *zapLogger) Panic(args ...interface{}) {
	z.l.Panic(args...)
}

func (z *zapLogger) Panicf(template string, args ...interface{}) {
	z.l.Panicf(template, args...)
}

func (z *zapLogger) Panicw(msg string, keysAndValues ...interface{}) {
	z.l.Panicw(msg, keysAndValues...)
}

func (z *zapLogger) Sync() error {
	return z.l.Sync()
}

func (z *zapLogger) WithFields(args ...interface{}) Logger {
	return &zapLogger{l: z.l.With(args...)}
}
