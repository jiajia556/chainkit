package service

import (
	"errors"
	"testing"
)

func TestIsFilterLogsRangeLimit(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "too many results",
			err:  &filterLogsQueryError{err: errors.New("query returned more than 10000 results")},
			want: true,
		},
		{
			name: "response too large",
			err:  &filterLogsQueryError{err: errors.New("Log response size exceeded")},
			want: true,
		},
		{
			name: "unrelated rpc error",
			err:  &filterLogsQueryError{err: errors.New("connection reset by peer")},
			want: false,
		},
		{
			name: "non getLogs error",
			err:  errors.New("response too large"),
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isFilterLogsRangeLimit(test.err); got != test.want {
				t.Fatalf("isFilterLogsRangeLimit() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestSuggestedFilterLogsStep(t *testing.T) {
	err := &filterLogsQueryError{
		fromBlock: 115701585,
		toBlock:   115701834,
		err: errors.New(
			"query returned more than 10000 results. Try with this block range [0x6E57751, 0x6E57785]",
		),
	}
	step, ok := suggestedFilterLogsStep(err)
	if !ok {
		t.Fatal("expected provider block range to be accepted")
	}
	if step != 53 {
		t.Fatalf("suggestedFilterLogsStep() = %d, want 53", step)
	}
}

func TestSuggestedFilterLogsStepRejectsGap(t *testing.T) {
	err := &filterLogsQueryError{
		fromBlock: 100,
		toBlock:   200,
		err:       errors.New("Try with this block range [0x65, 0x70]"),
	}
	if _, ok := suggestedFilterLogsStep(err); ok {
		t.Fatal("provider range that skips the attempted first block must be rejected")
	}
}
