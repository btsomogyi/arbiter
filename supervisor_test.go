package arbiter_test

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap/zapcore"
	"math/rand"
	"sync"
	"testing"

	a "github.com/btsomogyi/arbiter"
	at "github.com/btsomogyi/arbiter/telemetry"
	"github.com/btsomogyi/arbiter/logging"

	"github.com/davecgh/go-spew/spew"
)

const (
	beginRequest        int = iota // initiate a request message
	waitRequest                    // wait for the given request to reach indicated state
	cancelRequest                  // cancel the context for a given request
	processRequests                // ratchet through supervisor request processing loop (once)
	terminateSupervisor            // cleanly exit supervisor goroutine, waiting for exit
)

func ExampleSupervisor() {
	// TODO
}

// event structure specifies the specific events used to sequence a test. A sequence
// of events defines the exact order and attributes of each action that is allowed to
// interact with the Supervisor under test.
type event struct {
	action      int    // taken from above (required)
	req         string // named request event pertains to (required)
	expectWork  bool   // whether to expect the request to execute its closure function (default false)
	expectError error  // expected error from request (if any)
	// both startWait and finishWait can be used both to wait for worker processing to get to wait point, as well
	// as *ensure* worker processing arrives at wait point before other events proceedSignal (dual syncronization
	//	between supervisor and worker).
	startWait  bool // whether request should pause before executing closure function (optional)
	finishWait bool // whether request should pause after executing closure function (optional)
	cancelAll  bool // cancel the context of all in-flight requests
}

// histogramSummaries used to store the expected message counts for test scenarios.
type histogramSummaries map[at.MetricHistogram]int64

// processEvent is a convenience alias due to the frequency of this event's use in testing.
var processEvent = event{action: processRequests}

// requestDefs predefine requests that are reused in various combinations during tests,
// for expedience and code reduction.
var requestDefs = map[string]testReq{
	"record2version10":     {key: 2, value: 10},
	"record1version20":     {key: 1, value: 20},
	"record3version30":     {key: 3, value: 30},
	"record1version8":      {key: 1, value: 8},
	"record1version9":      {key: 1, value: 9},
	"record1version10":     {key: 1, value: 10},
	"record1version11":     {key: 1, value: 11},
	"record1version9other": {key: 1, value: 9},
}

// Set to true to emit debug messages from tests and logging from arbiter (supervisor and worker).
var debug = true

