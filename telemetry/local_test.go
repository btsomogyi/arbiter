package telemetry

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_SnapMetric(t *testing.T) {
	tests := map[string]struct {
		li        *LocalInstrumentor
		want      MetricSnap
		summaries map[MetricHistogram]int64
	}{
		"various metrics": {
			li: &LocalInstrumentor{
				gauges: map[MetricGauge]int64{
					QueueChanDepth:     10,
					ProcessingMapDepth: 102,
					WaitingMapDepth:    1002,
				},
				histograms: map[MetricHistogram]map[float64]int64{
					Transactions: {
						0:       1,
						1.31072: 1,
					},
					Messages: {
						5.24288: 2,
						0.16384: 2,
					},
				},
			},
			want: MetricSnap{
				Gauges: map[MetricGauge]int64{
					QueueChanDepth:     10,
					ProcessingMapDepth: 102,
					WaitingMapDepth:    1002,
				},
				Histograms: map[MetricHistogram]map[float64]int64{
					Transactions: {
						0:       1,
						1.31072: 1,
					},
					Messages: {
						5.24288: 2,
						0.16384: 2,
					},
				},
			},
			summaries: map[MetricHistogram]int64{
				Transactions: 2,
				Messages:     4,
			},
		},
		"empty metrics": {
			li: &LocalInstrumentor{
				gauges: map[MetricGauge]int64{
					QueueChanDepth:     0,
					ProcessingMapDepth: 0,
					WaitingMapDepth:    0,
				},
				histograms: map[MetricHistogram]map[float64]int64{},
			},
			want: MetricSnap{
				Gauges: map[MetricGauge]int64{
					QueueChanDepth:     0,
					ProcessingMapDepth: 0,
					WaitingMapDepth:    0,
				},
				Histograms: map[MetricHistogram]map[float64]int64{},
			},
			summaries: map[MetricHistogram]int64{},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			got := tc.li.SnapMetrics()
			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Errorf(diff)
			}
			summary := got.HistogramSummaries()
			for hist, hmap := range got.Histograms {
				for k, v := range hmap {
					t.Logf("hist %s key: %f value: %d\n", hist.String(), k, v)
				}

			}
			for k, v := range summary {
				t.Logf("summary key: %s value: %d\n", k.String(), v)
			}

			sumdiff := cmp.Diff(tc.summaries, summary)
			//t.Logf(sumdiff)
			if sumdiff != "" {
				t.Errorf(sumdiff)
			}
		})
	}
}

var emptyHistograms = map[MetricHistogram]map[float64]int64{}

