package utils

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

func getAuth(proxyAuthHeaderKey string, r *http.Request) (string, string, error) {
	authHeader := r.Header.Get(proxyAuthHeaderKey)
	if authHeader == "" {
		return "", "", nil
	}
	enc := strings.TrimPrefix(authHeader, "Basic ")
	str, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return "", "", err
	}
	last := strings.LastIndex(string(str), ":")
	if last == -1 {
		return "", "", errors.New("invalid auth header")
	}
	user, password := string(str[:last]), string(str[last+1:])

	return user, password, nil
}

func ProxyAuthenticate(proxyAuthHeaderKey string, passwords []string, r *http.Request) (string, int, error) {
	user, password, err := getAuth(proxyAuthHeaderKey, r)
	if err != nil {
		return "", http.StatusProxyAuthRequired, err
	}

	if len(passwords) == 0 {
		return user, http.StatusOK, nil
	}
	if password == "" {
		return "", http.StatusProxyAuthRequired, errors.New(http.StatusText(http.StatusProxyAuthRequired))
	}

	for _, p := range passwords {
		if p == password {
			return user, http.StatusOK, nil
		}
	}

	return "", http.StatusForbidden, errors.New(http.StatusText(http.StatusForbidden))
}

func WriteHttpResponseConn(conn net.Conn, status int, msg string, headers http.Header) error {
	if msg == "" {
		msg = http.StatusText(status)
	}
	if _, err := conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", status, msg))); err != nil {
		return err
	}
	for key, values := range headers {
		for _, value := range values {
			if _, err := conn.Write([]byte(key + ": " + value + "\r\n")); err != nil {
				return err
			}
		}
	}
	if _, err := conn.Write([]byte("\r\n")); err != nil {
		return err
	}
	return nil
}

func WriteHttpResponse(wr *bufio.ReadWriter, status int, msg string, headers http.Header) error {
	if msg == "" {
		msg = http.StatusText(status)
	}
	if _, err := wr.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", status, msg)); err != nil {
		return err
	}
	for key, values := range headers {
		for _, value := range values {
			if _, err := wr.WriteString(key + ": " + value + "\r\n"); err != nil {
				return err
			}
		}
	}
	if _, err := wr.WriteString("\r\n"); err != nil {
		return err
	}
	if err := wr.Flush(); err != nil {
		return err
	}
	return nil
}

func GetDialContext(timeout time.Duration, localAddr net.Addr) func(context.Context, string, string) (net.Conn, error) {
	return (&net.Dialer{
		Timeout:   timeout,
		LocalAddr: localAddr,
	}).DialContext
}

func IsValidIPv4(ip net.IP) bool {
	if ip.To4() == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsMulticast() {
		return false
	}
	return true
}

func IsValidIPv6(ip net.IP) bool {
	if ip.To4() != nil {
		return false
	}
	if ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() || ip.IsLoopback() || ip.IsMulticast() {
		return false
	}
	return true
}
