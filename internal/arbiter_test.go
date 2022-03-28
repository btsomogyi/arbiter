package internal

import (
	"testing"
)

func Test_MessageStatus(t *testing.T) {
	tests := map[string]struct {
		start    messageStatus
		actions  func(messageStatus) string
		expected string
	}{
		"one result": {
			start: msSuccess,
			actions: func(in messageStatus) string {
				return in.results()
			},
			expected: "success",
		},
		"multiple results": {
			start: msSuccess + msWaitlist,
			actions: func(in messageStatus) string {
				return in.results() + " " + in.waitlist()
			},
			expected: "success true",
		},
		"addStatus": {
			start: 0,
			actions: func(in messageStatus) string {
				in.addStatus(msCease)
				return in.results()
			},
			expected: "cease",
		},
		"removeStatus": {
			start: msFinalizeFailure,
			actions: func(in messageStatus) string {
				in.removeStatus(msFinalizeFailure)
				return in.results()
			},
			expected: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup service side components.
			result := tc.actions(tc.start)
			if result != tc.expected {
				t.Errorf("expected %q got %q", tc.expected, result)
			}
		})
	}
}
