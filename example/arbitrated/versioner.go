package arbitrated

import (
	"context"
	"errors"
	"fmt"

	"github.com/btsomogyi/arbiter"
	"github.com/btsomogyi/arbiter/example"
	"github.com/btsomogyi/arbiter/example/examplepb"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Versioner implements the Examplepb GRPC interface using the trivial record backing
// store of an ElementStore.
type Versioner struct {
	Elements   example.ElementStore
	workFunc   func() error
	Supervisor *arbiter.Supervisor
	examplepb.UnimplementedVersionerServer
}

// config contains the adjustable configuraiton of the Versioner.
type config struct {
	workFunc func() error
}

// A VersionerOption is a function that modifies the behavior of a Versioner.
type VersionerOption func(*config) error

// SetWorkFunc sets the workFunc for the Versioner.
func SetWorkFunc(wf func() error) VersionerOption {
	return func(c *config) error {
		c.workFunc = wf
		return nil
	}
}

// NewVersioner constructs and returns an empty Versioner.
func NewVersioner(s *arbiter.Supervisor, opts ...VersionerOption) *Versioner {
	v := &Versioner{
		Elements:   example.NewElementStore(),
		Supervisor: s,
	}
	return v
}

// GetVersion returns the current Element version for the requested key.  Utilizes Arbiter Supervisor
// to ensure serialized access to underlying data (ElementMap).
func (v *Versioner) GetVersion(ctx context.Context, req *examplepb.GetVersionRequest) (*examplepb.VersionResponse, error) {
	key := req.GetKey().Id
	var version int64

	request := &VersionerRequest{
		id: key,
		valid: func() error {
			// Only checking if key is found, ignoring returned value
			_, ok := v.Elements.Get(key)
			if !ok {
				return ErrKeyNotFound
			}
			return nil
		},
		finalize: func() error {
			v, ok := v.Elements.Get(key)
			if !ok {
				return ErrKeyNotFound
			}
			// Assign returned value to enclosed 'version' variable
			version = v
			return nil
		},
	}

	doNoWork := func(ctx context.Context) error {
		// There is no work to be performed when fetching, though logging or other functions
		// may be appropriate here.  The version will be fetched through the execution of the
		// 'finalize' function within the WithWorker closure.
		return nil
	}

	if err := v.Supervisor.WithWorker(ctx, request, doNoWork); err != nil {
		return nil, err
	}
	return &examplepb.VersionResponse{
		Key:     &examplepb.Key{Id: key},
		Version: &examplepb.Version{Id: version},
	}, nil
}

// UpdateVersion updates the current Element version for the requested key, then
// returns current (updated) value.  Utilizes Arbiter Supervisor to ensure serialized
// access to underlying data (ElementMap).
func (v *Versioner) UpdateVersion(ctx context.Context, req *examplepb.UpdateVersionRequest) (*examplepb.VersionResponse, error) {
	key := req.GetKey().Id
	version := req.GetVersion().Id

	request := &VersionerRequest{
		id:      key,
		version: version,
		valid: func() error {
			return isGreaterThanCurrent(v.Elements, key, version)
		},
		finalize: func() error {
			return v.Elements.Update(key, version)
		},
	}

	doTheWork := func(ctx context.Context) error {
		// This produces the side effects that are represented by the new version
		// (such as updating records, making changes, etc). For S&G we'll simulate
		// that attempts to update to prime versions encounter an error (for a
		// deterministic but sparse error condition).

		if isPrime(version) {
			return ErrFailedDoTheWork
		}
		return nil
	}

	if err := v.Supervisor.WithWorker(ctx, request, doTheWork); err != nil {
		return nil, err
	}
	return &examplepb.VersionResponse{
		Key:     &examplepb.Key{Id: key},
		Version: &examplepb.Version{Id: version},
	}, nil
}

func embedGrpcStatus(st *status.Status, msg proto.Message) error {
	st, err := st.WithDetails(msg)
	if err != nil {
		// If this errored, it will always error, so panic so we can figure out why.
		panic(fmt.Sprintf("Unexpected error attaching metadata: %v", err))
	}
	return st.Err()
}

// isGreaterThanCurrent checks if the version provided is greater than the version returned by ElementStore.
// A nil error return indicates that the provided key/value *is* greaten that currently persisted.
func isGreaterThanCurrent(db example.ElementStore, key int64, version int64) error {
	current, err := db.getVersion(key)
	if err != nil {
		if errors.Is(err, ErrDbKeyNotFound) {
			// key not found so anything is greater than empty
			return nil
		}
		return err
	}
	if version > current {
		return nil
	}

	st := status.New(codes.AlreadyExists, "request superseded")
	desc := "The requested version has already been superseded by a prior request"
	ei := &errdetails.ErrorInfo{
		Reason: "SUPERSEDED",
		Domain: desc,
		Metadata: map[string]string{
			"Key":             fmt.Sprintf("%d", key),
			"RequestVersion":  fmt.Sprintf("%d", version),
			"ExistingVersion": fmt.Sprintf("%d", current),
		},
	}
	return embedGrpcStatus(st, ei)
}

func isPrime(number int64) bool {
	if number == 0 || number == 1 {
		return false
	}
	for i := int64(2); i <= number/2; i++ {
		if number%i == 0 {
			return false
		}
	}
	return true
}
