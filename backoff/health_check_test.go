package backoff

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHttpProbeHealthCheckFn_Request(t *testing.T) {
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
