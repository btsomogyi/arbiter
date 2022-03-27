package locking

import (
	"context"
	"fmt"

	"github.com/btsomogyi/arbiter/example"
	pb "github.com/btsomogyi/arbiter/example/examplepb"

	"github.com/golang/protobuf/proto"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrFailedDoTheWork indicates a failure to perform the changes associated with
// an updated version.
var ErrFailedDoTheWork = fmt.Errorf("failed to do the thing")

// ErrInferiorVersion indicates a request for a lesser version was made and rejected.
var ErrInferiorVersion = fmt.Errorf("version requested is outdated")

// Versioner implements a trivial record store of Element versions.  Versions are
// considered more recent if 'Version' is higher.
type Versioner struct {
	//Elements map[int64]int64
	Elements example.ElementStore
	workFunc func() error
	//lock     sync.Mutex
	pb.UnimplementedVersionerServer
}

func (v *Versioner) getElementVersion(key int64) (int64, error) {
	//v.lock.Lock()
	//defer v.lock.Unlock()
	val, ok := v.Elements.Get(key)
	if ok {
		return val, nil
	}
	st := status.New(codes.NotFound, "key not found")
	fv := &errdetails.BadRequest_FieldViolation{
		Field:       "Key",
		Description: "The requested key was not found in the stored data",
	}
	br := &errdetails.BadRequest{}
	br.FieldViolations = append(br.FieldViolations, fv)
	return 0, embedGrpcStatus(st, fv)
}

func (v *Versioner) setElementVersionIfGreater(key, version int64) error {
	//v.lock.Lock()
	//defer v.lock.Unlock()
	current, ok := v.Elements.Get(key)
	if !ok || current < version {
		v.Elements.Update(key, version)
		return nil
	}
	st := status.New(codes.AlreadyExists, "request superseded")
	ei := &errdetails.ErrorInfo{
		Reason: "SUPERSEDED",
		Domain: "The requested version has already been superseded by a prior update",
		Metadata: map[string]string{
			"Id":           fmt.Sprintf("%d", key),
			"ThisVersion":  fmt.Sprintf("%d", version),
			"OtherVersion": fmt.Sprintf("%d", current),
		},
	}
	return embedGrpcStatus(st, ei)
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
func NewVersioner(opts ...VersionerOption) *Versioner {
	v := &Versioner{
		Elements: make(map[int64]int64),
	}
	return v
}

// UpdateVersion satisfies the examplepb GRPC Server interface.
func (v *Versioner) UpdateVersion(ctx context.Context, req *pb.UpdateVersionRequest) (*pb.VersionResponse, error) {
	key := req.GetKey().Id
	version := req.GetVersion().Id

	err := doTheWork(ctx, version)
	if err != nil {
		return nil, err
	}

	err = v.setElementVersionIfGreater(key, version)
	if err != nil {
		return nil, err
	}

	return &pb.VersionResponse{
		Key:     &pb.Key{Id: key},
		Version: &pb.Version{Id: version},
	}, nil
}

// GetVersion implements the examplepb GRPC Server GetVersion interface function.
func (v *Versioner) GetVersion(ctx context.Context, req *pb.GetVersionRequest) (*pb.VersionResponse, error) {
	key := req.GetKey().Id
	//version := req.GetVersion().Id

	version, err := v.getElementVersion(key)
	if err != nil {
		return nil, err
	}

	return &pb.VersionResponse{
		Key:     &pb.Key{Id: key},
		Version: &pb.Version{Id: version},
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

func doTheWork(ctx context.Context, version int64) error {
	// This produces the side effects that are represented by the new version
	// (such as updating records, making changes, etc). For S&G we'll simulate
	// that attempts to update to prime versions encounter an error (for a
	// deterministic but sparse error condition).

	if isPrime(version) {
		//return fmt.Errorf("%w: key %d version %d", ErrFailedDoTheWork, key, version)
		st := status.New(codes.ResourceExhausted, "work incomplete")
		ri := &errdetails.ResourceInfo{
			ResourceName: "Version",
			Description:  "The requested side effects were not completed.  So sad.",
		}
		st, err := st.WithDetails(ri)
		if err != nil {
			// If this errored, it will always error,
			// so panic so we can figure out why.
			panic(fmt.Sprintf("Unexpected error attaching metadata: %v", err))
		}
		return st.Err()
	}
	return nil
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
