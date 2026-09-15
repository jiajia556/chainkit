package service

import (
	"errors"
	"strings"
	"testing"
)

func TestWrapServiceErrorsAddsLocationAndPreservesCause(t *testing.T) {
	cause := errors.New("root cause")
	err := func() (err error) {
		defer wrapServiceErrors("TestOperation", &err)
		return cause
	}()

	if !errors.Is(err, cause) {
		t.Fatalf("wrapped error does not preserve cause: %v", err)
	}

	var serviceErr *Error
	if !errors.As(err, &serviceErr) {
		t.Fatalf("wrapped error is not *service.Error: %T", err)
	}
	if serviceErr.Operation != "TestOperation" {
		t.Fatalf("operation = %q, want TestOperation", serviceErr.Operation)
	}
	if serviceErr.Line <= 0 || !strings.HasSuffix(serviceErr.File, "service/error_location_test.go") {
		t.Fatalf("unexpected source location: %s:%d", serviceErr.File, serviceErr.Line)
	}
	if !strings.Contains(err.Error(), "service.TestOperation (service/error_location_test.go:") {
		t.Fatalf("error does not contain useful location: %v", err)
	}
}

func TestPublicServiceErrorIncludesBoundary(t *testing.T) {
	var chainService *ChainService
	_, err := chainService.BalanceAt("0x0000000000000000000000000000000000000000")

	var serviceErr *Error
	if !errors.As(err, &serviceErr) {
		t.Fatalf("public error is not *service.Error: %T", err)
	}
	if serviceErr.Operation != "BalanceAt" || !strings.HasSuffix(serviceErr.File, "service/query.go") || serviceErr.Line <= 0 {
		t.Fatalf("unexpected public error location: %v", err)
	}
}

func TestWrapServiceErrorsLeavesNilAlone(t *testing.T) {
	var err error
	wrapServiceErrors("TestOperation", &err)
	if err != nil {
		t.Fatalf("nil error was changed to %v", err)
	}
}
