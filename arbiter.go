package arbiter

import (
	"github.com/btsomogyi/arbiter/internal"
	"github.com/btsomogyi/arbiter/logging"
	"github.com/btsomogyi/arbiter/telemetry"
)

func NewSupervisor(opts ...internal.SupervisorOption) (*internal.Supervisor, error) {
	return internal.NewSupervisor(opts...)
}

// SetInstrumentor sets the Instrumentor for the Supervisor.
func SetInstrumentor(i telemetry.Instrumentor) internal.SupervisorOption {
	return internal.SetInstrumentor(i)
}

// SetChannelDepth sets the maximum size of the queue channel.
func SetChannelDepth(d uint) internal.SupervisorOption {
	return internal.SetChannelDepth(d)
}

// SetLogger provides a compatible structured logger for emitting log messages.
// If not provided, a no-op logger is created during supervisor initialization.
func SetLogger(l logging.Logger) internal.SupervisorOption {
	return internal.SetLogger(l)
}

// SetPollFunction sets the pollDone function in Supervisor for deterministic testing.
func SetPollFunction(f func()) internal.SupervisorOption {
	return internal.SetPollFunction(f)
}