// Test end to end operation of Supervisor
func Test_Supervisor(t *testing.T) {
	tests := map[string]struct {
		events      []event
		wantMetrics at.MetricSnap
		wantMsgs    histogramSummaries
		wantDB      []string
	}{
		"single request success": {
			events: []event{
				{
					action:     beginRequest,
					req:        "record2version10",
					expectWork: true,
					finishWait: true,
				},
				processEvent, // begin message start
				{
					action:     waitRequest,
					req:        "record2version10",
					finishWait: true,
				},
				processEvent, // end message start
				{
					action: terminateSupervisor,
				},
			},
			wantMetrics: at.MetricSnap{
				Gauges: map[at.MetricGauge]int64{
					at.QueueChanDepth:     0,
					at.ProcessingMapDepth: 0,
					at.WaitingMapDepth:    0,
				},
			},
			wantMsgs: histogramSummaries{
				at.Messages:     2,
				at.Transactions: 1,
			},
			wantDB: []string{
				"record2version10",
			},
		},
		"single request success (wait usage example)": {
			events: []event{
				{
					action:     beginRequest,
					req:        "record2version10",
					expectWork: true,
					startWait:  true,
					finishWait: true,
				},
				processEvent, // begin message start
				{
					action:    waitRequest,
					req:       "record2version10",
					startWait: true,
				},
				{
					action:     waitRequest,
					req:        "record2version10",
					finishWait: true,
				},
				processEvent, // end message start
				{
					action: terminateSupervisor,
				},
			},
			wantMetrics: at.MetricSnap{
				Gauges: map[at.MetricGauge]int64{
					at.QueueChanDepth:     0,
					at.ProcessingMapDepth: 0,
					at.WaitingMapDepth:    0,
				},
			},
			wantMsgs: histogramSummaries{
				at.Messages:     2,
				at.Transactions: 1,
			},
			wantDB: []string{
				"record2version10",
			},
		},
		"single request cancelled before begin": {
			events: []event{
				{
					action:      beginRequest,
					req:         "record2version10",
					expectWork:  false,
					expectError: context.Canceled,
				},
				{
					action: cancelRequest,
					req:    "record2version10",
				},
				processEvent, // begin message start
				processEvent, // end message start
				{
					action: terminateSupervisor,
				},
			},
			wantMetrics: at.MetricSnap{
				Gauges: map[at.MetricGauge]int64{
					at.QueueChanDepth:     0,
					at.ProcessingMapDepth: 0,
					at.WaitingMapDepth:    0,
				},
			},
			wantMsgs: histogramSummaries{
				at.Messages:     2,
				at.Transactions: 1,
			},
			wantDB: []string{},
		},
		"single request canceled after begin before work": {
			events: []event{
				{
					action:      beginRequest,
					req:         "record2version10",
					expectWork:  false,
					expectError: context.Canceled,
					startWait:   true,
				},
				processEvent, // begin message start
				{
					action: cancelRequest,
					req:    "record2version10",
				},
				{
					action:    waitRequest,
					req:       "record2version10",
					startWait: true,
				},
				processEvent, // end message start
				{
					action: terminateSupervisor,
				},
			},
			wantMetrics: at.MetricSnap{
				Gauges: map[at.MetricGauge]int64{
					at.QueueChanDepth:     0,
					at.ProcessingMapDepth: 0,
					at.WaitingMapDepth:    0,
				},
			},
			wantMsgs: histogramSummaries{
				at.Messages:     2,
				at.Transactions: 1,
			},
			wantDB: []string{},
		},
		"two requests one redundant": {
			// A valid request is sent, followed by an inferior version (redundant). Redundant version is discarded.
			// c1v20begin -> queue
			// queue -> c1v20begin -> "proceedSignal" response
			// c1v9begin -> queue
			// queue -> c1v9begin -> "ceaseSignal" response (because c1v20 already processing)
			// c1v20end -> queue
			// c1v9end -> queue
			// queue -> c1v20end -> "successSignal" response
			// queue -> c1v9end -> "failureSignal" response
			events: []event{
				{
					action:     beginRequest,
					req:        "record1version20",
					expectWork: true,
					startWait:  true,
				},
				processEvent, // record1version20 begin msg
				{
					action:      beginRequest,
					req:         "record1version9",
					expectWork:  false,
					expectError: ErrSupersededRequest,
				},
				processEvent, // record1version9 begin msg
				{
					action:    waitRequest,
					req:       "record1version20",
					startWait: true,
				},
				processEvent, // record1version20 end msg
				processEvent, // record1version9 end msg
				{
					action: terminateSupervisor,
				},
			},
			wantMetrics: at.MetricSnap{
				Gauges: map[at.MetricGauge]int64{
					at.QueueChanDepth:     0,
					at.ProcessingMapDepth: 0,
					at.WaitingMapDepth:    0,
				},
			},
			wantMsgs: histogramSummaries{
				at.Messages:     4,
				at.Transactions: 2,
			},
			wantDB: []string{
				"record1version20",
			},
		},
		"two requests one invalid": {
			// A valid request is sent and completed, followed by an inferior version (invalid). Invalid version is discarded.
			// c1v20begin -> queue -> c1v20begin -> "proceedSignal" response
			// c1v20end -> queue -> c1v20end -> "successSignal" response
			// c1v9begin -> queue -> c1v9begin -> "ceaseSignal" response (because c1v20 already in store)
			// c1v9end -> queue -> c1v9end -> "failureSignal" response (invalid message error)
			events: []event{
				{
					action:     beginRequest,
					req:        "record1version20",
					expectWork: true,
					finishWait: true,
				},
				processEvent, // record1version20 begin msg
				{
					action:     waitRequest,
					req:        "record1version20",
					finishWait: true,
				},
				processEvent, // record1version20 end msg
				{
					action:      beginRequest,
					req:         "record1version9",
					expectWork:  false,
					expectError: ErrInvalidRequest,
				},
				processEvent, // record1version9 begin msg
				processEvent, // record1version9 end msg
				{
					action: terminateSupervisor,
				},
			},
			wantMetrics: at.MetricSnap{
				Gauges: map[at.MetricGauge]int64{
					at.QueueChanDepth:     0,
					at.ProcessingMapDepth: 0,
					at.WaitingMapDepth:    0,
				},
			},
			wantMsgs: histogramSummaries{
				at.Messages:     4,
				at.Transactions: 2,
			},
			wantDB: []string{
				"record1version20",
			},
		},
		"two requests one redundant (duplicate message)": {
			// First message begins, and second message arrives with same version (redundant).
			// c1v9 begin -> queue -> "proceedSignal"
			// c1v9-2 begin -> queue -> "ceaseSignal" (redundant)
			// c1v9 end -> queue -> "successSignal"
			// c1v9-2 end -> queue -> "failureSignal"
			events: []event{
				{
					action:     beginRequest,
					req:        "record1version9",
					expectWork: true,
					startWait:  true,
				},
				processEvent, // record1version9 begin
				{
					action:      beginRequest,
					req:         "record1version9other",
					expectWork:  false,
					expectError: ErrSupersededRequest,
				},
				processEvent, // record1version9other begin
				{
					action:    waitRequest,
					req:       "record1version9",
					startWait: true,
				},
				processEvent, // record1version9 end
				processEvent, // record1version9other end
				{
					action: terminateSupervisor,
				},
			},
			wantMetrics: at.MetricSnap{
				Gauges: map[at.MetricGauge]int64{
					at.QueueChanDepth:     0,
					at.ProcessingMapDepth: 0,
					at.WaitingMapDepth:    0,
				},
			},
			wantMsgs: histogramSummaries{
				at.Messages:     4,
				at.Transactions: 2,
			},
			wantDB: []string{
				"record1version9",
			},
		},
		"two requests same record": {
			events: []event{
				{
					action:     beginRequest,
					req:        "record1version9",
					expectWork: true,
				},
				processEvent,
				{
					action:     beginRequest,
					req:        "record1version20",
					expectWork: true,
				},
				processEvent,
				processEvent,
				processEvent,
				{
					action: terminateSupervisor,
				},
			},
			wantMetrics: at.MetricSnap{
				Gauges: map[at.MetricGauge]int64{
					at.QueueChanDepth:     0,
					at.ProcessingMapDepth: 0,
					at.WaitingMapDepth:    0,
				},
			},
			wantMsgs: histogramSummaries{
				at.Messages:     4,
				at.Transactions: 2,
			},
			wantDB: []string{
				"record1version20",
			},
		},
		"two requests same record first canceled before start": {
			// First request is canceled prior to its begin message being processed.  Second request successful.
			// c1v9 begin -> queue
			// c1v9 canceled (causes WithWorker to receive immediate context canceled error response)
			// c1v9 end -> queue (c1v9 deferred func closes worker 'done' channel, then sends end message)
			// queue -> c1v9 begin -> "ceaseSignal" response (blackholed because worker closed 'done' channel above)
			// queue -> c1v9 end -> "failureSignal" response (blackholed because worker closed 'done' channel above)
			// c1v20 begin -> queue -> "proceedSignal" response
			// c1v20 end -> queue -> "successSignal" response
			events: []event{
				{
					action:      beginRequest,
					req:         "record1version9",
					expectWork:  false,
					expectError: context.Canceled,
				},
				{
					action: cancelRequest,
					req:    "record1version9",
				},
				processEvent,
				processEvent,
				{
					action:     beginRequest,
					req:        "record1version20",
					expectWork: true,
				},
				processEvent,
				processEvent,
				{
					action: terminateSupervisor,
				},
			},
			wantMetrics: at.MetricSnap{
				Gauges: map[at.MetricGauge]int64{
					at.QueueChanDepth:     0,
					at.ProcessingMapDepth: 0,
					at.WaitingMapDepth:    0,
				},
			},
			wantMsgs: histogramSummaries{
				at.Messages:     4,
				at.Transactions: 2,
			},
			wantDB: []string{
				"record1version20",
			},
		},
		"two requests same record first canceled after start": {
			// First request is canceled after its begin message has been processed.  Second request successful.
			// c1v9 begin -> queue -> c1v9 begin -> "proceedSignal" response
			// c1v9 canceled (causes work function passed to WithWorker to error prior to performing work)
			// c1v9 end -> queue (sent by deferred function after WithWorker error response)
			// queue -> c1v9 end -> "failureSignal" response (blackholed because worker closed 'done' channel above)
			// c1v20 begin -> queue -> "proceedSignal" response
			// c1v20 end -> queue -> "successSignal" response
			events: []event{
				{
					action:      beginRequest,
					req:         "record1version9",
					expectError: context.Canceled,
					expectWork:  false,
					startWait:   true,
				},
				processEvent,
				{
					action: cancelRequest,
					req:    "record1version9",
				},
				{
					action:    waitRequest,
					req:       "record1version9",
					startWait: true,
				},
				processEvent,
				{
					action:     beginRequest,
					req:        "record1version20",
					expectWork: true,
				},
				processEvent,
				processEvent,
				{
					action: terminateSupervisor,
				},
			},
			wantMetrics: at.MetricSnap{
				Gauges: map[at.MetricGauge]int64{
					at.QueueChanDepth:     0,
					at.ProcessingMapDepth: 0,
					at.WaitingMapDepth:    0,
				},
			},
			wantMsgs: histogramSummaries{
				at.Messages:     4,
				at.Transactions: 2,
			},
			wantDB: []string{
				"record1version20",
			},
		},
		"three requests different record": {
			events: []event{
				{
					action:     beginRequest,
					req:        "record2version10",
					expectWork: true,
				},
				processEvent,
				{
					action:     beginRequest,
					req:        "record1version20",
					expectWork: true,
				},
				{
					action:     beginRequest,
					req:        "record3version30",
					expectWork: true,
				},
				processEvent,
				processEvent,
				processEvent,
				processEvent,
				processEvent,
				{
					action: terminateSupervisor,
				},
			},
			wantMetrics: at.MetricSnap{
				Gauges: map[at.MetricGauge]int64{
					at.QueueChanDepth:     0,
					at.ProcessingMapDepth: 0,
					at.WaitingMapDepth:    0,
				},
			},
			wantMsgs: histogramSummaries{
				at.Messages:     6,
				at.Transactions: 3,
			},
			wantDB: []string{
				"record2version10",
				"record1version20",
				"record3version30",
			},
		},
		"three requests one duplicate of waiting (duplicate is redundant)": {
			// First message begins, and second message arrives and waits.  Duplicate of second message arrives, and replaces
			// it in the waitlist (because it has more time remaining in timeout).
			// c1v8 begin -> queue -> "proceedSignal"
			// c1v9 begin -> queue -> waiting map
			// c1v9-2 begin -> queue -> "failureSignal" (redundant)
			//placed in waiting map (replaces c1v9 which gets "ceaseSignal" response) XXX

			// c1v8 end -> queue -> "successSignal"
			// c1v9 end -> queue -> "successSignal"
			// c1v9-2 end -> queue -> XXX
			events: []event{
				{
					action:     beginRequest,
					req:        "record1version8",
					expectWork: true,
					startWait:  true,
				},
				processEvent, // record1version8 begin
				{
					action:     beginRequest,
					req:        "record1version9",
					expectWork: true,
					startWait:  true,
				},
				processEvent, // record1version9 begin - waitlisted
				{
					action:      beginRequest,
					req:         "record1version9other",
					expectWork:  false,
					expectError: ErrSupersededRequest,
				},
				processEvent, // record1version9other begin - ceaseSignal v9other
				processEvent, // record1version9 end
				{
					action:    waitRequest,
					req:       "record1version8",
					startWait: true,
				},
				processEvent, // record1version8 end
				{
					action:    waitRequest,
					req:       "record1version9",
					startWait: true,
				},
				processEvent, // record1version9 end
				{
					action: terminateSupervisor,
				},
			},
			wantMetrics: at.MetricSnap{
				Gauges: map[at.MetricGauge]int64{
					at.QueueChanDepth:     0,
					at.ProcessingMapDepth: 0,
					at.WaitingMapDepth:    0,
				},
			},
			wantMsgs: histogramSummaries{
				at.Messages:     6,
				at.Transactions: 3,
			},
			wantDB: []string{
				"record1version9",
			},
		},
		"three requests one waiting, third redundant (to second)": {
			// First request processes while second waits for it.  Third request supercedes second.
			// c1v9 begin -> queue -> "proceedSignal" response
			// c1v20 begin -> queue -> c1v20 added to wait map
			// c1v11 begin -> queue -> "ceaseSignal" response (redundant to c1v20)
			// c1v11 end -> queue
			// c1v9 end > queue
			// queue -> c1v11 end -> "failureSignal" response (redundant error)
			// queue -> c1v9 end -> "successSignal" response (c1v20 "proceedSignal" response, removed from wait list)
			// c1v20 end -> queue -> "sucesss" response
			events: []event{
				{
					action:     beginRequest,
					req:        "record1version9",
					expectWork: true,
					finishWait: true,
				},
				processEvent, // begin message start v9
				{
					action:     beginRequest,
					req:        "record1version20",
					expectWork: true,
				},
				processEvent, // begin message waitlist v20
				{
					action:      beginRequest,
					req:         "record1version11",
					expectWork:  false,
					expectError: ErrSupersededRequest,
				},
				processEvent, // begin message ceaseSignal v11
				{
					action:     waitRequest,
					req:        "record1version9",
					finishWait: true,
				},
				processEvent, // end message v11
				processEvent, // end message v9, v20 proceedSignal
				processEvent, // end message v20
				{
					action: terminateSupervisor,
				},
			},
			wantMetrics: at.MetricSnap{
				Gauges: map[at.MetricGauge]int64{
					at.QueueChanDepth:     0,
					at.ProcessingMapDepth: 0,
					at.WaitingMapDepth:    0,
				},
			},
			wantMsgs: histogramSummaries{
				at.Messages:     6,
				at.Transactions: 3,
			},
			wantDB: []string{
				"record1version20",
			},
		},
		"three requests one waiting, then superceded": {
			// First request processes while second waits for it.  Third request supercedes second.
			// c1v9 begin -> queue -> "proceedSignal" response
			// c1v10 begin -> queue -> c1v10 added to wait map
			// c1v11 begin -> queue -> c1v11 added to wait map, c1v10 sent "ceaseSignal" response
			// c1v10 end -> queue
			// c1v9 end > queue
			// queue -> c1v10 end -> "failureSignal" response (superceded error)
			// queue -> c1v9 end -> "successSignal" response
			// c1v11 end -> queue -> "sucesss" response
			events: []event{
				{
					action:     beginRequest,
					req:        "record1version9",
					expectWork: true,
					finishWait: true,
				},
				processEvent, // begin message start v9
				{
					action:      beginRequest,
					req:         "record1version10",
					expectWork:  false,
					expectError: ErrSupersededRequest,
				},
				processEvent, // begin message waitlist v10
				{
					action:     beginRequest,
					req:        "record1version11",
					expectWork: true,
				},
				processEvent, // begin message waitlist v11, ceaseSignal v10
				{
					action:     waitRequest,
					req:        "record1version9",
					finishWait: true,
				},
				processEvent, // end message v10
				processEvent, // end message v9, v11 proceedSignal
				processEvent, // end message v11
				{
					action: terminateSupervisor,
				},
			},
			wantMetrics: at.MetricSnap{
				Gauges: map[at.MetricGauge]int64{
					at.QueueChanDepth:     0,
					at.ProcessingMapDepth: 0,
					at.WaitingMapDepth:    0,
				},
			},
			wantMsgs: histogramSummaries{
				at.Messages:     6,
				at.Transactions: 3,
			},
			wantDB: []string{
				"record1version11",
			},
		},
	}

	// Begin Test Execution Loop
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup service side components.
			arbiter, done, db, ctx, wg, li, err := testSetupWithPolling()
			if debug {
				arbiter, done, db, ctx, wg, li, err = testSetupWithPollingAndSupervisorLogging(t)
			}
			if err != nil {
				t.Fatalf("failed to setup test: %e", err)
			}
			go arbiter.Process()

			type request struct {
				req            testReq
				expectedWork   bool
				resultingWork  *bool
				start          *chan struct{}
				finish         *chan struct{}
				expectedError  error
				resultingError error
				workFunc       func(context.Context) error
				cancel         context.CancelFunc
			}

			r := map[string]*request{}

			for _, event := range tc.events {
				switch event.action {
				case beginRequest:
					var executed bool
					var workFunc func(context.Context) error
					var start chan struct{}
					var finish chan struct{}
					if event.startWait {
						start = make(chan struct{})
					}
					if event.finishWait {
						finish = make(chan struct{})
					}
					reqname := event.req
					workFunc = func(ctx context.Context) error {
						debugLog(t, "workFun reqName %q: enter\n", reqname)
						if start != nil {
							debugLog(t, "workFun reqName %q: start address %p(%p)\n", reqname, &start, start)
							start <- struct{}{}
							close(start)
							debugLog(t, "workFun reqName %q: start wait passed\n", reqname)
						}
						if err := ctx.Err(); err != nil {
							return err
						}
						executed = true
						debugLog(t, "workFun reqName %q: executed passed (%p)\n", reqname, &executed)
						if finish != nil {
							debugLog(t, "workFun reqName %q: finish address %p(%p)\n", reqname, &finish, finish)
							finish <- struct{}{}
							close(finish)
							debugLog(t, "workFun reqName %q: finish wait passed\n", reqname)
						}
						return nil
					}

					cancelCtx, cancelFunc := context.WithCancel(ctx)
					thisReq := request{
						req:           requestDefs[event.req],
						expectedWork:  event.expectWork,
						resultingWork: &executed,
						workFunc:      workFunc,
						cancel:        cancelFunc,
						expectedError: event.expectError,
					}
					if start != nil {
						thisReq.start = &start
					}
					if finish != nil {
						thisReq.finish = &finish
					}
					setupTestItem(&thisReq.req, db)
					r[event.req] = &thisReq
					debugLog(t, "begin reqName %q: %s\n", reqname, spew.Sdump(thisReq))

					wg.Add(1)
					go func() {
						defer wg.Done()
						thisReq.resultingError = arbiter.WithWorker(cancelCtx, &thisReq.req, thisReq.workFunc)
					}()
				case processRequests:
					debugLog(t, "process event")
					<-done
					<-done
				case waitRequest:
					request, ok := r[event.req]
					if !ok {
						t.Fatalf("unable to retrieve request %q", event.req)
					}
					debugLog(t, "wait reqName %q event:\n %s\n request: \n %s\n", event.req, spew.Sdump(event), spew.Sdump(request))
					if event.startWait && request.start == nil {
						t.Fatalf("attempting to wait for request with 'start' = false: %q", event.req)
					}
					if event.finishWait && request.finish == nil {
						t.Fatalf("attempting to wait for request with 'finish' = false: %q", event.req)
					}
					if event.startWait {
						debugLog(t, "wait reqName %q: request.start address %p(%p)\n", event.req, request.start, *request.start)
						<-*request.start
					}
					if event.finishWait {
						debugLog(t, "wait reqName %q: request.finish address %p(%p)\n", event.req, request.finish, *request.finish)
						<-*request.finish
					}
				case cancelRequest:
					request, ok := r[event.req]
					if !ok {
						t.Fatalf("unable to retrieve request %q", event.req)
					}
					request.cancel()
				case terminateSupervisor:
					if event.cancelAll {
						for _, req := range r {
							req.cancel()
						}
					}
					wg.Wait()
					arbiter.Terminate()
					<-done
				}
			}

			// Validate test results.
			for reqName, request := range r {
				debugLog(t, "validate reqName %q: %s\n", reqName, spew.Sdump(request))
				if request.resultingWork != nil && request.expectedWork != *(request.resultingWork) {
					t.Errorf("request %q expected work result %t, was %t", reqName, request.expectedWork, *(request.resultingWork))
				}
				if !errors.Is(request.resultingError, request.expectedError) {
					t.Errorf("request %q expected error %#v, got: %#v", reqName, request.expectedError, request.resultingError)
				}
			}
			results := li.SnapMetrics()
			checkMetrics(t, results, tc.wantMetrics)
			msgs := results.HistogramSummaries()
			checkMessages(t, msgs, tc.wantMsgs)
			for _, expectInDB := range tc.wantDB {
				request, ok := r[expectInDB]
				if !ok {
					t.Errorf("request %q not found in result set", expectInDB)
				} else {
					checkDb(t, db, &request.req)
				}
			}
		})
	}
}

