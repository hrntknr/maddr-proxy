package maddrproxy

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
)

func NewProxyClient(p *proxy, proxyUrlModifyFunc func(*url.URL)) *http.Client {
	proxyInstance := httptest.NewServer(http.HandlerFunc(p.serve))
	return &http.Client{
		Transport: &http.Transport{
			Proxy: func(*http.Request) (*url.URL, error) {
				tmp, err := url.Parse(proxyInstance.URL)
				if err != nil {
					return nil, err
				}
				if proxyUrlModifyFunc != nil {
					proxyUrlModifyFunc(tmp)
				}
				return tmp, nil
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}
