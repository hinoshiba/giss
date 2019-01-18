package httpcl

import (
	"time"
	"net/http"
	"net"
	"crypto/tls"
)

func NewClient() (*http.Client, error) {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{ InsecureSkipVerify: true },
		TLSHandshakeTimeout: 10 * time.Second,
		Dial: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).Dial,
	}
	return &http.Client{Transport: tr}, nil
}
