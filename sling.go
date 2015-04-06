package sling

import (
	"encoding/json"
	goquery "github.com/google/go-querystring/query"
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
	HttpClient *http.Client
	// HTTP method (GET, POST, etc.)
	Method string
	// raw url string for requests
	RawUrl string
	// url tagged query structs
	queryStructs []interface{}
}

// New returns a new Sling with an http DefaultClient.
func New() *Sling {
	return &Sling{
		HttpClient:   http.DefaultClient,
		queryStructs: make([]interface{}, 0),
	}
}

// Copy Creation

// New returns a copy of the Sling. This is useful for creating a new,
// mutable Sling with properties from a base Sling. For example,
//
// 	baseSling := sling.New().Client(client).Base("https://api.io/")
// 	fooSling := baseSling.New().Get("foo/")
// 	barSling := baseSling.New().Get("bar/")
//
// fooSling and barSling will send requests to https://api.io/foo/ and
// https://api.io/bar/ respectively and baseSling is unmodified.
func (s *Sling) New() *Sling {
	return &Sling{
		HttpClient:   s.HttpClient,
		Method:       s.Method,
		RawUrl:       s.RawUrl,
		queryStructs: append([]interface{}{}, s.queryStructs...),
	}
}

// Fluent setters

// Client sets the http Client used to do requests. If a nil client is given,
// the http.DefaultClient will be used.
func (s *Sling) Client(httpClient *http.Client) *Sling {
	if httpClient == nil {
		s.HttpClient = http.DefaultClient
	} else {
		s.HttpClient = httpClient
	}
	return s
}

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

// QueryStruct adds the queryStruct to the slice of queryStructs which are
// encoded as url query parameters when a request is created.
// See https://godoc.org/github.com/google/go-querystring/query for url
// tagging options.
func (s *Sling) QueryStruct(queryStruct interface{}) *Sling {
	if queryStruct != nil {
		s.queryStructs = append(s.queryStructs, queryStruct)
	}
	return s
}

// Performing Requests

// Request returns a new http.Request created with the Sling properties.
func (s *Sling) Request() (*http.Request, error) {
	reqURL, err := url.Parse(s.RawUrl)
	if err != nil {
		return nil, err
	}
	err = addQueryStructs(reqURL, s.queryStructs)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(s.Method, reqURL.String(), nil)
	if err != nil {
		return nil, err
	}
	return req, err
}

// addQueryStructs parses url tagged query structs using go-querystring to
// convert them to url.Values and encode them onto the url.RawQuery. Returns
// any error that occurs during parsing.
func addQueryStructs(reqURL *url.URL, queryStructs []interface{}) error {
	urlValues, err := url.ParseQuery(reqURL.RawQuery)
	if err != nil {
		return err
	}
	for _, queryStruct := range queryStructs {
		queryValues, err := goquery.Values(queryStruct)
		if err != nil {
			return err
		}
		for key, values := range queryValues {
			for _, value := range values {
				urlValues.Add(key, value)
			}
		}
	}
	reqURL.RawQuery = urlValues.Encode()
	return nil
}

// Do sends the HTTP request and decodes the response into the value pointed
// to by v. It wraps http.Client.Do, but handles closing the Response Body.
// The Response and any error doing the request are returned.
//
// Note that non-2xx StatusCodes are valid responses, not errors.
func (s *Sling) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := s.HttpClient.Do(req)
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

// Receive creates a new HTTP request, sends it, and decodes the response into
// the value pointed to by v. Receive is shorthand for calling Request and Do.
func (s *Sling) Receive(v interface{}) (*http.Response, error) {
	req, err := s.Request()
	if err != nil {
		return nil, err
	}
	return s.Do(req, v)
}
