package telemetry

// NopInstrumentor satisfies Instrumentor interface with no-op functions.
type NopInstrumentor struct{}

// LocalInstrumenter implements the Instrumentor interface using in-memory metrics.
var _ Instrumentor = (*NopInstrumentor)(nil)

// NewNopInstrumentor returns a No-op Instrumentor.
func NewNopInstrumentor() NopInstrumentor {
	return NopInstrumentor{}
}

func (ni NopInstrumentor) QueueChanDepth(_ int64) {
}

func (ni NopInstrumentor) IncQueueChanDepth() {
}

func (ni NopInstrumentor) DecQueueChanDepth() {
}

func (ni NopInstrumentor) ProcessingMapDepth(_ int64) {
}

func (ni NopInstrumentor) IncProcessingMapDepth() {
}

func (ni NopInstrumentor) DecProcessingMapDepth() {
}

func (ni NopInstrumentor) WaitingMapDepth(_ int64) {
}

func (ni NopInstrumentor) IncWaitingMapDepth() {
}

func (ni NopInstrumentor) DecWaitingMapDepth() {
}

func (ni NopInstrumentor) Messages(_ float64, _ ...Labels) {
}

func (ni NopInstrumentor) Worktime(_ float64, _ ...Labels) {
}

func (ni NopInstrumentor) Transactions(_ float64, _ ...Labels) {
}
