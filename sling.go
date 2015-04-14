package sling

import (
	"bytes"
	"encoding/json"
	goquery "github.com/google/go-querystring/query"
	"io"
	"net/http"
	"net/url"
)

const (
	HEAD            = "HEAD"
	GET             = "GET"
	POST            = "POST"
	PUT             = "PUT"
	PATCH           = "PATCH"
	DELETE          = "DELETE"
	contentType     = "Content-Type"
	jsonContentType = "application/json"
)

// Sling is an HTTP Request builder and sender.
type Sling struct {
	// http Client for doing requests
	HttpClient *http.Client
	// HTTP method (GET, POST, etc.)
	Method string
	// raw url string for requests
	RawUrl string
	// stores key-values pairs to add to request's Headers
	Header http.Header
	// url tagged query structs
	queryStructs []interface{}
	// json tagged body struct
	jsonBody interface{}
}

// New returns a new Sling with an http DefaultClient.
func New() *Sling {
	return &Sling{
		HttpClient:   http.DefaultClient,
		Header:       make(http.Header),
		queryStructs: make([]interface{}, 0),
	}
}

// New returns a copy of a Sling for creating a new Sling with properties
// from a base Sling. For example,
//
// 	baseSling := sling.New().Client(client).Base("https://api.io/")
// 	fooSling := baseSling.New().Get("foo/")
// 	barSling := baseSling.New().Get("bar/")
//
// fooSling and barSling will both use the same client, but send requests to
// https://api.io/foo/ and https://api.io/bar/ respectively.
//
// Note that jsonBody and queryStructs item values are copied so if pointer
// values are used, mutating the original value will mutate the value within
// the child Sling.
func (s *Sling) New() *Sling {
	// copy Headers pairs into new Header map
	headerCopy := make(http.Header)
	for k, v := range s.Header {
		headerCopy[k] = v
	}
	return &Sling{
		HttpClient:   s.HttpClient,
		Method:       s.Method,
		RawUrl:       s.RawUrl,
		Header:       headerCopy,
		queryStructs: append([]interface{}{}, s.queryStructs...),
		jsonBody:     s.jsonBody,
	}
}

// Http Client

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

// Method

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

// Header

// Add adds the key, value pair in Headers, appending values for existing keys
// to the key's values. Header keys are canonicalized.
func (s *Sling) Add(key, value string) *Sling {
	s.Header.Add(key, value)
	return s
}

// Set sets the key, value pair in Headers, replacing existing values
// associated with key. Header keys are canonicalized.
func (s *Sling) Set(key, value string) *Sling {
	s.Header.Set(key, value)
	return s
}

// Url

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

// QueryStruct appends the queryStruct to the Sling's queryStructs. The value
// pointed to by each queryStruct will be encoded as url query parameters on
// new requests (see Request()).
// The queryStruct argument should be a pointer to a url tagged struct. See
// https://godoc.org/github.com/google/go-querystring/query for details.
func (s *Sling) QueryStruct(queryStruct interface{}) *Sling {
	if queryStruct != nil {
		s.queryStructs = append(s.queryStructs, queryStruct)
	}
	return s
}

// Body

// JsonBody sets the Sling's jsonBody. The value pointed to by the jsonBody
// will be JSON encoded to set the Body on new requests (see Request()).
// The jsonBody argument should be a pointer to a json tagged struct. See
// https://golang.org/pkg/encoding/json/#MarshalIndent for details.
func (s *Sling) JsonBody(jsonBody interface{}) *Sling {
	if jsonBody != nil {
		s.jsonBody = jsonBody
		s.Set(contentType, jsonContentType)
	}
	return s
}

// Requests

// Request returns a new http.Request created with the Sling properties.
// Returns any errors parsing the RawUrl, encoding query structs, encoding
// the body, or creating the http.Request.
func (s *Sling) Request() (*http.Request, error) {
	reqURL, err := url.Parse(s.RawUrl)
	if err != nil {
		return nil, err
	}
	err = addQueryStructs(reqURL, s.queryStructs)
	if err != nil {
		return nil, err
	}
	var body io.Reader
	if s.jsonBody != nil {
		body, err = encodeJsonBody(s.jsonBody)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(s.Method, reqURL.String(), body)
	if err != nil {
		return nil, err
	}
	addHeaders(req, s.Header)
	return req, err
}

// addQueryStructs parses url tagged query structs using go-querystring to
// encode them to url.Values and format them onto the url.RawQuery. Any
// query parsing or encoding errors are returned.
func addQueryStructs(reqURL *url.URL, queryStructs []interface{}) error {
	urlValues, err := url.ParseQuery(reqURL.RawQuery)
	if err != nil {
		return err
	}
	// encodes query structs into a url.Values map and merges maps
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
	// url.Values format to a sorted "url encoded" string, e.g. "key=val&foo=bar"
	reqURL.RawQuery = urlValues.Encode()
	return nil
}

// encodeJsonBody JSON encodes the value pointed to by jsonBody into an
// io.Reader, typically for use as a Request Body.
func encodeJsonBody(jsonBody interface{}) (io.Reader, error) {
	var buf = new(bytes.Buffer)
	if jsonBody != nil {
		buf = &bytes.Buffer{}
		err := json.NewEncoder(buf).Encode(jsonBody)
		if err != nil {
			return nil, err
		}
	}
	return buf, nil
}

// addHeaders adds the key, value pairs from the given http.Header to the
// request. Values for existing keys are appended to the keys values.
func addHeaders(req *http.Request, header http.Header) {
	for key, values := range header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
}

// Sending

// Receive creates a new HTTP request, sends it, and decodes the response into
// the value pointed to by v. Receive is shorthand for calling Request and Do.
func (s *Sling) Receive(v interface{}) (*http.Response, error) {
	req, err := s.Request()
	if err != nil {
		return nil, err
	}
	return s.Do(req, v)
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
