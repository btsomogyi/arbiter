package telemetry

import (
	//"fmt"
	//"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const subsystemSpecifier = "arbiter"

var promBuckets = prometheus.ExponentialBuckets(0.00002, 2, 20)

// PromInstrumentor implements the arbiter Instrumentor interface using prometheus metrics package.
type PromInstrumentor struct {
	gaugeMetrics     map[MetricGauge]prometheus.Gauge
	histogramMetrics map[MetricHistogram]*prometheus.HistogramVec
}

// PromInstrumenter implements the Instrumentor interface using prometheus metrics.

var _ Instrumentor = (*PromInstrumentor)(nil)

// InstrumentPromMetrics sets up arbiter metrics in the prometheus metrics node.
func InstrumentPromMetrics() *PromInstrumentor {
	pi := PromInstrumentor{
		gaugeMetrics:     make(map[MetricGauge]prometheus.Gauge),
		histogramMetrics: make(map[MetricHistogram]*prometheus.HistogramVec),
	}
	for k, v := range MetricGauges {
		pi.gaugeMetrics[k] = promauto.NewGauge(prometheus.GaugeOpts{
			Subsystem: subsystemSpecifier,
			Name:      k.String(),
			Help:      v,
		})
	}

	for k, v := range MetricHistograms {
		pi.histogramMetrics[k] = promauto.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: subsystemSpecifier,
			Name:      k.String(),
			Help:      v,
			Buckets:   promBuckets,
		}, MetricHistogramLabels[k])
	}

	return &pi
}

func (pi *PromInstrumentor) QueueChanDepth(value int64) {
	pi.gaugeMetrics[QueueChanDepth].Set(float64(value))
}

func (pi *PromInstrumentor) IncQueueChanDepth() {
	pi.gaugeMetrics[QueueChanDepth].Inc()
}

func (pi *PromInstrumentor) DecQueueChanDepth() {
	pi.gaugeMetrics[QueueChanDepth].Dec()
}

func (pi *PromInstrumentor) ProcessingMapDepth(value int64) {
	pi.gaugeMetrics[ProcessingMapDepth].Set(float64(value))
}

func (pi *PromInstrumentor) IncProcessingMapDepth() {
	pi.gaugeMetrics[ProcessingMapDepth].Inc()
}

func (pi *PromInstrumentor) DecProcessingMapDepth() {
	pi.gaugeMetrics[ProcessingMapDepth].Dec()
}

func (pi *PromInstrumentor) WaitingMapDepth(value int64) {
	pi.gaugeMetrics[WaitingMapDepth].Set(float64(value))
}

func (pi *PromInstrumentor) IncWaitingMapDepth() {
	pi.gaugeMetrics[WaitingMapDepth].Inc()
}

func (pi *PromInstrumentor) DecWaitingMapDepth() {
	pi.gaugeMetrics[WaitingMapDepth].Dec()
}

func (pi *PromInstrumentor) Messages(value float64, labels ...Labels) {
	pi.histogramMetrics[Messages].With(prometheus.Labels(aggLabels(labels...))).Observe(value)
}

func (pi *PromInstrumentor) Worktime(value float64, labels ...Labels) {
	pi.histogramMetrics[Worktime].With(prometheus.Labels(aggLabels(labels...))).Observe(value)
}

func (pi *PromInstrumentor) Transactions(value float64, labels ...Labels) {
	pi.histogramMetrics[Transactions].With(prometheus.Labels(aggLabels(labels...))).Observe(value)
}

func aggLabels(labels ...Labels) Labels {
	var agg Labels
	for _, l := range labels {
		for k, v := range l {
			agg[k] = v
		}
	}
	return agg
}
