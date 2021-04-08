// Package yggdrasil provides a client library for interacting with
// the Yggdrasil authentication service.
package yggdrasil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	defaultBaseURL   = "https://authserver.mojang.com/"
	defaultUserAgent = "go-yggdrasil"
)

// A Client manages communication with the Yggdrasil authentication
// service.
type Client struct {
	// HTTP client used to communicate with the API.
	client *http.Client

	// Base URL for API requests. Defaults to Mojang's authentication
	// server.
	BaseURL *url.URL

	// User Agent used when communicating with the API.
	UserAgent string
}

// NewClient returns a new Yggdrasil API client. If a nil httpClient is
// provided, http.DefaultClient will be used.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseUrl, _ := url.Parse(defaultBaseURL)
	
	return &Client{
		client:    httpClient,
		BaseURL:   baseUrl,
		UserAgent: defaultUserAgent,
	}
}

// NewRequest creates an API request. A relative URL can be provided
// in urlStr, in which case it is resolved to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash.
// If specified, the value pointed to by the body is JSON encoded and
// included as the request body.
func (c *Client) NewRequest(method string, urlStr string, body interface{}) (*http.Request, error) {
	// Resolve absolute URL
	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	// Encode body as JSON
	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	// Create the request
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	// Set request headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	return req, nil
}

// Do sends an API request and returns the API response. The API response
// is JSON decoded and stored in the value pointed to by v.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = CheckResponse(resp)
	if err != nil {
		return resp, err
	}

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
	}

	return resp, err
}

type ErrorResponse struct {
	Response *http.Response

	ErrorName string `json:"error"`
	ErrorMessage string `json:"errorMessage"`
	Cause string `json:"cause"`
}

var _ error = (*ErrorResponse)(nil)

func (r *ErrorResponse) Error() string {
	if r.Cause != "" {
		return fmt.Sprintf("%v %v: (%v) %v (caused by %v)",
			r.Response.Request.Method, r.Response.Request.URL,
			r.ErrorName, r.ErrorMessage, r.Cause)
	}

	return fmt.Sprintf("%v %v: (%v) %v",
		r.Response.Request.Method, r.Response.Request.URL,
		r.ErrorName, r.ErrorMessage)
}

// CheckResponse checks the API response for any errors, and returns an
// error if so.
//
// A response is considered in error if the response status code isn't
// equal to 200.
func CheckResponse(r *http.Response) error {
	if r.StatusCode == 200 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		json.Unmarshal(data, errorResponse)
	}

	return errorResponse
}
