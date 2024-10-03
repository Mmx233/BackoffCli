package backoff

import (
	"crypto/tls"
	"github.com/Mmx233/BackoffCli/backoff"
	"github.com/Mmx233/BackoffCli/internal/config"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/url"
)

func NewHealthCheckFn(logger log.FieldLogger) (backoff.HealthChecker, error) {
	var healthCheckFn backoff.ProbeHealthCheckFn

	switch {
	case config.Config.TcpAddr != "":
		_, err := net.ResolveUDPAddr("tcp", config.Config.TcpAddr)
		if err != nil {
			return nil, err
		}
		healthCheckFn = backoff.NewTcpProbeHealthCheckFn(backoff.TcpProbeHealthCheckConfig{
			Addr:    config.Config.TcpAddr,
			Timeout: config.Config.TcpTimeout,
			Dialer:  net.Dialer{},
		})
	case config.Config.HttpUrl != "":
		_, err := url.Parse(config.Config.HttpUrl)
		if err != nil {
			return nil, err
		}

		httpClient := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		}
		if !config.Config.HttpFollowRedirect {
			httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
		}
		if config.Config.HttpInsecure {
			httpClient.Transport.(*http.Transport).TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		var header http.Header
		if len(config.Config.HttpHeader) != 0 {
			header = make(http.Header, len(config.Config.HttpHeader))
			for k, v := range config.Config.HttpHeader {
				header.Add(k, v)
			}
		}

		healthCheckFn = backoff.NewHttpProbeHealthCheckFn(backoff.HttpProbeHealthCheckConfig{
			Client:         httpClient,
			Header:         header,
			Method:         config.Config.HttpMethod,
			URL:            config.Config.HttpUrl,
			Timeout:        config.Config.HttpTimeout,
			FollowRedirect: config.Config.HttpFollowRedirect,
			HttpStatusCode: config.Config.HttpStatusCode,
			Keyword:        config.Config.HttpKeyword,
		})
	}

	if healthCheckFn != nil {
		return backoff.NewProbeHealthChecker(healthCheckFn, backoff.ProbeHealthCheckerConfig{
			Logger:           logger,
			CheckInterval:    config.Config.ProbeInterval,
			InitialDelay:     config.Config.ProbeInitialDelay,
			SuccessThreshold: config.Config.ProbeThresholdSuccess,
			FailureThreshold: config.Config.ProbeThresholdFailure,
		}), nil
	}
	return nil, nil
}
