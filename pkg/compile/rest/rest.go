package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"kusionstack.io/KCLVM/kclvm-go/pkg/spec/gpyrpc"
	"kusionstack.io/kusion/pkg/log"
)

const (
	compilePath       = "/api:protorpc/KclvmService.ExecProgram"
	defaultEndpoint   = "http://127.0.0.1:2021"
	defaultCompileURL = "http://127.0.0.1:2021/api:protorpc/KclvmService.ExecProgram"
	defaultPingURL    = "http://127.0.0.1:2021/api:protorpc/BuiltinService.Ping"
)

type Client struct {
	client *http.Client
}

func newClient(client *http.Client) *Client {
	return &Client{
		client: client,
	}
}

func New() (*Client, error) {
	c := newClient(http.DefaultClient)
	var e error
	for i := 0; i < 10; i++ {
		if err := Ping(c); err != nil {
			log.Errorf(err.Error())
			e = err
		} else {
			return c, nil
		}
		time.Sleep(time.Second)
	}
	return nil, e
}

func Ping(c *Client) error {
	res := new(PingResponse)
	req := gpyrpc.Ping_Args{
		Value: "",
	}
	if _, err := c.Post(defaultPingURL, req, res); err != nil {
		return err
	}
	if len(res.Error) > 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (c *Client) Get(url string, output interface{}) (*http.Response, error) {
	return c.do("GET", url, nil, output)
}

func (c *Client) Post(url string, input interface{}, output interface{}) (*http.Response, error) {
	return c.do("POST", url, input, output)
}

func (c *Client) Put(url string, input interface{}, output interface{}) (*http.Response, error) {
	return c.do("PUT", url, input, output)
}

func (c *Client) Delete(url string, input interface{}, output interface{}) (*http.Response, error) {
	return c.do("DELETE", url, input, output)
}

func (c *Client) do(method string, url string, input interface{}, output interface{}) (*http.Response, error) {
	buf := &bytes.Buffer{}
	if input != nil {
		err := json.NewEncoder(buf).Encode(input)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, url, buf)
	req = withTimeOut(req, 100*time.Second)
	if err != nil {
		return nil, err
	}
	setJsonContentType(req)
	// setJsonAcceptType(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("%v: %v", url, resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(output)
	return resp, err
}

func withTimeOut(req *http.Request, timeout time.Duration) *http.Request {
	ctx, cancel := context.WithCancel(context.TODO())
	time.AfterFunc(timeout, func() {
		cancel()
	})
	return req.WithContext(ctx)
}

func setJsonContentType(req *http.Request) {
	req.Header.Set("Content-Type", "accept/json")
}

func setJsonAcceptType(req *http.Request) {
	req.Header.Set("Accept", "application/json")
}

func (c *Client) Compile(req *gpyrpc.ExecProgram_Args) (*Result, error) {
	res := new(Result)
	_, err := c.Post(defaultCompileURL, req, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
