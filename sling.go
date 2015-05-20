package sling

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	goquery "github.com/google/go-querystring/query"
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
	formContentType = "application/x-www-form-urlencoded"
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
	// elective querystruct url encoding
	EncodeQueryStructs bool
	// url tagged query structs
	queryStructs []interface{}
	// json tagged body struct
	jsonBody interface{}
	// url tagged body struct (form)
	bodyStruct interface{}
}

// New returns a new Sling with an http DefaultClient.
func New() *Sling {
	return &Sling{
		HttpClient:         http.DefaultClient,
		Header:             make(http.Header),
		EncodeQueryStructs: true,
		queryStructs:       make([]interface{}, 0),
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
		HttpClient:         s.HttpClient,
		Method:             s.Method,
		RawUrl:             s.RawUrl,
		Header:             headerCopy,
		EncodeQueryStructs: s.EncodeQueryStructs,
		queryStructs:       append([]interface{}{}, s.queryStructs...),
		jsonBody:           s.jsonBody,
		bodyStruct:         s.bodyStruct,
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

// Head sets the Sling method to HEAD and sets the given pathURL.
func (s *Sling) Head(pathURL string) *Sling {
	s.Method = HEAD
	return s.Path(pathURL)
}

// Get sets the Sling method to GET and sets the given pathURL.
func (s *Sling) Get(pathURL string) *Sling {
	s.Method = GET
	return s.Path(pathURL)
}

// Post sets the Sling method to POST and sets the given pathURL.
func (s *Sling) Post(pathURL string) *Sling {
	s.Method = POST
	return s.Path(pathURL)
}

// Put sets the Sling method to PUT and sets the given pathURL.
func (s *Sling) Put(pathURL string) *Sling {
	s.Method = PUT
	return s.Path(pathURL)
}

// Patch sets the Sling method to PATCH and sets the given pathURL.
func (s *Sling) Patch(pathURL string) *Sling {
	s.Method = PATCH
	return s.Path(pathURL)
}

// Delete sets the Sling method to DELETE and sets the given pathURL.
func (s *Sling) Delete(pathURL string) *Sling {
	s.Method = DELETE
	return s.Path(pathURL)
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
func (s *Sling) Base(rawURL string) *Sling {
	s.RawUrl = rawURL
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

// BodyStruct sets the Sling's bodyStruct. The value pointed to by the
// bodyStruct will be url encoded to set the Body on new requests.
// The bodyStruct argument should be a pointer to a url tagged struct. See
// https://godoc.org/github.com/google/go-querystring/query for details.
func (s *Sling) BodyStruct(bodyStruct interface{}) *Sling {
	if bodyStruct != nil {
		s.bodyStruct = bodyStruct
		s.Set(contentType, formContentType)
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
	err = addQueryStructs(reqURL, s.queryStructs, s.EncodeQueryStructs)
	if err != nil {
		return nil, err
	}
	body, err := s.getRequestBody()
	if err != nil {
		return nil, err
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
func addQueryStructs(reqURL *url.URL, queryStructs []interface{}, doEncodeQueryStructs bool) error {
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
				if doEncodeQueryStructs {
					urlValues.Add(key, url.QueryEscape(value))
				} else {
					urlValues.Add(key, value)
				}
			}
		}
	}
	if doEncodeQueryStructs {
		reqURL.RawQuery = urlValues.Encode()
	} else {
		reqURL.RawQuery = UnEncodedQueryString(urlValues)
	}
	return nil
}

// UnEncodedQueryString encodes the values into ``Non URL encoded'' form
// ("bar=baz&foo=quux") sorted by key.
func UnEncodedQueryString(v url.Values) string {
	if v == nil {
		return ""
	}
	var buf bytes.Buffer
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		prefix := k + "="
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(prefix)
			buf.WriteString(v)
		}
	}
	return buf.String()
}

// getRequestBody returns the io.Reader which should be used as the body
// of new Requests.
func (s *Sling) getRequestBody() (body io.Reader, err error) {
	if s.jsonBody != nil && s.Header.Get(contentType) == jsonContentType {
		body, err = encodeJSONBody(s.jsonBody)
		if err != nil {
			return nil, err
		}
	} else if s.bodyStruct != nil && s.Header.Get(contentType) == formContentType {
		body, err = encodeBodyStruct(s.bodyStruct)
		if err != nil {
			return nil, err
		}
	}
	return body, nil
}

// encodeJSONBody JSON encodes the value pointed to by jsonBody into an
// io.Reader, typically for use as a Request Body.
func encodeJSONBody(jsonBody interface{}) (io.Reader, error) {
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

// encodeBodyStruct url encodes the value pointed to by bodyStruct into an
// io.Reader, typically for use as a Request Body.
func encodeBodyStruct(bodyStruct interface{}) (io.Reader, error) {
	values, err := goquery.Values(bodyStruct)
	if err != nil {
		return nil, err
	}
	return strings.NewReader(values.Encode()), nil
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
