package backoff

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestProbeHealthChecker_Success(t *testing.T) {
	t.Parallel()

	var count atomic.Uint32
	firstCall := make(chan struct{}, 1)
	errChan := NewProbeHealthChecker(func(ctx context.Context) error {
		if count.Add(1) == 1 {
			firstCall <- struct{}{}
		}
		return nil
	}, ProbeHealthCheckerConfig{
		CheckInterval:    time.Millisecond * 2,
		InitialDelay:     time.Millisecond,
		SuccessThreshold: 5,
	})(context.Background())

	assert.Equal(t, uint32(0), count.Load(), "initial delay not work")
	select {
	case <-firstCall:
		require.Equal(t, uint32(1), count.Load(), "first fn call not work")
	case <-time.After(time.Second):
		t.Fatal(t, "first fn call timeout")
	}

	select {
	case err := <-errChan:
		assert.Nil(t, err, "probe health check fail")
		assert.Equal(t, uint32(5), count.Load(), "success threshold not work properly")
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestProbeHealthChecker_Failure(t *testing.T) {
	t.Parallel()

	var count atomic.Uint32
	errChan := NewProbeHealthChecker(func(ctx context.Context) error {
		count.Add(1)
		return assert.AnError
	}, ProbeHealthCheckerConfig{
		CheckInterval:    time.Millisecond,
		FailureThreshold: 5,
	})(context.Background())

	select {
	case err := <-errChan:
		require.ErrorIs(t, err, assert.AnError)
		assert.Equal(t, uint32(5), count.Load(), "failure threshold not work properly")
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

	assert.Nil(t, NewHttpProbeHealthCheckFn(HttpProbeHealthCheckConfig{
		URL: url,
	})(context.Background()), "request failed")
}

func TestHttpProbeHealthCheckFn_NoRedirect(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/health" {
			http.Redirect(rw, req, "/", http.StatusMovedPermanently)
			return
		}
		rw.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	url := server.URL + "/health"

	var errorUnexpectedHttpStatus *ErrorUnexpectedHttpStatus
	assert.ErrorAs(t, NewHttpProbeHealthCheckFn(HttpProbeHealthCheckConfig{
		URL:            url,
		FollowRedirect: false,
	})(context.Background()), &errorUnexpectedHttpStatus, "no redirect setting not taking effect")
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
	require.Nil(t, NewHttpProbeHealthCheckFn(HttpProbeHealthCheckConfig{
		URL:     url,
		Keyword: "ok",
	})(context.Background()), "expected keyword 'ok'")
	var errorKeywordNotFound *ErrorKeywordNotFound
	assert.ErrorAs(t, NewHttpProbeHealthCheckFn(HttpProbeHealthCheckConfig{
		URL:     url,
		Keyword: "not_ok",
	})(context.Background()), &errorKeywordNotFound, "expected keyword not matched")

	// response code
	require.Nil(t, NewHttpProbeHealthCheckFn(HttpProbeHealthCheckConfig{
		URL:            url,
		HttpStatusCode: 200,
	})(context.Background()), "expected http code 200")
	var errorUnexpectedHttpStatus *ErrorUnexpectedHttpStatus
	assert.ErrorAs(t, NewHttpProbeHealthCheckFn(HttpProbeHealthCheckConfig{
		URL:            url,
		HttpStatusCode: 201,
	})(context.Background()), &errorUnexpectedHttpStatus, "expected http code not match")
}

func TestHttpProbeHealthCheckFn_Timeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		select {
		case <-req.Context().Done():
		case <-time.After(time.Second):
		}
	}))
	defer server.Close()

	assert.NotNil(t, NewHttpProbeHealthCheckFn(HttpProbeHealthCheckConfig{
		URL:     server.URL,
		Timeout: time.Millisecond,
	})(context.Background()), "timeout not taking effect")
}

func TestTcpProbeHealthCheckFn_Success(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "create tcp listener failed")
	defer listener.Close()

	errChan := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		errChan <- NewTcpProbeHealthCheckFn(TcpProbeHealthCheckConfig{
			Addr:    listener.Addr().String(),
			Timeout: time.Millisecond * 5,
		})(ctx)
	}()

	conn, err := listener.Accept()
	require.NoError(t, err, "accept conntection failed")
	_ = conn.Close()

	select {
	case err := <-errChan:
		require.Nil(t, err)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}
