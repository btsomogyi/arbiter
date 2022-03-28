package internal

import (
	"context"
	"github.com/btsomogyi/arbiter/interfaces"
	"time"

	"github.com/btsomogyi/arbiter/logging"
	"github.com/btsomogyi/arbiter/telemetry"
)

// Supervisor contains the primary channels used for synchronization between Worker and Supervisor.
type Supervisor struct {
	queue       chan message
	terminate   chan struct{}
	processing  *messageMap
	waiting     *messageMap
	metrics     telemetry.Instrumentor
	logger      logging.Logger
	pollDone    func()
	initialized bool
}

// config contains the adjustable configuraiton of the Supervisor.
type config struct {
	channelDepth uint
	Instrument   telemetry.Instrumentor
	pollDone     func()
	logger       logging.Logger
}

// configuration is the default configuration of the Supervisor.
var configuration = &config{
	channelDepth: channelDepth,
	pollDone:     func() {},
}

// A SupervisorOption is a function that modifies the behavior of a Supervisor.
type SupervisorOption func(*config) error

// SetInstrumentor sets the Instrumentor for the Supervisor.
func SetInstrumentor(i telemetry.Instrumentor) SupervisorOption {
	return func(c *config) error {
		c.Instrument = i
		return nil
	}
}

// SetChannelDepth sets the maximum size of the queue channel.
func SetChannelDepth(d uint) SupervisorOption {
	return func(c *config) error {
		c.channelDepth = d
		return nil
	}
}

// SetLogger provides a compatible structured logger for emitting log messages.
// If not provided, a no-op logger is created during supervisor initialization.
func SetLogger(l logging.Logger) SupervisorOption {
	return func(c *config) error {
		c.logger = l
		return nil
	}
}

// SetPollFunction sets the pollDone function in Supervisor for deterministic testing.
func SetPollFunction(f func()) SupervisorOption {
	return func(c *config) error {
		c.pollDone = f
		return nil
	}
}

// NewSupervisor returns an initialized arbiter server.
func NewSupervisor(opts ...SupervisorOption) (*Supervisor, error) {
	cfg := configuration
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}
	s := &Supervisor{}
	s.init(cfg)

	return s, nil
}

// init initizalizes internal structures and is invoked before processing begins.
func (s *Supervisor) init(c *config) {
	if s.initialized == false {
		s.processing = newMessageMap()
		s.waiting = newMessageMap()
		s.queue = make(chan message, c.channelDepth)
		s.terminate = make(chan struct{})
		s.metrics = c.Instrument
		s.pollDone = c.pollDone
		s.logger = c.logger
	}
	if s.logger == nil {
		// Create default silent logger if uninitialized.
		s.logger = logging.NewNoopLogger()
	}
	if s.metrics == nil {
		s.metrics = telemetry.NewNopInstrumentor()
	}
	s.metrics.QueueChanDepth(0)
	s.metrics.ProcessingMapDepth(0)
	s.metrics.WaitingMapDepth(0)
	s.initialized = true
}

// Terminate shutdowns the arbiter supervisor goroutine.
func (s *Supervisor) Terminate() {
	if s.initialized {
		close(s.terminate)
	}
}

