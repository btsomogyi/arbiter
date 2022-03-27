package logging

var _ Logger = (*NoopLogger)(nil)

func NewNoopLogger() NoopLogger {
	return NoopLogger{}
}

type NoopLogger struct {
}

func (n NoopLogger) WithFields(tuples ...LogTuple) Logger {
	return n
}

func (n NoopLogger) Debug(s string, tuples []LogTuple) {
}

func (n NoopLogger) Info(s string, tuples []LogTuple) {
}

func (n NoopLogger) Warn(s string, tuples []LogTuple) {
}

func (n NoopLogger) Error(s string, tuples []LogTuple) {
}

func (n NoopLogger) Panic(s string, tuples []LogTuple) {
}

func (n NoopLogger) DPanic(s string, tuples []LogTuple) {
}
