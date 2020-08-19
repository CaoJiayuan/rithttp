package rithttp

import (
	"net/http"
	"net/url"
	"runtime"
	"time"
)

type ConfigInterceptor func(client *http.Client)
type RequestInterceptor func(req *http.Request)

const version = "0.0.1"

func UserAgent() string {
	return "RithHttp/" + version + " (" + runtime.Version() + ")"
}

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

func (c *Client) Get(url string, query ...url.Values) (*HttpResponse, error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	if len(query) > 0 {
		q := req.URL.Query()
		for k, v := range query[0] {
			for _, s := range v {
				q.Add(k, s)
			}
		}

		req.URL.RawQuery = q.Encode()
	}

	return c.Do(req), nil
}

func (c *Client) Request(method, url string, now ...bool) (*Holder, error) {
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, err
	}

	return c.AsyncDo(req, now...), nil
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
	}

	req.Header.Set("User-Agent", UserAgent())
}

func NewClient() *Client {
	return &Client{}
}
