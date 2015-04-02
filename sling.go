package sling

import (
	"encoding/json"
	"net/http"
	"net/url"
)

const (
	GET  = "GET"
	POST = "POST"
)

// Sling is an HTTP Request builder and sender.
type Sling struct {
	// http Client for doing requests
	httpClient *http.Client
	// HTTP method (GET, POST, etc.)
	Method string
	// base url for requests
	BaseUrl string
	// path url to resolve relative to BaseUrl
	PathUrl string
}

// New returns a new Sling.
func New(httpClient *http.Client) *Sling {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Sling{httpClient: httpClient}
}

// Copy Creation

// Request returns a copy of the Sling, which is useful for creating a new,
// mutable Sling with properties from a base Sling.
// For example,
// baseSling := sling.New()
// baseSling.BaseUrl = "https://api.example.com"
// fooSling := baseSling.Request().Get("/foo")
// barSling := baseSling.Request().Get("/bar")
//
// This creates a Sling which will send requests to https://api.example.com/foo
// and another which will send requests to https://api.example.com/bar.
func (s *Sling) Request() *Sling {
	return &Sling{
		httpClient: s.httpClient,
		Method:     s.Method,
		BaseUrl:    s.BaseUrl,
		PathUrl:    s.PathUrl,
	}
}

// Fluent setters

// Base sets the Sling BaseUrl
func (s *Sling) Base(baseUrl string) *Sling {
	s.BaseUrl = baseUrl
	return s
}

// Path sets the Sling PathUrl.
func (s *Sling) Path(pathUrl string) *Sling {
	s.PathUrl = pathUrl
	return s
}

// Get sets the Sling method to GET and sets the given pathUrl.
func (s *Sling) Get(pathUrl string) *Sling {
	s.Method = GET
	return s.Path(pathUrl)
}

// Post sets the Sling method to POST and sets the given pathUrl.
func (s *Sling) Post(pathUrl string) *Sling {
	s.Method = POST
	return s.Path(pathUrl)
}

// Performing Requests

// NewRequest returns a new http.Request built by merging the Request config
// with the Sling properties (e.g. "http://example.com" + "/resource").
func (s *Sling) HttpRequest() (*http.Request, error) {
	baseURL, err := url.Parse(s.BaseUrl)
	if err != nil {
		return nil, err
	}
	pathURL, err := url.Parse(s.PathUrl)
	if err != nil {
		return nil, err
	}
	reqURL := baseURL.ResolveReference(pathURL)
	req, err := http.NewRequest(s.Method, reqURL.String(), nil)
	if err != nil {
		return nil, err
	}
	return req, err
}

// Fire sends the HTTP request and decodes the response into the value pointed
// to by v. It wraps http.Client.Do, but handles closing the Response Body.
// The Response and any error doing the request are returned.
//
// Note that non-2xx StatusCodes are valid responses, not errors.
func (s *Sling) Fire(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return resp, err
	}
	// when err is nil, resp contains a non-nil resp.Body which must be closed
	defer resp.Body.Close()
	err = decodeResponse(resp, v)
	return resp, err
}

// decodeResponse decodes Response Body encoded as JSON into the value pointed
// to by v. Caller should call resp.Body.Close().
func decodeResponse(resp *http.Response, v interface{}) error {
	return json.NewDecoder(resp.Body).Decode(v)
}

// Do creates a new HTTP request, sends it, and decodes the response into the
// value pointed to by v. Do is shorthand for calling HttpRequest and Fire.
func (s *Sling) Do(v interface{}) (*http.Response, error) {
	req, err := s.HttpRequest()
	if err != nil {
		return nil, err
	}
	return s.Fire(req, v)
}
