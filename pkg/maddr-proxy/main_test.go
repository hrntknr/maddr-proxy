package maddrproxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestHttpsProxy(t *testing.T) {
	dummyServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	proxy := NewProxy([]string{})
	resp, err := NewProxyClient(proxy, nil).Get(dummyServer.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestHttpProxy(t *testing.T) {
	tt := []struct {
		name          string
		auth          []string
		urlModifyFunc func(*url.URL)
		headers       map[string][]string
		expected      int
	}{
		{
			name:     "normal",
			expected: http.StatusOK,
		},
		{
			name:     "auth required",
			auth:     []string{"password"},
			expected: http.StatusProxyAuthRequired,
		},
		{
			name: "auth forbidden",
			auth: []string{"password"},
			urlModifyFunc: func(u *url.URL) {
				u.User = url.UserPassword("", "test")
			},
			expected: http.StatusForbidden,
		},
		{
			name: "auth ok",
			auth: []string{"password"},
			urlModifyFunc: func(u *url.URL) {
				u.User = url.UserPassword("", "password")
			},
			expected: http.StatusOK,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			dummyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			proxy := NewProxy(tc.auth)
			u, _ := url.Parse(dummyServer.URL)
			resp, err := NewProxyClient(proxy, tc.urlModifyFunc).Do(&http.Request{
				Method: http.MethodGet,
				URL:    u,
				Header: http.Header(tc.headers),
			})
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tc.expected {
				t.Fatalf("expected status %d, got %d", tc.expected, resp.StatusCode)
			}
		})
	}
}
