package deer

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"
)

type Request struct{}
type Response struct {
	Err   error
	Resp  *http.Response
	Trace Trace
}

type Trace struct {
	DNSLookup        time.Duration `json:"dns_lookup"`
	TCPConnection    time.Duration `json:"tcp_connection"`
	TLSHandshake     time.Duration `json:"tls_handshake"`
	ServerProcessing time.Duration `json:"server_processing"`
	ContentTransfer  time.Duration `json:"content_transfer"`
	Total            time.Duration `json:"total"`
}

const (
	tDnsStart = iota
	tDnsDone
	tConnectStart
	tConnectDone
	tGotConn
	tGotFirstByte
	tTlsStart
	tTlsDone
	tReqStart
	tReqDone
)

func (*Request) Get(address string, timeout time.Duration) *Response {
	var (
		resp  Response
		times [10]time.Time
	)

	req, err := http.NewRequest("GET", address, strings.NewReader(""))
	if err != nil {
		resp.Err = err
		return &resp
	}
	req.Header = http.Header{}
	req.Header["User-Agent"] = []string{"OhDeer/0.0.1"}

	// Inspired by:
	// https://github.com/davecheney/httpstat/blob/master/main.go
	// 🙏
	//
	// https://golang.org/pkg/net/http/httptrace/#ClientTrace
	//
	trace := &httptrace.ClientTrace{
		DNSStart: func(dnsStartInfo httptrace.DNSStartInfo) {
			times[tDnsStart] = time.Now()
		},
		DNSDone: func(dnsDoneInfo httptrace.DNSDoneInfo) {
			times[tDnsDone] = time.Now()
		},
		ConnectStart: func(network, addr string) {
			times[tConnectStart] = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			if err != nil {
				//TODO
			}
			times[tConnectDone] = time.Now()
		},
		GotConn: func(_ httptrace.GotConnInfo) {
			times[tGotConn] = time.Now()
		},
		GotFirstResponseByte: func() {
			times[tGotFirstByte] = time.Now()
		},
		TLSHandshakeStart: func() {
			times[tTlsStart] = time.Now()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, err error) {
			if err != nil {
				//TODO
			}
			times[tTlsDone] = time.Now()
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(context.Background(), trace))

	// TODO: allow to parametrize network to tcp6
	dialCtx := func(ctx context.Context, _, addr string) (net.Conn, error) {
		return (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: timeout,
			DualStack: false,
		}).DialContext(ctx, "tcp4", addr)
	}

	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          20,
		IdleConnTimeout:       timeout,
		TLSHandshakeTimeout:   timeout,
		ExpectContinueTimeout: timeout,
		DialContext:           dialCtx,
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // no redirects
		},
	}

	resp.Resp, resp.Err = client.Do(req)
	io.Copy(ioutil.Discard, resp.Resp.Body) // simulate body read

	times[tReqDone] = time.Now()
	times[tReqStart] = times[tDnsStart]
	if times[tReqStart].IsZero() {
		times[tReqStart] = times[tConnectStart]
	}

	resp.Trace.DNSLookup = times[tDnsDone].Sub(times[tDnsStart])
	resp.Trace.TCPConnection = times[tConnectDone].Sub(times[tConnectStart])
	resp.Trace.TLSHandshake = times[tTlsDone].Sub(times[tTlsStart])
	resp.Trace.ServerProcessing = times[tGotFirstByte].Sub(times[tGotConn])
	resp.Trace.ContentTransfer = times[tReqDone].Sub(times[tGotFirstByte])
	resp.Trace.Total = times[tReqDone].Sub(times[tReqStart])

	// fmt.Println(resp.Trace)

	return &resp
}

func (t Trace) String() string {
	return fmt.Sprintf("DNS: %dms, TCP: %dms, TLS: %dms, Server: %dms, Transfer: %dms, Total: %dms",
		int(t.DNSLookup/time.Millisecond),
		int(t.TCPConnection/time.Millisecond),
		int(t.TLSHandshake/time.Millisecond),
		int(t.ServerProcessing/time.Millisecond),
		int(t.ContentTransfer/time.Millisecond),
		int(t.Total/time.Millisecond),
	)
}
