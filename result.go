package rithttp

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type ResponseResolver func(response *HttpResponse)
type ErrorResolver func(err error)

const (
	resultIdle = iota
	resultResolving
	resultResolved
)

type HttpResponse struct {
	Response *http.Response
	Err      error
}

func (r *HttpResponse) UnmarshalJson(v interface{}) error {
	b, e := ioutil.ReadAll(r.Response.Body)

	if e != nil {
		return e
	}

	err := json.Unmarshal(b, v)

	return err
}

func (r *HttpResponse) IsSuccessful() bool {
	return r.Response.StatusCode == 200
}

func (r *HttpResponse) ReadBody() ([]byte, error) {
	return ioutil.ReadAll(r.Response.Body)
}

type ResultHolder struct {
	result     chan *HttpResponse
	response   *HttpResponse
	state      int
	errorState int
}

func (r *ResultHolder) Then(resolver ResponseResolver) *ResultHolder {
	if r.state != resultIdle {
		return r
	}
	r.resolve()

	r.state = resultResolving
	func(res *HttpResponse) {
		if res != nil {
			defer func() {
				r.state = resultResolved
			}()
			resolver(res)
		}
	}(r.response)
	return r
}

func (r *ResultHolder) resolve() {
	if r.response == nil {
		r.response = <-r.result
	}
}

func (r *ResultHolder) GetResponse() *HttpResponse {
	r.resolve()
	return r.response
}

func (r *ResultHolder) Catch(resolver ErrorResolver) *ResultHolder {
	if r.errorState != resultIdle {
		return r
	}
	r.resolve()
	r.errorState = resultResolving
	e := r.response.Err
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
