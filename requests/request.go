package requests

import (
	"bytes"
	"io"
	"net/http"
)

type Response struct {
	Url        string
	Body       []byte
	Header     http.Header
	StatusCode int
}

type RequestOptions struct {
	Body   []byte
	Header http.Header
}

func Request(method string, url string, opt *RequestOptions) (*Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	if opt != nil {
		if opt.Body != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(opt.Body))
		}
		if opt.Header != nil {
			req.Header = opt.Header
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &Response{
		Url:        url,
		Body:       buf,
		Header:     resp.Header,
		StatusCode: resp.StatusCode,
	}, nil
}

func Options(url string, opt *RequestOptions) (*Response, error) {
	return Request("OPTIONS", url, opt)
}

func Get(url string, opt *RequestOptions) (*Response, error) {
	return Request("GET", url, opt)
}

func Head(url string, opt *RequestOptions) (*Response, error) {
	return Request("HEAD", url, opt)
}

func Post(url string, opt *RequestOptions) (*Response, error) {
	return Request("POST", url, opt)
}

func Put(url string, opt *RequestOptions) (*Response, error) {
	return Request("PUT", url, opt)
}

func Delete(url string, opt *RequestOptions) (*Response, error) {
	return Request("DELETE", url, opt)
}

func Trace(url string, opt *RequestOptions) (*Response, error) {
	return Request("TRACE", url, opt)
}

func Connect(url string, opt *RequestOptions) (*Response, error) {
	return Request("CONNECT", url, opt)
}
