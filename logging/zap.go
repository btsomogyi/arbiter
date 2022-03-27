package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ Logger = (*ZapLogger)(nil)

func NewZapLogger(ws zapcore.WriteSyncer, options ...Option) *ZapLogger {
	cfg := zap.NewDevelopmentEncoderConfig()
	zc := zapcore.NewCore(zapcore.NewJSONEncoder(cfg), ws, zap.DebugLevel)
	l := zap.New(zc)
	zl := &ZapLogger{
		logger: l,
		sugar: l.Sugar(),
	}
	for _, option := range options {
		option.apply(zl)
	}
	return zl
}

type ZapLogger struct {
	logger *zap.Logger
	sugar *zap.SugaredLogger
}

// An Option configures a ZapLogger.
type Option interface {
	apply(*ZapLogger)
}

type zapOptionFunc func(*ZapLogger)

func (z zapOptionFunc) apply(log *ZapLogger) {
	z(log)
}

func StdLogRedirect() Option {
	return zapOptionFunc(func(logger *ZapLogger) {
		zap.RedirectStdLog(logger.logger)
	})
}

func Development() Option {
	return zapOptionFunc(func(logger *ZapLogger) {
		logger.logger = logger.logger.WithOptions(zap.Development())
	})
}

func ApplyZapOption(option zap.Option) Option {
	return zapOptionFunc(func(logger *ZapLogger){
		logger.logger = logger.logger.WithOptions(option)
	})
}

func (z ZapLogger) WithFields(tuples ...LogTuple) Logger {
	logger := z.sugar
	for _, tuple := range tuples {
		logger = logger.With(tuple.Field, tuple.Value)
	}
	z.sugar = logger
	return z
}

func (z ZapLogger) StdLogRedirect() {
	zap.RedirectStdLog(z.logger)
}

func (z ZapLogger) Debug(s string, tuples []LogTuple) {
	z.sugar.Debug(s, flatten(tuples))
}

func (z ZapLogger) Info(s string, tuples []LogTuple) {
	z.sugar.Info(s, flatten(tuples))
}

func (z ZapLogger) Warn(s string, tuples []LogTuple) {
	z.sugar.Warn(s, flatten(tuples))
}

func (z ZapLogger) Error(s string, tuples []LogTuple) {
	z.sugar.Error(s, flatten(tuples))
}

func (z ZapLogger) Panic(s string, tuples []LogTuple) {
	z.sugar.Panic(s, flatten(tuples))
}

func (z ZapLogger) DPanic(s string, tuples []LogTuple) {
	z.sugar.DPanic(s, flatten(tuples))
}

func flatten(tuples []LogTuple) []interface{} {
	var flattened []interface{}
	for _, tuple := range tuples {
		flattened = append(flattened, tuple.Field, tuple.Value)
	}
	return flattened
}