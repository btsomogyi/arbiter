package telemetry

import (
	"sync"
)

// localBuckets defines quints for local histogram, with '0' representing infinity.
var localBuckets = []float64{
	0.00002, 0.00004, 0.00008, 0.00016, 0.00032, 0.00064, 0.00128, 0.00256, 0.00512, 0.01024,
	0.02048, 0.04096, 0.08192, 0.16384, 0.32768, 0.65536, 1.31072, 2.62144, 5.24288, 0.00000,
}

// LocalInstrumentor implements internal logging (used in tests or for logged metrics).
type LocalInstrumentor struct {
	gauges     map[MetricGauge]int64
	histograms map[MetricHistogram]map[float64]int64
	atomic     sync.Mutex
}

// LocalInstrumenter implements the Instrumentor interface using in-memory metrics.
var _ Instrumentor = (*LocalInstrumentor)(nil)

// NewLocalInstrumentor creates a localInstrumentor.
func NewLocalInstrumentor() *LocalInstrumentor {
	g := make(map[MetricGauge]int64)
	h := make(map[MetricHistogram]map[float64]int64)
	li := LocalInstrumentor{
		gauges:     g,
		histograms: h,
	}
	for name := range MetricGauges {
		li.gauges[name] = 0
	}
	for name := range MetricHistograms {
		li.histograms[name] = make(map[float64]int64)
	}
	return &li
}

// supervisorMetrics contains the telemetry data available to report on the throughput and queue depths
// of the arbiter server process.
type supervisorMetrics struct {
	gauges     map[MetricGauge]int64
	histograms map[MetricHistogram]map[float64]int64
	atomic     sync.Mutex
}

func (li *LocalInstrumentor) QueueChanDepth(value int64) {
	li.setGauge(QueueChanDepth, value)
}

func (li *LocalInstrumentor) IncQueueChanDepth() {
	li.incGauge(QueueChanDepth)
}

func (li *LocalInstrumentor) DecQueueChanDepth() {
	li.decGauge(QueueChanDepth)
}

func (li *LocalInstrumentor) ProcessingMapDepth(value int64) {
	li.setGauge(ProcessingMapDepth, value)
}

func (li *LocalInstrumentor) IncProcessingMapDepth() {
	li.incGauge(ProcessingMapDepth)
}

func (li *LocalInstrumentor) DecProcessingMapDepth() {
	li.decGauge(ProcessingMapDepth)
}

func (li *LocalInstrumentor) WaitingMapDepth(value int64) {
	li.setGauge(WaitingMapDepth, value)
}

func (li *LocalInstrumentor) IncWaitingMapDepth() {
	li.incGauge(WaitingMapDepth)
}

func (li *LocalInstrumentor) DecWaitingMapDepth() {
	li.decGauge(WaitingMapDepth)
}

func (li *LocalInstrumentor) Messages(value float64, _ ...Labels) {
	li.addHistogramEntry(Messages, value)
}

func (li *LocalInstrumentor) Worktime(value float64, _ ...Labels) {
	li.addHistogramEntry(Worktime, value)
}

func (li *LocalInstrumentor) Transactions(value float64, _ ...Labels) {
	li.addHistogramEntry(Transactions, value)
}

// setGauge sets the parameter metric.
func (li *LocalInstrumentor) setGauge(m MetricGauge, value int64) {
	li.atomic.Lock()
	li.gauges[m] = value
	li.atomic.Unlock()
}

// setGauge sets the parameter metric.
func (li *LocalInstrumentor) incGauge(m MetricGauge) {
	li.atomic.Lock()
	li.gauges[m]++
	li.atomic.Unlock()
}

// setGauge sets the parameter metric.
func (li *LocalInstrumentor) decGauge(m MetricGauge) {
	li.atomic.Lock()
	li.gauges[m]--
	li.atomic.Unlock()
}

// addHistogramEntry adds an entry to parameter metric.
func (li *LocalInstrumentor) addHistogramEntry(m MetricHistogram, value float64) {
	li.atomic.Lock()
	var hist map[float64]int64
	hist, ok := li.histograms[m]
	if !ok {
		hist = make(map[float64]int64)
		li.histograms[m] = hist
	}
	var match bool
	for _, v := range localBuckets {
		if value <= v {
			hist[v]++
			match = true
			break
		}
	}
	// If no quint match found, store in '0' map entry ('0' is stand-in for +Inf).
	if !match {
		hist[0]++
	}
	li.atomic.Unlock()
}

// MetricSnap contains a snapshot of telemetry data available when the snapshot was created.
type MetricSnap struct {
	Gauges     map[MetricGauge]int64
	Histograms map[MetricHistogram]map[float64]int64
}

// HistogramSummaries rolls up each histogram to a summary of the number of entries in each.
func (ms MetricSnap) HistogramSummaries() map[MetricHistogram]int64 {
	summary := make(map[MetricHistogram]int64)
	for hist, hmap := range ms.Histograms {
		if len(hmap) > 0 {
			for _, count := range hmap {
				summary[hist] += count
			}
		}
	}
	return summary
}

// SnapMetrics locks subject LocalInstrumentor, copies all metrics into, and returns, a MetricSnap.
func (li *LocalInstrumentor) SnapMetrics() MetricSnap {
	snap := MetricSnap{
		Gauges:     make(map[MetricGauge]int64),
		Histograms: make(map[MetricHistogram]map[float64]int64),
	}

	li.atomic.Lock()
	for mg, mval := range li.gauges {
		snap.Gauges[mg] = mval
	}
	for mh, hmap := range li.histograms {
		if len(hmap) > 0 {
			snap.Histograms[mh] = make(map[float64]int64)
			for mhkey, mhvalue := range hmap {
				snap.Histograms[mh][mhkey] = mhvalue
			}
		}
	}
	li.atomic.Unlock()
	return snap
}
