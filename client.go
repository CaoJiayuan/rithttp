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

func (c *Client) Get(url string) {

}

func (c *Client) Do(req *http.Request) *HttpResponse {
	c.bootRequest(req)
	holder := c.AsyncDo(req, true)
	resp := holder.GetResponse()
	return resp
}

func (c *Client) AsyncDo(req *http.Request, now ...bool) *Holder {
	c.bootRequest(req)
	holder := &Holder{
		result: make(chan *HttpResponse),
		state:  resultIdle,
		client: c,
		req:    req,
	}

	if len(now) > 0 && now[0] {
		holder.do()
	}

	return holder
}

func (c *Client) bootRequest(req *http.Request) {
	if c.http == nil {
		c.http = &http.Client{
			Timeout: 5 * time.Second,
		}
		// run config interceptor before first request
		if c.configInterceptor != nil {
			c.configInterceptor(c.http)
		}
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