// TODO: Test Supervisor state at various stages of operations

// Supervisor receives begin message and begins processing.  Confirm
// message metric is generated, and processing list is incremented.

// Supervisor receives begin message that is added to waitlist.
// Confirm Metrics shows increased Waitlist count, and wait list contains
// message.

// Supervisor receives a new message with equal key to one already in the
// waitlist.  Confirm new message is placed in waitlist and existing waitlist
// message is sent 'ceaseSignal' response.

// An end message is received, and the finalize function returns an error.  Confirm
// the message metrics have 'finalizefailure' status asserted, that the response
// to the end message response is 'failureSignal', and contains the error.

var benchmarkIterations = 100

func Benchmark_randomRequests(b *testing.B) {
	b.ReportAllocs()
	r := rand.New(rand.NewSource(int64(benchmarkIterations)))

	for n := 0; n < b.N; n++ {
		b.StopTimer()
		arbiter, db, ctx, wg, _, err := testSetup()
		if err != nil {
			b.Fatalf("failed to setup test: %e", err)
		}
		go arbiter.Process()

		testFunc := func(ctx context.Context) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			return nil
		}

		var requests = map[int]*testReq{}

		for i := 1; i <= benchmarkIterations; i++ {
			key := r.Intn(benchmarkIterations) + 1
			request := testReq{
				key:   int64(key),
				value: int64(i), // Always increasing
			}

			setupTestItem(&request, db)
			requests[key] = &request
		}

		b.StartTimer()
		for _, req := range requests {
			wg.Add(1)
			go func(r *testReq) error {
				defer wg.Done()
				return arbiter.WithWorker(ctx, r, testFunc)
			}(req)
		}
		wg.Wait()
		b.StopTimer()
		arbiter.Terminate()
	}
}

