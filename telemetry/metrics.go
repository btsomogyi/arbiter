package telemetry

// Static strings used in metric definitions.
const (
	component              = "arbiter"
	QueueChanDepthDesc     = "QueueChanDepth"     // measure the main supervisor input channel (unprocessed messages)
	ProcessingMapDepthDesc = "ProcessingMapDepth" // measure number of concurrently processed requests
	WaitingMapDepthDesc    = "WaitingMapDepth"    // measures the number of requests currently waiting for in-flight reqs to finish
	MessageDesc            = "Message"            // time metrics for messages passed between Supervisor and Worker
	WorkDesc               = "Worktime"           // time to complete the work being arbitrated (passed in closure function)
	TransactionDesc        = "Transaction"        // complete transaction time including arbiter overhead
)

// Labels are used to signify dimensions of the stored metrics (states/results/statuses).
type Labels map[string]string

// MetricGauge is the enum used to index the Gauge map.
type MetricGauge int

// MetricHistogram is the enum used to index the Histogram map.
type MetricHistogram int

// MetricGauge index constants
const (
	QueueChanDepth     MetricGauge = iota // point in time number of entries in begin channel.
	ProcessingMapDepth                    // point in time number of entries in the processing map.
	WaitingMapDepth                       // point in time number of entries in the waiting map.
)

// MetricHistogram index constants
const (
	Messages     MetricHistogram = iota // time between send of Begin message and it being processed.
	Worktime                            // time between send of End message and it being processed.
	Transactions                        // time between send of begin of transaction and completion.
)

func (m MetricGauge) String() string {
	return [...]string{
		"QueueChanDepth",
		"ProcessingMapDepth",
		"WaitingMapDepth",
	}[m]
}

func (m MetricHistogram) String() string {
	return [...]string{
		"Messages",
		"Worktime",
		"Transaction",
	}[m]
}

// MetricGauges is the collection of Gauge metrics implemented by package.
var MetricGauges = map[MetricGauge]string{
	QueueChanDepth:     "Number of messages in queue channel",
	ProcessingMapDepth: "Number of active processing messages",
	WaitingMapDepth:    "Number of waiting messages",
}

// MetricHistograms is the collection of Histogram metrics implemented by package.
var MetricHistograms = map[MetricHistogram]string{
	Messages:     "Time before supervisor processes messages sent from worker",
	Worktime:     "Time it takes to complete the work being arbitrated (passed in closure function)",
	Transactions: "Total time between begin of transaction and completion",
}

// MetricHistogramLabels provides the label keys for Histogram Vectors in Prometheus.
var MetricHistogramLabels = map[MetricHistogram][]string{
	Messages: {
		"state",
		"signal",
		"waitlisted",
		"finalizefailed",
	},
	Worktime: {
		"signal",
	},
	Transactions: {
		"signal",
	},
}

// Instrumentor is the interface to be implemented by any concrete telemetry type.
type Instrumentor interface {
	QueueChanDepth(int64)
	IncQueueChanDepth()
	DecQueueChanDepth()
	ProcessingMapDepth(int64)
	IncProcessingMapDepth()
	DecProcessingMapDepth()
	WaitingMapDepth(int64)
	IncWaitingMapDepth()
	DecWaitingMapDepth()
	Messages(float64, ...Labels)
	Worktime(float64, ...Labels)
	Transactions(float64, ...Labels)
}
