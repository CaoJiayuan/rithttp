package rithttp

import (
	"net/http"
	"time"
)

type ConfigInterceptor func(client *http.Client)
type RequestInterceptor func(req *http.Request)

var (
	PresetUserAgent = "RithHttp/1.0"
	PresetHeader    = http.Header{
		"User-Agent": []string{PresetUserAgent},
	}
)

type Client struct {
	http               *http.Client
	configInterceptor  ConfigInterceptor
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

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	c.bootRequest(req)
	return c.http.Do(req)
}

func (c *Client) AsyncDo(req *http.Request) *ResultHolder {
	c.bootRequest(req)
	holder := &ResultHolder{
		result: make(chan *HttpResponse),
		state:  resultIdle,
	}
	go func() {
		r, e := c.http.Do(req)
		holder.result <- &HttpResponse{
			Response: r,
			Err:      e,
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
		if len(req.Header) == 0 {
			req.Header = PresetHeader
		}
	}
}

func NewClient() *Client {
	return &Client{}
}