func Benchmark_randomRepeatingRequests(b *testing.B) {
	b.ReportAllocs()
	r := rand.New(rand.NewSource(int64(benchmarkIterations)))

	for n := 0; n < b.N; n++ {
		b.StopTimer()
		arbiter, db, ctx, wg, _, err := testSetup()
		if err != nil {
			b.Fatalf("failed to setup test: %e", err)
		}
		go arbiter.Process()

		testFunc := func(ctx context.Context) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			return nil
		}

		var requests = map[int]*testReq{}

		for i := 1; i <= benchmarkIterations; i++ {
			key := r.Intn(benchmarkIterations/10) + 1 // smaller domain
			request := testReq{
				key:   int64(key),
				value: int64(i), // Always increasing
			}

			setupTestItem(&request, db)
			requests[key] = &request
		}

		b.StartTimer()
		for _, req := range requests {
			wg.Add(1)
			go func(r *testReq) error {
				defer wg.Done()
				return arbiter.WithWorker(ctx, r, testFunc)
			}(req)
		}
		wg.Wait()
		b.StopTimer()
		arbiter.Terminate()
		// TODO: Report on request/repeat counts
	}
}

func Benchmark_randomRepeatingRedundantRequests(b *testing.B) {
	b.ReportAllocs()
	r := rand.New(rand.NewSource(int64(benchmarkIterations)))

	for n := 0; n < b.N; n++ {
		b.StopTimer()
		arbiter, db, ctx, wg, _, err := testSetup()
		if err != nil {
			b.Fatalf("failed to setup test: %e", err)
		}
		go arbiter.Process()

		testFunc := func(ctx context.Context) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			return nil
		}

		var requests = map[int]*testReq{}

		for i := 1; i <= benchmarkIterations; i++ {
			key := r.Intn(benchmarkIterations/10) + 1 // smaller domain
			request := testReq{
				key:   int64(key),
				value: int64(r.Intn(i)), // Tending to increase
			}

			setupTestItem(&request, db)
			requests[key] = &request
		}

		b.StartTimer()
		for _, req := range requests {
			wg.Add(1)
			go func(r *testReq) error {
				defer wg.Done()
				return arbiter.WithWorker(ctx, r, testFunc)
			}(req)
		}
		wg.Wait()
		b.StopTimer()
		arbiter.Terminate()
		// TODO: Report on request/redundant counts
	}
}