// Process is primary logic loop of Arbiter supervisor goroutine. This should be invoked as soon as
// message processing should begin. Process will automatically call init() as required.
func (s *Supervisor) Process() {
	if !s.initialized {
		s.init(configuration)
	}

	// Receives messages from supervisor queue and dispatches them based on state (begin/end).
	// Terminates function when supervisor terminate channel is closed. Note messages must be passed
	// by pointer to processBegin/end to ensure identity is preserved.
	for {
		s.pollDone()
		select {
		case m := <-s.queue:
			s.metrics.DecQueueChanDepth()
			switch m.(type) {
			case *beginMessage:
				m.setLatency()
				s.processBegin(m)
				ms := m.getStatus()
				s.logger.Debug("Supervisor completed begin message processing", []logging.LogTuple{
					{"key", m.request().GetKey()},
					{"duration", m.getLatency()},
					{"state", beginState.String()},
					{"results", ms.results()},
					{"waitlist", ms.waitlist()},
					{"finalizefailure", ms.finalizefailure()},
				})
			case *endMessage:
				m.setLatency()
				s.processEnd(m)
				ms := m.getStatus()
				s.logger.Debug("Supervisor completed end message processing", []logging.LogTuple{
					{"key", m.request().GetKey()},
					{"duration", m.getLatency()},
					{"state", beginState.String()},
					{"results", ms.results()},
					{"waitlist", ms.waitlist()},
					{"finalizefailure", ms.finalizefailure()},
				})
			}
		case <-s.terminate:
			return
		}
		s.pollDone()
	}
}

// processBegin consumes the BeginMessage message and either stores message in waitingMap, responds
// with 'ceaseSignal' message to worker, or responds with 'proceedSignal' message to worker.
func (s *Supervisor) processBegin(m message) {
	_, ok := m.(*beginMessage)
	if !ok {
		// This is by design an unreachable condition, but left in to detect future package modifications that
		// may violate that design. Only messages with underlying `beginMessage` type are sent to processBegin().
		s.logger.DPanic("Non-beginMessage sent to processBegin()", []logging.LogTuple{
			{Field: "request key", Value: m.request().GetKey()},
		})
	}

	// Check if valid, and reject if not.
	if err := m.request().Valid(); err != nil {
		m.setStatus(msCease)
		s.pushMessageMetrics(m)
		m.respond(beginState, ceaseSignal, err)
		return
	}

	// Message valid, check for immediate handling or waitlisting.
	s.enqueMessage(m)
}

// enqueMessage checks the processing messageMap for a message with same Request Key and
// determines if the incoming message supersedes any found (ceased if doesn't supersede).
// For valid Messages with an active processing entry for that key, check the waiting messageMap
// to determine if it should be stored (replacing any existing inferior message), or ceaseSignal'd
// as superseded by the currently waiting message.
func (s *Supervisor) enqueMessage(m message) {
	reqKey := m.request().GetKey()

	// Check processing map.
	inProcessMsg, foundProcessing := s.processing.getMessage(reqKey)
	if !foundProcessing {
		// nothing found active, activate new message immediately.
		m.setStatus(msProceed)
		s.activateMessage(m)
		return
	}
	// If new message does not supersede in process message, then new message is redundant, and should
	// CEASE immediately.
	if err := m.request().Supersedes(inProcessMsg.request()); err != nil {
		m.setStatus(msCease)
		s.pushMessageMetrics(m)
		m.respond(beginState, ceaseSignal, err)
		return
	}

	// Check Waiting map, and if not found, add new message to waiting map (awaiting completeion
	// of current in-flight Processing map entry).
	waitingMsg, waitFound := s.waiting.getMessage(reqKey)
	if !waitFound {
		m.setStatus(msWaitlist)
		s.metrics.IncWaitingMapDepth()
		s.waiting.add(m)
		return
	}

	// If the new message does NOT supersede the waiting message, then it is redundant
	// and is Ceased, leaving the waiting message on the waitlist.
	if err := m.request().Supersedes(waitingMsg.request()); err != nil {
		m.setStatus(msCease)
		s.pushMessageMetrics(m)
		m.respond(beginState, ceaseSignal, err)
		return
	}

	// Otherwise, new message supersedes waiting message; waiting message is Ceased and
	// new message put in waiting list in its place.
	// TODO [BTS]: Enforce Supersedes as reciprical ( a.sup(b) || b.sup(a) == true).  Will
	// likely require Go2 generic constraints to programatically enforce.
	err := waitingMsg.request().Supersedes(m.request())

	s.metrics.DecWaitingMapDepth()
	waitingMsg.setStatus(msCease)
	s.pushMessageMetrics(waitingMsg)
	waitingMsg.respond(beginState, ceaseSignal, err)

	m.setStatus(msWaitlist)
	s.metrics.IncWaitingMapDepth()
	s.waiting.add(m)

	return
}

