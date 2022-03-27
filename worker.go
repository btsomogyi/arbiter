package arbiter

import (
	"context"
	"time"

	"github.com/btsomogyi/arbiter/telemetry"
)

// worker contains the supervisor reference and status values used to properly close channels.
type worker struct {
	queue     chan<- message
	metrics   telemetry.Instrumentor
	response  chan response
	done      chan struct{}
	ctx       context.Context
	status    signal
	beginSent time.Time
	endSent   time.Time
	workStart time.Time
	request   Request
	signature *worker
}

func (w *worker) deferredFunc() {
	close(w.done)
	// Only send end if one has not already been sent.
	if w.endSent.IsZero() {
		// worker sends whatever status is stored in worker "status" field.
		w.sendEnd()
	}
}

func (w *worker) responseToWorkerFunc(t state, s signal, e error) {
	response := response{
		state: t,
		sig:   s,
		err:   e,
	}
	select {
	case <-w.done:
		// client has ceased, abort any attempt to return messages.
	default:
		w.response <- response
	}
}

func (w *worker) sendBegin() {
	msg := beginMessage{
		req:          w.request,
		responseFunc: w.responseToWorkerFunc,
		timestamp:    time.Now(),
		workerSig:    w.signature,
	}
	w.beginSent = time.Now()
	w.metrics.IncQueueChanDepth()
	w.queue <- &msg
}

// sendEnd creates an EndMessage from the signal embedded in worker and sends it to supervisor.
func (w *worker) sendEnd() {
	msg := endMessage{
		req:          w.request,
		signal:       w.status,
		responseFunc: w.responseToWorkerFunc,
		timestamp:    time.Now(),
		workerSig:    w.signature,
	}
	w.queue <- &msg
	w.metrics.IncQueueChanDepth()
	w.endSent = time.Now()
}

// recvResponse blocks returning the response from the Supervisor to the worker.
// If the context is canceled prior to Supervisor response, respond to worker with
func (w *worker) recvResponse(state state, defaultSig signal) response {
	select {
	case resp := <-w.response:
		return resp
	case <-w.ctx.Done():
		return response{
			state: state,
			sig:   defaultSig,
			err:   w.ctx.Err(),
		}
	}
}

func (w *worker) duration() float64 {
	return timeElapsedInSeconds(w.beginSent)
}

func (w *worker) workDuration() float64 {
	return timeElapsedInSeconds(w.workStart)
}
