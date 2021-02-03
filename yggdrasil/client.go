// Package yggdrasil provides a client library for interacting with
// the Yggdrasil authentication service.
package yggdrasil

import (
	"encoding/json"
	"fmt"
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

	// TODO: pull apart common errors

	return errorResponse
}
