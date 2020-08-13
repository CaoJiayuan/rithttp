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

type Holder struct {
	req        *http.Request
	result     chan *HttpResponse
	response   *HttpResponse
	client     *Client
	state      int
	errorState int
	sdr        *SimpleDelayRequest
}

func (h *Holder) Then(resolver ResponseResolver) *Holder {
	if h.state != resultIdle {
		return h
	}
	h.do().wait()
	h.state = resultResolving
	func(res *HttpResponse) {
		if res != nil {
			defer func() {
				h.state = resultResolved
			}()
			resolver(res)
		}
	}(h.response)
	return h
}

func (h *Holder) Chain() *SimpleDelayRequest {
	if h.sdr != nil {
		return h.sdr
	}
	h.sdr = &SimpleDelayRequest{
		h.req,
	}
	return h.sdr
}

func (h *Holder) do() *Holder {

	go func() {
		r, e := h.client.http.Do(h.Chain().Request)
		h.result <- &HttpResponse{
			Response: r,
			Err:      e,
		}
	}()
	return h
}

func (h *Holder) wait() {
	if h.response == nil {
		h.response = <-h.result
	}
}

func (h *Holder) GetResponse() *HttpResponse {
	h.wait()
	return h.response
}

func (h *Holder) Catch(resolver ErrorResolver) *Holder {
	if h.errorState != resultIdle {
		return h
	}
	h.wait()
	h.errorState = resultResolving
	e := h.response.Err
	func(err error) {
		if err != nil {
			defer func() {
				h.errorState = resultResolved
			}()
			resolver(err)
		}
	}(e)

	return h
}

type SimpleDelayRequest struct {
	*http.Request
}

func (sdr *SimpleDelayRequest) SetHeader(key, value string) *SimpleDelayRequest {
	sdr.Header.Set(key, value)
	return sdr
}

func (sdr *SimpleDelayRequest) do() {

}
