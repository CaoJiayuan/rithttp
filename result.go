package rithttp

import "net/http"

type ResponseResolver func(response *http.Response)
type ErrorResolver func(err error)

const (
	resultIdle                  = iota
	resultResolving
	resultResolved
)

type resultHolder struct {
	res chan *http.Response
	err chan error
	state int
	errorState int
}

func (r *resultHolder) Then(resolver ResponseResolver) *resultHolder {
	if r.state != resultIdle {
		return r
	}

	r.state = resultResolving
	func(res *http.Response) {
		if res != nil {
			defer func() {
				r.state = resultResolved
			}()
			resolver(res)
		}
	}(<-r.res)
	return r
}

func (r *resultHolder) Catch(resolver ErrorResolver) *resultHolder {
	if r.errorState != resultIdle {
		return r
	}
	r.errorState = resultResolving
	func(err error) {
		if err != nil {
			defer func() {
				r.errorState = resultResolved
			}()
			resolver(err)
		}
	}(<-r.err)
	return r
}
