package sling

import (
	"encoding/json"
	"net/http"
	"net/url"
)

const (
	HEAD   = "HEAD"
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	PATCH  = "PATCH"
	DELETE = "DELETE"
)

// Sling is an HTTP Request builder and sender.
type Sling struct {
	// http Client for doing requests
	httpClient *http.Client
	// HTTP method (GET, POST, etc.)
	Method string
	// raw url string for requests
	RawUrl string
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
// baseSling := sling.New(nil).Base("https://api.io/")
// fooSling := baseSling.Request().Get("foo/")
// barSling := baseSling.Request().Get("bar/")
//
// fooSling and barSling will send requests to https://api.io/foo/ and
// https://api.io/bar/ respectively and baseSling is unmodified.
func (s *Sling) Request() *Sling {
	return &Sling{
		httpClient: s.httpClient,
		Method:     s.Method,
		RawUrl:     s.RawUrl,
	}
}

// Fluent setters

// Base sets the RawUrl. If you intend to extend the url with Path,
// baseUrl should be specified with a trailing slash.
func (s *Sling) Base(rawurl string) *Sling {
	s.RawUrl = rawurl
	return s
}

// Path extends the RawUrl with the given path by resolving the reference to
// an absolute URL. If parsing errors occur, the RawUrl is left unmodified.
func (s *Sling) Path(path string) *Sling {
	baseURL, baseErr := url.Parse(s.RawUrl)
	pathURL, pathErr := url.Parse(path)
	if baseErr == nil && pathErr == nil {
		s.RawUrl = baseURL.ResolveReference(pathURL).String()
		return s
	}
	return s
}

// Head sets the Sling method to HEAD and sets the given pathUrl.
func (s *Sling) Head(pathUrl string) *Sling {
	s.Method = HEAD
	return s.Path(pathUrl)
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

// Put sets the Sling method to PUT and sets the given pathUrl.
func (s *Sling) Put(pathUrl string) *Sling {
	s.Method = PUT
	return s.Path(pathUrl)
}

// Patch sets the Sling method to PATCH and sets the given pathUrl.
func (s *Sling) Patch(pathUrl string) *Sling {
	s.Method = PATCH
	return s.Path(pathUrl)
}

// Delete sets the Sling method to DELETE and sets the given pathUrl.
func (s *Sling) Delete(pathUrl string) *Sling {
	s.Method = DELETE
	return s.Path(pathUrl)
}

// Performing Requests

// NewRequest returns a new http.Request created with the Sling properties.
func (s *Sling) HttpRequest() (*http.Request, error) {
	reqURL, err := url.Parse(s.RawUrl)
	if err != nil {
		return nil, err
	}
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
	if v != nil {
		err = decodeResponse(resp, v)
	}
	return resp, err
}

// decodeResponse decodes Response Body encoded as JSON into the value pointed
// to by v. Caller must provide non-nil v and close resp.Body once complete.
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