//
// Test setup and helper functions.
//

// ErrInvalidRequest indicates request has failed the "IsValid()" pre-processing check.
var ErrInvalidRequest = errors.New("invalid request")

// ErrSupersededRequest indicates request has been superseded by a subsequent request prior to be processed.
var ErrSupersededRequest = errors.New("request superseded")

func signalError(t *testing.T, stage, expected, received string, reason error) {
	t.Helper()
	t.Fatalf("expected response to %s message: %s received: %s error: %e", stage, expected, received, reason)
}

func debugLog(t *testing.T, format string, args ...interface{}) {
	if debug {
		t.Logf(format, args...)
	}
}

func testSetupWithPollingAndSupervisorLogging(t *testing.T) (*a.Supervisor, chan struct{}, *mtxMap, context.Context, *sync.WaitGroup, *at.LocalInstrumentor, error) {
	done := make(chan struct{})
	pollDone := func() {
		done <- struct{}{}
	}

	ws := zapcore.AddSync(testWriter{t})
	logger := logging.NewZapLogger(ws)
	supervisorOptions := []a.SupervisorOption{
		a.SetPollFunction(pollDone),
		a.SetLogger(logger),
	}
	s, m, c, w, f, e := testSetup(supervisorOptions...)
	return s, done, m, c, w, f, e
}

