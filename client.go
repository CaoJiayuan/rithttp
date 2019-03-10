package rithttp

import (
	"net/http"
	"time"
)

type ConfigInterceptor func(client *http.Client)
type RequestInterceptor func(req *http.Request)


type Client struct {
	http              *http.Client
	configInterceptor ConfigInterceptor
	requestInterceptor RequestInterceptor
}

func (c *Client) OnConfig(on ConfigInterceptor) *Client {
	c.configInterceptor = on
	return c
}

func (c *Client) OnRequest(on RequestInterceptor) *Client {
	c.requestInterceptor = on
	return c
}

func (c *Client) Do(req *http.Request) (*http.Response, error)  {
	c.bootRequest(req)
	return c.http.Do(req)
}

func (c *Client) AsyncDo(req *http.Request) *resultHolder {
	c.bootRequest(req)
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

func (c *Client) bootRequest(req *http.Request) {
	if c.http == nil {
		c.http = &http.Client{
			Timeout: 5 * time.Second,
		}
	}
	if c.configInterceptor != nil {
		c.configInterceptor(c.http)
	}

	if c.requestInterceptor != nil {
		c.requestInterceptor(req)
	}
}

func NewClient() *Client {
	return &Client{}
}