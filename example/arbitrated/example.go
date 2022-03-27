package arbitrated

import (
	"fmt"

	"github.com/btsomogyi/arbiter"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrFailedDoTheWork indicates a failure to perform the changes associated with
// an updated version.
var ErrFailedDoTheWork error

// ErrKeyNotFound indicates a request to fetch version associated with a key did
// not find that key installed.
var ErrKeyNotFound error

func init() {
	// Initialize GRPC codes for ErrFailedDoTheWork
	st1 := status.New(codes.ResourceExhausted, "work incomplete")
	ri := &errdetails.ResourceInfo{
		ResourceName: "Version",
		Description:  "The requested side effects were not completed.  So sad.",
	}
	ErrFailedDoTheWork = embedGrpcStatus(st1, ri)

	// Initialize GRPC codes for ErrKeyNotFound
	st2 := status.New(codes.NotFound, "key not found")
	fv := &errdetails.BadRequest_FieldViolation{
		Field:       "Key",
		Description: "The requested key was not found in the stored data",
	}
	br := &errdetails.BadRequest{}
	br.FieldViolations = append(br.FieldViolations, fv)
	ErrKeyNotFound = embedGrpcStatus(st2, br)
}

// VersionerRequest implements the Arbiter package 'Request' interface to be
// used with the Arbiter Supervisor.
type VersionerRequest struct {
	id       int64
	version  int64
	valid    func() error
	finalize func() error
}

// GetKey returns the element id, to be used as Request map key.
func (v *VersionerRequest) GetKey() int64 {
	return v.id
}

// GetValue returns the chassis version, to be used as the Request value.
func (v *VersionerRequest) GetValue() int64 {
	return v.version
}

// Supersedes determines whether a given request has a higher version number
// than the request being compared.
func (v *VersionerRequest) Supersedes(o arbiter.Request) error {
	otherVersionReq, ok := o.(*VersionerRequest)
	if !ok {
		st := status.New(codes.Internal, "failed to cast request as 'VersionerRequest'")
		ei := &errdetails.ErrorInfo{
			Reason: "INTERNAL",
			Domain: "VersionerRequest.Supersedes()",
			Metadata: map[string]string{
				"OtherRequestId":     fmt.Sprintf("%d", o.GetKey()),
				"ThisRequestId":      fmt.Sprintf("%d", v.id),
				"ThisRequestVersion": fmt.Sprintf("%d", v.version),
			},
		}
		return embedGrpcStatus(st, ei)
	}

	thisVersion := v.version
	otherVersion := otherVersionReq.version
	if thisVersion > otherVersion {
		return nil
	}

	st := status.New(codes.AlreadyExists, "request superseded")
	ei := &errdetails.ErrorInfo{
		Reason: "SUPERSEDED",
		Domain: "The requested version has already been superseded by a prior update",
		Metadata: map[string]string{
			"Id":           fmt.Sprintf("%d", v.id),
			"ThisVersion":  fmt.Sprintf("%d", thisVersion),
			"OtherVersion": fmt.Sprintf("%d", otherVersion),
		},
	}
	return embedGrpcStatus(st, ei)
}

// Valid determines whether a request is valid to be processed.
func (v *VersionerRequest) Valid() error {
	if v.valid == nil {
		return fmt.Errorf("ElementRequest id %d uninitialized 'isValid' function", v.id)
	}
	return v.valid()
}

// Finalize writes the completed version to the persistent datastore.
func (v *VersionerRequest) Finalize() error {
	if v.finalize == nil {
		return fmt.Errorf("ElementRequest id %d uninitialized 'finalize' function", v.id)
	}
	return v.finalize()
}

var _ arbiter.Request = (*VersionerRequest)(nil)
