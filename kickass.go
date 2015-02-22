package kickass

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	libraryVersion = "0.1"
	userAgent      = "go-kickass/" + libraryVersion
	defaultBaseURL = "https://kickass.to/"
)

type Client struct {
	// HTTP client used to communicate with the kickass.
	client *http.Client

	// Base URL for kickass requests. Defaults to https://kickass.to/,
	// but can be set to the sandbox url. BaseURL should always end with a
	// trailing slash.
	BaseURL *url.URL

	// User agent used when communicating with the kickass.to.
	UserAgent string
}

func NewClient(client *http.Client) *Client {
	if client == nil {
		client = http.DefaultClient
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{client: client, BaseURL: baseURL, UserAgent: userAgent}

	return c
}

func (c *Client) NewRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}
	if c.UserAgent != "" {
		req.Header.Add("User-Agent", c.UserAgent)
	}

	return req, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	err = CheckResponse(resp)
	if err != nil {
		return resp, err
	}

	return resp, err
}

type Error int

func (e Error) Error() string {
	return fmt.Sprintf("kickass: %d", e)
}

func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	return Error(r.StatusCode)
}
