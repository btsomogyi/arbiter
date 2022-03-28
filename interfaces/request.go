package interfaces

// Request is interface used as key to waiting and processing maps.
// Note that GetKey() must return a hashable Value (not enforcable in Go1). TODO [BTS] Go2 Constraints
type Request interface {
	GetKey() int64
	// Valid checks to ensure a request being taken from queue is still valid to process.  Return an error if
	// the request is no longer valid.
	Valid() error
	// Supersedes checks if receiver request represents a newer request than provided request.
	// Return an error if receiver request is superseded, nil error if it supersedes the provided request.
	Supersedes(Request) error
	// Finalize all concluding work/persistence/cleanup once request is processed.
	Finalize() error
}
