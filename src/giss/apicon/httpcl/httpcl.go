package httpcl

import (
	"time"
	"net/http"
	"net/url"
	"net"
	"crypto/tls"
)

func NewClient(ps string) (*http.Client, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{ InsecureSkipVerify: true },
		TLSHandshakeTimeout: 10 * time.Second,
		Dial: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).Dial,
	}
	if ps != "" {
		pu, err := url.Parse(ps)
		if err != nil {
			return nil, err
		}
		tr.Proxy = http.ProxyURL(pu)
	}

	return &http.Client{Transport: tr}, nil
}
