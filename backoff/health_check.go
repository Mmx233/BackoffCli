package backoff

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type ProbeHealthCheckFn func(ctx context.Context) error

type ProbeHealthCheckerConfig struct {
	CheckInterval    time.Duration
	InitialDelay     time.Duration
	SuccessThreshold int
	FailureThreshold int
}

func NewProbeHealthChecker(fn ProbeHealthCheckFn, conf ProbeHealthCheckerConfig) HealthChecker {
	return func(ctx context.Context) <-chan error {
		errChan := make(chan error, 1)
		var success, failure int
		go func() {
			if conf.InitialDelay != 0 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(conf.InitialDelay):
				}
			}

			for {
				err := fn(ctx)
				if err != nil {
					failure++
					if failure >= conf.FailureThreshold {
						errChan <- err
						return
					}
					success = 0
				} else {
					success++
					if success >= conf.SuccessThreshold {
						errChan <- nil
						return
					}
					failure = 0
				}

				select {
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				case <-time.After(conf.CheckInterval):
					// continue
				}
			}
		}()
		return errChan
	}
}

type HttpProbeHealthCheckConfig struct {
	// If http.Client is not nil, some config will not take effect.
	Client  *http.Client
	Timeout time.Duration

	FollowRedirect bool

	Method string
	URL    string
	Header http.Header
	// HttpStatusCode determines which HTTP code is
	// considered successful. If HttpStatusCode is 0,
	// any status between 200 and 299 is considered a success.
	HttpStatusCode int
	// If Keyword is empty, the health check will pass only
	// when the response body contains the keyword.
	// The keyword must not contain line breaks.
	Keyword string
}

func NewHttpProbeHealthCheckFn(conf HttpProbeHealthCheckConfig) ProbeHealthCheckFn {
	client := conf.Client
	if client == nil {
		client = &http.Client{}
	}
	client.Timeout = conf.Timeout
	if client.Transport == nil {
		client.Transport = &http.Transport{}
	}
	transport := client.Transport.(*http.Transport)
	transport.DisableKeepAlives = true
	if !conf.FollowRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return func(ctx context.Context) error {
		if conf.Timeout != 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, conf.Timeout)
			defer cancel()
		}

		req, err := http.NewRequest(conf.Method, conf.URL, nil)
		if err != nil {
			return err
		}
		req.WithContext(ctx)
		req.Header = conf.Header
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if conf.HttpStatusCode != 0 {
			if resp.StatusCode != conf.HttpStatusCode {
				return &ErrorUnexpectedHttpStatus{HttpStatus: resp.StatusCode}
			}
		} else if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return &ErrorUnexpectedHttpStatus{HttpStatus: resp.StatusCode}
		}

		if conf.Keyword != "" {
			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				if strings.Contains(scanner.Text(), conf.Keyword) {
					return nil
				}
			}
			return &ErrorKeywordNotFound{Keyword: conf.Keyword}
		} else {
			_, err = io.Copy(io.Discard, resp.Body)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

type TcpProbeHealthCheckConfig struct {
	Addr    string
	Timeout time.Duration
	Dialer  net.Dialer
}

func NewTcpProbeHealthCheckFn(conf TcpProbeHealthCheckConfig) ProbeHealthCheckFn {
	return func(ctx context.Context) error {
		if conf.Timeout != 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, conf.Timeout)
			defer cancel()
		}

		conn, err := (&net.Dialer{
			Timeout: conf.Timeout,
		}).DialContext(ctx, "tcp", conf.Addr)
		if err != nil {
			return err
		}
		_ = conn.Close()
		return nil
	}
}
