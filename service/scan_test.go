package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWaitForRPCRequestEnforcesMinimumInterval(t *testing.T) {
	service := &ChainService{}
	interval := 25 * time.Millisecond
	service.SetRPCRequestInterval(interval)

	if err := service.waitForRPCRequest(context.Background()); err != nil {
		t.Fatal(err)
	}
	startedAt := time.Now()
	if err := service.waitForRPCRequest(context.Background()); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(startedAt); elapsed < interval-2*time.Millisecond {
		t.Fatalf("second RPC request waited %v, want at least %v", elapsed, interval)
	}
}

func TestWaitForRPCRequestHonorsContext(t *testing.T) {
	service := &ChainService{}
	service.SetRPCRequestInterval(time.Second)
	if err := service.waitForRPCRequest(context.Background()); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	startedAt := time.Now()
	err := service.waitForRPCRequest(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("waitForRPCRequest() error = %v, want context canceled", err)
	}
	if elapsed := time.Since(startedAt); elapsed > 100*time.Millisecond {
		t.Fatalf("canceled wait took too long: %v", elapsed)
	}
}

func TestWaitForRPCRequestIsSharedAcrossServices(t *testing.T) {
	limiter := &rpcRequestLimiter{}
	first := &ChainService{rpcRequestLimiter: limiter}
	second := &ChainService{rpcRequestLimiter: limiter}
	interval := 25 * time.Millisecond
	first.SetRPCRequestInterval(interval)

	if err := first.waitForRPCRequest(context.Background()); err != nil {
		t.Fatal(err)
	}
	startedAt := time.Now()
	if err := second.waitForRPCRequest(context.Background()); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(startedAt); elapsed < interval-2*time.Millisecond {
		t.Fatalf("request on second service waited %v, want at least %v", elapsed, interval)
	}
}

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