func testSetupWithPolling() (*a.Supervisor, chan struct{}, *mtxMap, context.Context, *sync.WaitGroup, *at.LocalInstrumentor, error) {
	done := make(chan struct{})
	pollDone := func() {
		done <- struct{}{}
	}

	supervisorOptions := []a.SupervisorOption{
		a.SetPollFunction(pollDone),
	}
	s, m, c, w, f, e := testSetup(supervisorOptions...)
	return s, done, m, c, w, f, e
}

func testSetup(opts ...a.SupervisorOption) (*a.Supervisor, *mtxMap, context.Context, *sync.WaitGroup, *at.LocalInstrumentor, error) {
	li := at.NewLocalInstrumentor()
	supervisorOptions := []a.SupervisorOption{
		a.SetInstrumentor(li),
	}
	supervisorOptions = append(supervisorOptions, opts...)
	arbiter, err := a.NewSupervisor(supervisorOptions...)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	var db mtxMap
	var wg sync.WaitGroup
	return arbiter, &db, context.Background(), &wg, li, nil
}

func setupTestFuncNoWait() (*bool, func(context.Context) error) {
	var hasRun bool
	testFunc := func(ctx context.Context) error {
		hasRun = true
		if err := ctx.Err(); err != nil {
			return err
		}
		return nil
	}
	return &hasRun, testFunc
}

