package rithttp

import (
	"net/http"
)

type ResponseResolver func(response *http.Response)
type ErrorResolver func(err error)

const (
	resultIdle                  = iota
	resultResolving
	resultResolved
)

type HttpResponse struct {
	resp *http.Response
	err error
}

type resultHolder struct {
	result     chan *HttpResponse
	response   *HttpResponse
	state      int
	errorState int
}

func (r *resultHolder) Then(resolver ResponseResolver) *resultHolder {
	if r.state != resultIdle {
		return r
	}
	r.resolve()

	r.state = resultResolving
	func(res *http.Response) {
		if res != nil {
			defer func() {
				r.state = resultResolved
			}()
			resolver(res)
		}
	}(r.response.resp)
	return r
}

func (r *resultHolder) resolve()  {
	if r.response == nil {
		r.response = <- r.result
	}
}

func (r *resultHolder) GetResult() chan *HttpResponse{
	return r.result
}

func (r *resultHolder) GetResponse() *http.Response {
	r.resolve()
	return r.response.resp
}

func (r *resultHolder) GetError() error {
	r.resolve()
	return r.response.err
}

func (r *resultHolder) Catch(resolver ErrorResolver) *resultHolder {
	if r.errorState != resultIdle {
		return r
	}
	r.resolve()
	r.errorState = resultResolving
	e := r.response.err
	func(err error) {
		if err != nil {
			defer func() {
				r.errorState = resultResolved
			}()
			resolver(err)
		}
	}(e)

	return r
}
