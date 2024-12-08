package maddrproxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hrntknr/maddr-proxy/pkg/utils"
	"golang.org/x/sync/errgroup"
)

const timeout = 10 * time.Second
const proxyAuthHeaderKey = "Proxy-Authorization"

type proxy struct {
	passwords []string
}

func NewProxy(passwords []string) *proxy {
	p := &proxy{
		passwords: passwords,
	}

	return p
}

func (p *proxy) resolveIface(hint string, iface *net.Interface, target string) (net.Addr, string, error) {
	host, _, err := net.SplitHostPort(target)
	if err != nil {
		return nil, "", err
	}
	hosts, err := net.LookupHost(host)
	if err != nil {
		return nil, "", err
	}
	targetHasIPv4, targetHasIPv6 := false, false
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			if ip.To4() == nil {
				targetHasIPv6 = true
			} else {
				targetHasIPv4 = true
			}
		}
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, "", err
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && utils.IsValidIPv6(ipnet.IP) && targetHasIPv6 && (hint == "tcp6" || hint == "tcp") {
			return &net.TCPAddr{IP: ipnet.IP, Port: 0}, "tcp6", nil
		}
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && utils.IsValidIPv4(ipnet.IP) && targetHasIPv4 && (hint == "tcp4" || hint == "tcp") {
			return &net.TCPAddr{IP: ipnet.IP, Port: 0}, "tcp4", nil
		}
	}
	return nil, "", errors.New("no suitable address found")
}

func (p *proxy) resolve(target string, user string) (net.Addr, string, error) {
	if user != "" {
		if ip := net.ParseIP(user); ip != nil {
			addr := &net.TCPAddr{
				IP:   ip,
				Port: 0,
			}
			if ip.To4() == nil {
				return addr, "tcp6", nil
			} else {
				return addr, "tcp4", nil
			}
		} else {
			ifaceName := user
			hint := "tcp"
			if index := strings.Index(user, ":"); index != -1 {
				hint = user[:index]
				ifaceName = user[index+1:]
				if hint != "tcp" && hint != "tcp4" && hint != "tcp6" {
					return nil, "", fmt.Errorf("invalid hint: %s", hint)
				}
			}
			iface, err := net.InterfaceByName(ifaceName)
			if err != nil {
				return nil, "", fmt.Errorf("failed to find interface: %w", err)
			}
			return p.resolveIface(hint, iface, target)
		}
	}
	return nil, "tcp", nil
}

func (p *proxy) handleConn(req *http.Request, user string, conn net.Conn) (net.Conn, error) {
	addr, network, err := p.resolve(req.Host, user)
	if err != nil {
		return nil, err
	}
	peer, err := utils.GetDialContext(timeout, addr)(context.Background(), network, req.Host)
	if err != nil {
		return nil, err
	}

	utils.WriteHttpResponseConn(conn, http.StatusOK, "Connection established", nil)
	return peer, nil
}

func (p *proxy) handleReq(req *http.Request, user string) (net.Conn, error) {
	addr, network, err := p.resolve(req.Host, user)
	if err != nil {
		return nil, err
	}
	peer, err := utils.GetDialContext(timeout, addr)(context.Background(), network, req.Host)
	if err != nil {
		return nil, err
	}

	if err = req.Write(peer); err != nil {
		peer.Close()
		return nil, err
	}

	return peer, nil
}

func (p *proxy) serve(w http.ResponseWriter, req *http.Request) {
	conn, wr, err := w.(http.Hijacker).Hijack()
	if err != nil {
		utils.WriteHttpResponse(wr, http.StatusInternalServerError, "", http.Header{"X-Proxy-Error": []string{err.Error()}})
		return
	}
	defer conn.Close()

	user, code, err := utils.ProxyAuthenticate(proxyAuthHeaderKey, p.passwords, req)
	if err != nil {
		h := http.Header{"X-Proxy-Error": []string{err.Error()}}
		if code == http.StatusProxyAuthRequired {
			h.Set("Proxy-Authenticate", "Basic realm=\"Proxy\"")
		}
		utils.WriteHttpResponse(wr, code, "", h)
		return
	}

	var peer net.Conn
	switch req.Method {
	case http.MethodConnect:
		_peer, err := p.handleConn(req, user, conn)
		if err != nil {
			utils.WriteHttpResponse(wr, http.StatusInternalServerError, "", http.Header{"X-Proxy-Error": []string{err.Error()}})
			return
		}
		peer = _peer
	case http.MethodGet:
		_peer, err := p.handleReq(req, user)
		if err != nil {
			utils.WriteHttpResponse(wr, http.StatusInternalServerError, "", http.Header{"X-Proxy-Error": []string{err.Error()}})
			return
		}
		peer = _peer
	default:
		utils.WriteHttpResponse(wr, http.StatusMethodNotAllowed, "", nil)
		return
	}
	defer peer.Close()

	wg := &errgroup.Group{}
	wg.Go(func() error {
		if _, err := io.Copy(conn, peer); err != nil {
			return err
		}
		conn.Close()
		return nil
	})
	wg.Go(func() error {
		if _, err := io.Copy(peer, conn); err != nil {
			return err
		}
		peer.Close()
		return nil
	})
	if err := wg.Wait(); err != nil {
		utils.WriteHttpResponse(wr, http.StatusMethodNotAllowed, "", http.Header{"X-Proxy-Error": []string{err.Error()}})
		return
	}
}

func (p *proxy) ListenAndServe(addr string) error {
	server := &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(p.serve),
	}

	return server.ListenAndServe()
}