func setupTestFuncWithWait(exec chan struct{}) (*bool, func(context.Context) error) {
	var hasRun bool
	testFunc := func(ctx context.Context) error {
		if exec != nil {
			<-exec
		}
		hasRun = true
		if err := ctx.Err(); err != nil {
			return err
		}
		return nil
	}
	return &hasRun, testFunc
}

// setupTestItem initializes the functions for a testReq using the
// store provided.
func setupTestItem(ti *testReq, db *mtxMap) {
	if db.db == nil {
		db.init()
	}
	ti.valid = func() error {
		cur := db.get(ti.key)
		if ti.value > cur {
			return nil
		}
		return fmt.Errorf("%w: test item %d:%d not valid", ErrInvalidRequest, ti.key, ti.value)
	}
	ti.finalize = func() error {
		db.set(ti.key, ti.value)
		return nil
	}
}

// Check the summary of messages processed and transactions completed.
func checkMessages(t *testing.T, result histogramSummaries, expect histogramSummaries) {
	t.Helper()

	var err bool
	for histName, expected := range expect {
		if got := result[histName]; got != expected {
			err = true
			t.Logf("Arbiter count %s expected: %d actual %d\n", histName, expected, got)
		}
	}
	if err {
		t.Errorf("Arbiter counts did not match expected values\n")
	}
}