// activateMessage adds message to processing messageMap, notifies worker of message
// to proceedSignal, and increments counters.
func (s *Supervisor) activateMessage(m message) {
	s.metrics.IncProcessingMapDepth()
	s.processing.add(m)
	m.setStatus(msProceed)
	s.pushMessageMetrics(m)
	m.respond(beginState, proceedSignal, nil)
}

// processEnd consumes the EndMessage message and attempts to store successful results to Datastore.
// Either successSignal or failureSignal result in purging of map data for {id, version} tuple and promoting
// any waiting versions to processing map (with corresponding send of message).
func (s *Supervisor) processEnd(m message) {
	em, ok := m.(*endMessage)
	if !ok {
		// This is by design an unreachable condition, but left in to detect future package modifications that
		// may violate that design. Only messages with underlying `endMessage` type are sent to processEnd().
		s.logger.DPanic("Non-endMessage sent to processEnd()", []logging.LogTuple{
			{Field: "request key", Value: m.request().GetKey()},
		})
	}

	switch em.signal {
	case failureSignal:
		m.setStatus(msFailure)
		s.pushMessageMetrics(m)
		m.respond(endState, failureSignal, nil)
	case successSignal:
		m.setStatus(msSuccess)
		err := m.request().Finalize()
		if err != nil {
			m.setStatus(msFinalizeFailure)
			s.pushMessageMetrics(m)
			m.respond(endState, failureSignal, err)
		} else {
			s.pushMessageMetrics(m)
			m.respond(endState, successSignal, nil)
		}
	default:
		// This is by design an unreachable condition, but left in to detect future modifications that
		// may violate that design.  The worker code has no means of setting an invalid value.
		s.logger.DPanic("Unexpected status sent in end message", []logging.LogTuple{
			{Field: "request key", Value: m.request().GetKey()},
			{Field: "message signal", Value: em.signal},
		})
	}

	s.purgeMessage(m)
}

// purgeMessage checks waiting and processing messageMaps for message with same
// Request key and signals any Messages that are promoted from waiting to processing.
func (s *Supervisor) purgeMessage(m message) {
	reqKey := m.request().GetKey()

	// Check waiting map for exact message, if found, remove and return.
	if foundWaiting := s.waiting.containsMessage(m); foundWaiting {
		s.metrics.DecWaitingMapDepth()
		s.waiting.remove(m)
		return
	}

	// Check processing map for exact message, if found, remove.
	if foundProcessing := s.processing.containsMessage(m); foundProcessing {
		s.metrics.DecProcessingMapDepth()
		s.processing.remove(m)
		s.promoteFromWaiting(reqKey)
	}

}

func (s *Supervisor) promoteFromWaiting(reqKey int64) {
	// Check waiting map.
	if waitingMsg, foundWaiting := s.waiting.getMessage(reqKey); foundWaiting {
		// Remove from waiting messageMap and activate.
		s.metrics.DecWaitingMapDepth()
		s.waiting.remove(waitingMsg)
		s.activateMessage(waitingMsg)
	}
}

func (s *Supervisor) pushMessageMetrics(m message) {
	var state string
	switch m.(type) {
	case *beginMessage:
		state = beginState.String()
	case *endMessage:
		state = endState.String()
	}
	ms := m.getStatus()
	s.metrics.Messages(m.getLatency(), telemetry.Labels{
		"state":          state,
		"signal":         ms.results(),
		"waitlisted":     ms.waitlist(),
		"finalizefailed": ms.finalizefailure(),
	})
}

