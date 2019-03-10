package rithttp

import (
	"net/http"
	"time"
)



type Client struct {
	http *http.Client
}

func (c *Client) Do(req *http.Request) (*http.Response, error)  {
	return c.http.Do(req)
}

func (c *Client) AsyncDo(req *http.Request) *resultHolder {
	holder := &resultHolder{
		result: make(chan *HttpResponse),
		state:  resultIdle,
	}
	go func() {
		r, e := c.http.Do(req)
		holder.result <- &HttpResponse{
			resp:r,
			err:e,
		}
	}()

	return holder
}

func NewClient() *Client {
	return &Client{
		http: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}