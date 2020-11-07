package deer

import (
	"net/http"
	"time"
)

type Request struct{}
type Response struct {
	Err  error
	Resp *http.Response
}

func (*Request) Get(addr string, timeout time.Duration) *Response {
	var resp Response

	client := http.Client{
		Timeout: timeout,
	}
	resp.Resp, resp.Err = client.Get(addr)

	return &resp
}
