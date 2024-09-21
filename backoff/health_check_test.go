package backoff

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestProbeHealthChecker_Success(t *testing.T) {
	t.Parallel()

	var count atomic.Uint32
	errChan := NewProbeHealthChecker(func(ctx context.Context) error {
		count.Add(1)
		return nil
	}, ProbeHealthCheckerConfig{
		CheckInterval:    time.Millisecond * 2,
		InitialDelay:     time.Millisecond,
		SuccessThreshold: 5,
	})(context.Background())

	if count.Load() != 0 {
		t.Fatal("initial delay not work")
	}
	time.Sleep(time.Millisecond * 2)
	if count.Load() != 1 {
		t.Fatal("first fn call not work")
	}

	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if count.Load() != 5 {
			t.Fatalf("success threshold not work properly, expected 5, got %d", count.Load())
		}
	case <-time.After(time.Millisecond * 15):
		t.Fatal("timeout")
	}
}

func TestProbeHealthChecker_Failure(t *testing.T) {
	t.Parallel()

	var count atomic.Uint32
	expectedErr := errors.New("test error")
	errChan := NewProbeHealthChecker(func(ctx context.Context) error {
		count.Add(1)
		return expectedErr
	}, ProbeHealthCheckerConfig{
		CheckInterval:    time.Millisecond,
		FailureThreshold: 5,
	})(context.Background())

	select {
	case err := <-errChan:
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected error: '%v', got: '%v'", expectedErr, err)
		}
		if count.Load() != 5 {
			t.Fatalf("failure threshold not work properly, expected 5, got %d", count.Load())
		}
	case <-time.After(time.Millisecond * 10):
		t.Fatal("timeout")
	}
}

func TestHttpProbeHealthCheckFn_Request(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/health" {
			rw.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	url := server.URL + "/health"

	if err := NewHttpProbeHealthCheckFn(HttpProbeHealthCheckConfig{
		URL: url,
	})(context.Background()); err != nil {
		t.Fatalf("request failed: %v", err)
	}
}

func TestHttpProbeHealthCheckFn_Response_Check(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/health" {
			_, _ = rw.Write([]byte("ok"))
		}
	}))
	defer server.Close()

	url := server.URL + "/health"

	// keyword
	if err := NewHttpProbeHealthCheckFn(HttpProbeHealthCheckConfig{
		URL:     url,
		Keyword: "ok",
	})(context.Background()); err != nil {
		t.Fatalf("expected keyword '%s', got error: %v", "ok", err)
	}
	if err := NewHttpProbeHealthCheckFn(HttpProbeHealthCheckConfig{
		URL:     url,
		Keyword: "not_ok",
	})(context.Background()); true {
		var errorKeywordNotFound *ErrorKeywordNotFound
		ok := errors.As(err, &errorKeywordNotFound)
		if !ok {
			t.Errorf("expected keyword not found, got error: %v", err)
		}
	}

	// response code
	if err := NewHttpProbeHealthCheckFn(HttpProbeHealthCheckConfig{
		URL:            url,
		HttpStatusCode: 200,
	})(context.Background()); err != nil {
		t.Fatalf("expected http code 200, got error: %v", err)
	}
	if err := NewHttpProbeHealthCheckFn(HttpProbeHealthCheckConfig{
		URL:            url,
		HttpStatusCode: 201,
	})(context.Background()); true {
		var errorKeywordNotFound *ErrorUnexpectedHttpStatus
		ok := errors.As(err, &errorKeywordNotFound)
		if !ok {
			t.Fatalf("expected http code not match, got error: %v", err)
		}
	}
}

func TestHttpProbeHealthCheckFn_Timeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(time.Second)
	}))
	defer server.Close()

	fn := NewHttpProbeHealthCheckFn(HttpProbeHealthCheckConfig{
		URL:     server.URL,
		Timeout: time.Millisecond,
	})
	err := fn(context.Background())
	if err == nil {
		t.Error("timeout not taking effect")
	}
}
