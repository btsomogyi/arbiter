package example

import (
	"errors"
	"fmt"
)

var (
	// ErrFailedDoTheWork indicates a failure to perform the changes associated with
	// an updated version.
	ErrFailedDoTheWork = errors.New("processing failed")

	// ErrKeyNotFound indicates a request to fetch version associated with a key did
	// not find that key installed.
	ErrKeyNotFound = errors.New("key not found")

	// ErrInferiorVersion indicates a request for a lesser version was made and rejected.
	ErrInferiorVersion = fmt.Errorf("version requested is outdated")
)

func IsPrime(number int64) bool {
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
