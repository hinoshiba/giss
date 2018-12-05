package httpcl

import (
	"os"
	"net/url"
	"net/http"
	"crypto/tls"
)

func NewClient() (*http.Client, error) {
	http_proxy := os.Getenv("http_proxy")
	if http_proxy == "" {
		http_proxy = os.Getenv("https_proxy")
	}
	if http_proxy != "" {
		proxy, err := url.Parse(http_proxy)
		if err != nil {
			return nil, err
		}
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{ InsecureSkipVerify: true },
			Proxy: http.ProxyURL(proxy),
		}
		return &http.Client{Transport: tr}, nil
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{ InsecureSkipVerify: true },
	}
	return &http.Client{Transport: tr}, nil
}