// Check queue depth, waiting and processing map counts against expected value.
func checkMetrics(t *testing.T, result at.MetricSnap, expect at.MetricSnap) {
	t.Helper()
	var err bool
	for gaugeName, expected := range expect.Gauges {
		if got := result.Gauges[gaugeName]; got != expected {
			err = true
			t.Logf("Arbiter metric %s expected: %d actual %d\n", gaugeName, expected, got)
		}
	}
	if err {
		t.Errorf("Arbiter metrics did not match expected values\n")
	}
}

func checkDb(t *testing.T, m *mtxMap, ti *testReq) {
	t.Helper()
	if actual := m.get(ti.key); actual != ti.value {
		t.Errorf("Key %d expected: %d got: %d\n", ti.key, ti.value, actual)
	}
}

// testReq implements arbiter.Request for tests.
type testReq struct {
	key      int64
	value    int64
	valid    func() error
	finalize func() error
}

func (t *testReq) GetKey() int64 {
	return t.key
}

func (t *testReq) Supersedes(o a.Request) error {
	otherTestReq, ok := o.(*testReq)
	if !ok {
		return fmt.Errorf("Failed to cast request as 'testReq'")
	}
	if t.value > otherTestReq.value {
		return nil
	}
	return fmt.Errorf("%w: %d:%d superseded by %d:%d", ErrSupersededRequest, t.key, t.value, otherTestReq.key, otherTestReq.value)
}

func (t *testReq) Valid() error {
	if t.valid == nil {
		return fmt.Errorf("function 'valid' uninitialized")
	}
	return t.valid()
}

func (t *testReq) Finalize() error {
	if t.finalize == nil {
		return fmt.Errorf("function 'finalize' uninitialized")
	}
	return t.finalize()
}

// mtxMap acts as persistent storage during testing.
type mtxMap struct {
	db  map[int64]int64
	mtx sync.Mutex
}

func (m *mtxMap) init() {
	m.db = make(map[int64]int64)
}

func (m *mtxMap) set(k int64, v int64) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.db[k] = v
}

func (m *mtxMap) get(k int64) int64 {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	return m.db[k]
}

func (m *mtxMap) dump() map[int64]int64 {
	dump := make(map[int64]int64)
	m.mtx.Lock()
	for k, v := range m.db {
		dump[k] = v
	}
	m.mtx.Unlock()
	return dump
}

func (m *mtxMap) String() []string {
	var dump []string
	m.mtx.Lock()
	for k, v := range m.db {
		dump = append(dump, fmt.Sprintln("k: ", k, "v: :", v))
	}
	m.mtx.Unlock()
	return dump
}

type testWriter struct {
	t *testing.T
}

func (tw testWriter) Write(p []byte) (n int, err error) {
	tw.t.Log(string(p))
	return len(p), nil
}