// GenerateWorker initializes a new Worker and returns to caller. One worker instance
// should be used for each request submitted to the Supervisor. Note that worker
// signature is generated from this structure, which ensures that even if Worker is
// passed by value by the consumer which generated the Worker, all messages will
// continue to have this same unique signature value, disambiguating messages
// originating from this Worker from messages originating from other Workers.
func (s *Supervisor) generateWorker(ctx context.Context, r interfaces.Request) (*worker, func()) {
	w := worker{
		queue:    s.queue,
		metrics:  s.metrics,
		status:   failureSignal,
		response: make(chan response, channelDepth),
		done:     make(chan struct{}),
		request:  r,
		ctx:      ctx,
	}

	w.signature = &w
	return &w, (&w).deferredFunc
}

// WithWorker creates a closure to invoke an Arbiter Worker and handles all
// communication with the Arbiter Supervisor.  This allows all machinery of
// interaction between worker and supervisor to be predetermined and private
// to Arbiter package.
func (s *Supervisor) WithWorker(ctx context.Context, r interfaces.Request, fn func(context.Context) error) error {
	w, df := s.generateWorker(ctx, r)
	defer df()

	s.logger.Debug("WithWorker function entered", []logging.LogTuple{
		{"key", w.request.GetKey()},
		{"status", w.status.String()},
	})

	w.sendBegin()
	beginResponse := w.recvResponse(beginState, ceaseSignal)

	s.logger.Debug("WithWorker received beginResponse", []logging.LogTuple{
		{"key", w.request.GetKey()},
		{"status", w.status.String()},
		{"response", beginResponse.sig.String()},
	})

	if beginResponse.sig != proceedSignal {
		duration := w.duration()
		s.metrics.Transactions(duration, telemetry.Labels{
			"signal": failureSignal.String(),
		})
		s.logger.Debug("WithWorker transaction completed with error", []logging.LogTuple{
			{"key", w.request.GetKey()},
			{"duration", duration},
			{"response", beginResponse.sig.String()},
		})
		return beginResponse.err
	}

	w.workStart = time.Now()
	if err := fn(ctx); err != nil {
		workDuration := w.workDuration()
		duration := w.duration()
		s.metrics.Worktime(workDuration, telemetry.Labels{
			"signal": successSignal.String(),
		})
		s.metrics.Transactions(duration, telemetry.Labels{
			"signal": failureSignal.String(),
		})
		s.logger.Debug("WithWorker transaction completed with closure error", []logging.LogTuple{
			{"key", w.request.GetKey()},
			{"duration", duration},
			{"worktime", workDuration},
			{"response", beginResponse.sig.String()},
		})
		return err
	}

	w.status = successSignal
	workDuration := w.workDuration()
	s.metrics.Worktime(workDuration, telemetry.Labels{
		"signal": successSignal.String(),
	})

	s.logger.Debug("WithWorker completed provided work function", []logging.LogTuple{
		{"key", w.request.GetKey()},
		{"status", w.status.String()},
	})

	w.sendEnd()
	endResponse := w.recvResponse(endState, failureSignal)

	if endResponse.sig != successSignal {
		duration := w.duration()
		s.metrics.Transactions(duration, telemetry.Labels{
			"signal": failureSignal.String(),
		})
		s.logger.Debug("WithWorker transaction completed with error", []logging.LogTuple{
			{"key", w.request.GetKey()},
			{"duration", duration},
			{"worktime", workDuration},
			{"response", endResponse.sig.String()},
		})
		return endResponse.err
	}

	s.logger.Debug("WithWorker received endResponse", []logging.LogTuple{
		{"key", w.request.GetKey()},
		{"status", w.status.String()},
		{"response", endResponse.sig.String()},
	})

	duration := w.duration()
	s.metrics.Transactions(w.duration(), telemetry.Labels{
		"signal": successSignal.String(),
	})
	s.logger.Debug("WithWorker transaction completed", []logging.LogTuple{
		{"key", w.request.GetKey()},
		{"duration", duration},
		{"worktime", workDuration},
	})
	return nil
}
