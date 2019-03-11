package requests4go

import "net/http"

type Response struct {
	resp *http.Response
}

func (r *Response) Status() string {
	return r.resp.Status
}
