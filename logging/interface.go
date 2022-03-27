package logging

type LogTuple struct {
	Field string
	Value interface{}
}

type Logger interface {
	WithFields(...LogTuple) Logger
	Debug(string, []LogTuple)
	Info(string, []LogTuple)
	Warn(string, []LogTuple)
	Error(string, []LogTuple)
	Panic(string, []LogTuple)
	DPanic(string, []LogTuple)
}