func Test_LocalInstrumentor(t *testing.T) {
	//type metricOp interface{}
	tests := map[string]struct {
		metricOps func(*LocalInstrumentor) *LocalInstrumentor
		want      MetricSnap
	}{
		"set single gauge": {
			metricOps: func(li *LocalInstrumentor) *LocalInstrumentor {
				li.QueueChanDepth(10)
				return li
			},
			want: MetricSnap{
				Gauges: map[MetricGauge]int64{
					QueueChanDepth:     10,
					ProcessingMapDepth: 0,
					WaitingMapDepth:    0,
				},
				Histograms: emptyHistograms,
			},
		},
		"reset single gauge": {
			metricOps: func(li *LocalInstrumentor) *LocalInstrumentor {
				li.QueueChanDepth(10)
				li.QueueChanDepth(5)
				return li
			},
			want: MetricSnap{
				Gauges: map[MetricGauge]int64{
					QueueChanDepth:     5,
					ProcessingMapDepth: 0,
					WaitingMapDepth:    0,
				},
				Histograms: emptyHistograms,
			},
		},
		"increment single gauge": {
			metricOps: func(li *LocalInstrumentor) *LocalInstrumentor {
				li.QueueChanDepth(10)
				li.IncQueueChanDepth()
				return li
			},
			want: MetricSnap{
				Gauges: map[MetricGauge]int64{
					QueueChanDepth:     11,
					ProcessingMapDepth: 0,
					WaitingMapDepth:    0,
				},
				Histograms: emptyHistograms,
			},
		},
		"decrement single gauge": {
			metricOps: func(li *LocalInstrumentor) *LocalInstrumentor {
				li.QueueChanDepth(10)
				li.DecQueueChanDepth()
				return li
			},
			want: MetricSnap{
				Gauges: map[MetricGauge]int64{
					QueueChanDepth:     9,
					ProcessingMapDepth: 0,
					WaitingMapDepth:    0,
				},
				Histograms: emptyHistograms,
			},
		},
		"vary multiple gauges": {
			metricOps: func(li *LocalInstrumentor) *LocalInstrumentor {
				li.QueueChanDepth(10)
				li.ProcessingMapDepth(100)
				li.WaitingMapDepth(1000)
				li.DecQueueChanDepth()
				li.IncProcessingMapDepth()
				li.DecWaitingMapDepth()
				li.DecQueueChanDepth()
				li.IncProcessingMapDepth()
				li.DecWaitingMapDepth()
				return li
			},
			want: MetricSnap{
				Gauges: map[MetricGauge]int64{
					QueueChanDepth:     8,
					ProcessingMapDepth: 102,
					WaitingMapDepth:    998,
				},
				Histograms: emptyHistograms,
			},
		},
		"reset multiple gauges": {
			metricOps: func(li *LocalInstrumentor) *LocalInstrumentor {
				li.QueueChanDepth(10)
				li.ProcessingMapDepth(100)
				li.WaitingMapDepth(1000)
				li.QueueChanDepth(50)
				li.ProcessingMapDepth(51)
				li.WaitingMapDepth(52)
				return li
			},
			want: MetricSnap{
				Gauges: map[MetricGauge]int64{
					QueueChanDepth:     50,
					WaitingMapDepth:    52,
					ProcessingMapDepth: 51,
				},
				Histograms: emptyHistograms,
			},
		},
		"single histogram": {
			metricOps: func(li *LocalInstrumentor) *LocalInstrumentor {
				li.Transactions(5.1)
				return li
			},
			want: MetricSnap{
				Gauges: map[MetricGauge]int64{
					QueueChanDepth:     0,
					WaitingMapDepth:    0,
					ProcessingMapDepth: 0,
				},
				Histograms: map[MetricHistogram]map[float64]int64{
					Transactions: {
						5.24288: 1,
					},
				},
			},
		},
		"single histogram infinity": {
			metricOps: func(li *LocalInstrumentor) *LocalInstrumentor {
				li.Transactions(10)
				return li
			},
			want: MetricSnap{
				Gauges: map[MetricGauge]int64{
					QueueChanDepth:     0,
					WaitingMapDepth:    0,
					ProcessingMapDepth: 0,
				},
				Histograms: map[MetricHistogram]map[float64]int64{
					Transactions: {
						0: 1,
					},
				},
			},
		},
		"single histogram multiple entries": {
			metricOps: func(li *LocalInstrumentor) *LocalInstrumentor {
				li.Transactions(10)
				li.Transactions(5)
				li.Transactions(5)
				li.Transactions(1)
				li.Transactions(0.10)
				return li
			},
			want: MetricSnap{
				Gauges: map[MetricGauge]int64{
					QueueChanDepth:     0,
					WaitingMapDepth:    0,
					ProcessingMapDepth: 0,
				},
				Histograms: map[MetricHistogram]map[float64]int64{
					Transactions: {
						0:       1,
						5.24288: 2,
						1.31072: 1,
						0.16384: 1,
					},
				},
			},
		},
		"multiple histograms": {
			metricOps: func(li *LocalInstrumentor) *LocalInstrumentor {
				li.Transactions(10)
				li.Messages(5)
				li.Messages(5)
				li.Transactions(1)
				li.Messages(0.10)
				li.Messages(0.10)
				return li
			},
			want: MetricSnap{
				Gauges: map[MetricGauge]int64{
					QueueChanDepth:     0,
					WaitingMapDepth:    0,
					ProcessingMapDepth: 0,
				},
				Histograms: map[MetricHistogram]map[float64]int64{
					Transactions: {
						0:       1,
						1.31072: 1,
					},
					Messages: {
						5.24288: 2,
						0.16384: 2,
					},
				},
			},
		},
		"multiple gauges and histograms": {
			metricOps: func(li *LocalInstrumentor) *LocalInstrumentor {
				li.QueueChanDepth(10)
				li.ProcessingMapDepth(100)
				li.WaitingMapDepth(1000)
				li.Transactions(10)
				li.Messages(5)
				li.Messages(5)
				li.Transactions(1)
				li.Messages(0.10)
				li.Messages(0.10)
				li.DecQueueChanDepth()
				li.IncProcessingMapDepth()
				li.IncWaitingMapDepth()
				li.DecQueueChanDepth()
				li.IncProcessingMapDepth()
				li.IncWaitingMapDepth()
				li.IncQueueChanDepth()
				li.IncQueueChanDepth()
				return li
			},
			want: MetricSnap{
				Gauges: map[MetricGauge]int64{
					QueueChanDepth:     10,
					ProcessingMapDepth: 102,
					WaitingMapDepth:    1002,
				},
				Histograms: map[MetricHistogram]map[float64]int64{
					Transactions: {
						0:       1,
						1.31072: 1,
					},
					Messages: {
						5.24288: 2,
						0.16384: 2,
					},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			li := NewLocalInstrumentor()

			tc.metricOps(li)

			got := li.SnapMetrics()
			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
