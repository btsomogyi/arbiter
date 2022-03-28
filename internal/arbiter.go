package internal

import (
	"fmt"
	"github.com/btsomogyi/arbiter/interfaces"
	"time"
)

// channelDepth is default channel depth for all Arbiter supervisor/worker channels.
const channelDepth = 10

// state indicates whether a message is requesting to begin or is at an end.
type state int

// beginState/endState signify the phase of communication between the Worker and Supervisor.
const (
	nullState  state = iota // Null indicates uninitialized value.
	beginState              // begin indicates worker is requesting to begin processing.
	endState                // end indicates worker has completed processing.
)

func (s state) String() string {
	return [...]string{"nullState", "beginState", "endState"}[s]
}

// signal carries response messages back to worker from arbiter supervisor goroutine.
type signal int

// Set of response 'signal' values.  Null assists in debugging (providing clear indication
// when a 'signal' has not been explicitly set).  proceedSignal/ceaseSignal are the action signals
// provided to Workers in response to begin Messages.  successSignal/failureSignal are the status
// signals provided to Supervisor from workers regarding requests, or from Supervisor to Workers
// regarding status of the Finalize action taken by Supervisor on behalf of Workers.
const (
	nullSignal    signal = iota // Null indicates uninitialized value.
	proceedSignal               // proceedSignal indicates arbiter worker should proceed with processing.
	ceaseSignal                 // ceaseSignal indicates arbiter worker should abort processing.
	successSignal               // successSignal indicates arbiter worker completed processing successfully.
	failureSignal               // failureSignal indicates arbiter worker failed to complete processing.
)

func (s signal) String() string {
	return [...]string{"nullSignal", "proceedSignal", "ceaseSignal", "successSignal", "failureSignal"}[s]
}

// response is the message sent back to worker from Arbiter supervisor to worker. state indicates
// whether response to begin or end message, and signal and Error provide response to worker.
type response struct {
	state state
	sig   signal
	err   error
}

// String is stringer for response members.
func (r *response) String() string {
	return fmt.Sprintf("state: %s signal: %s error: %e", r.state, r.sig, r.err)
}

// responseFunc is a function passed to arbiter supervisor goroutine to faciliate
// panic-free responses to arbiter worker over an enclosed channel. state indicates
// begin/end type, and signal provides indication to worker about request status
// (proceedSignal/ceaseSignal for beginState responses, successSignal/failureSignal
// for end responses).
type responseFunc func(state, signal, error)

type message interface {
	request() interfaces.Request
	respond(state, signal, error)
	signature() *worker
	same(message) bool
	setLatency()
	getLatency() float64
	setStatus(messageStatus)
	unsetStatus(messageStatus)
	getStatus() messageStatus
}

var _ message = (*beginMessage)(nil)

type beginMessage struct {
	req          interfaces.Request
	responseFunc responseFunc
	timestamp    time.Time
	latency      float64
	workerSig    *worker
	status       messageStatus
}

func (m *beginMessage) request() interfaces.Request {
	return m.req
}

func (m *beginMessage) respond(state state, signal signal, err error) {
	m.responseFunc(state, signal, err)
}

func (m *beginMessage) signature() *worker {
	return m.workerSig
}

func (m *beginMessage) setLatency() {
	m.latency = timeElapsedInSeconds(m.timestamp)
}

func (m *beginMessage) getLatency() float64 {
	return m.latency
}

func (m *beginMessage) setStatus(ms messageStatus) {
	m.status.addStatus(ms)
}

func (m *beginMessage) unsetStatus(ms messageStatus) {
	m.status.removeStatus(ms)
}

func (m *beginMessage) getStatus() messageStatus {
	return m.status
}

// same identifies one message 'm' as having the same identity with an'other' message 'o',
// having identical requests and being from the same worker (identical worker signatures).
func (m *beginMessage) same(o message) bool {
	return m.signature() == o.signature() && m.request() == o.request()
}

var _ message = (*endMessage)(nil)

type endMessage struct {
	req          interfaces.Request
	responseFunc responseFunc
	signal       signal
	timestamp    time.Time
	latency      float64
	workerSig    *worker
	status       messageStatus
}

func (m *endMessage) request() interfaces.Request {
	return m.req
}

func (m *endMessage) respond(state state, signal signal, err error) {
	m.responseFunc(state, signal, err)
}

func (m *endMessage) signature() *worker {
	return m.workerSig
}

func (m *endMessage) setLatency() {
	m.latency = timeElapsedInSeconds(m.timestamp)
}

func (m *endMessage) getLatency() float64 {
	return m.latency
}

func (m *endMessage) setStatus(ms messageStatus) {
	m.status.addStatus(ms)
}

func (m *endMessage) unsetStatus(ms messageStatus) {
	m.status.removeStatus(ms)
}

func (m *endMessage) getStatus() messageStatus {
	return m.status
}

// same identifies one message 'm' as having the same identity with an'other' message 'o',
// having identical requests and being from the same worker (identical worker signatures).
func (m *endMessage) same(o message) bool {
	return m.signature() == o.signature() && m.request() == o.request()
}

// messageStatus captures the status of a message for reporting within metrics.
type messageStatus int

// Set of values relevant to message processing states.
const (
	msProceed         messageStatus = 1 << iota // 1 << 0 which is 00000001
	msSuccess                                   // 1 << 1 which is 00000010
	msCease                                     // 1 << 2 which is 00000100
	msFailure                                   // 1 << 3 which is 00001000
	msWaitlist                                  // 1 << 4 which is 00010000
	msFinalizeFailure                           // 1 << 5 which is 00100000
)

// addStatus idempotently adds the passed status bits to the message status.
// All status is cumulative (multiple status values may be added to a message status).
func (m *messageStatus) addStatus(s messageStatus) {
	if *m&s == 0 {
		*m += s
	}
}

// removeStatus idempotently removes the passed status bits to the message status.
func (m *messageStatus) removeStatus(s messageStatus) {
	if *m&s != 0 {
		*m -= s
	}
}

// results checks messageStatus bits and returns one value for status label.
func (m messageStatus) results() string {
	if m&msProceed != 0 {
		return "proceed"
	}
	if m&msSuccess != 0 {
		return "success"
	}
	if m&msCease != 0 {
		return "cease"
	}
	if m&msFailure != 0 {
		return "failure"
	}
	return ""
}

func (m messageStatus) waitlist() string {
	if m&msWaitlist != 0 {
		return "true"
	}
	return "false"
}

func (m messageStatus) finalizefailure() string {
	if m&msFinalizeFailure != 0 {
		return "true"
	}
	return "false"
}

func timeElapsedInSeconds(start time.Time) float64 {
	return float64(time.Since(start).Nanoseconds()) / float64(time.Nanosecond)
}
