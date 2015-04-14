package sling

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// Url-tagged query struct
var paramsA = struct {
	Limit int `url:"limit"`
}{
	30,
}
var paramsB = struct {
	KindName string `url:"kind_name"`
	Count    int    `url:"count"`
}{
	"recent",
	25,
}

// Json-tagged model struct
type FakeModel struct {
	Text          string  `json:"text,omitempty"`
	FavoriteCount int64   `json:"favorite_count,omitempty"`
	Temperature   float64 `json:"temperature,omitempty"`
}

func TestNew(t *testing.T) {
	sling := New()
	if sling.HttpClient != http.DefaultClient {
		t.Errorf("expected %v, got %v", http.DefaultClient, sling.HttpClient)
	}
	if sling.queryStructs == nil {
		t.Errorf("queryStructs not initialized with make")
	}
}

func TestSlingNew(t *testing.T) {
	cases := []*Sling{
		&Sling{HttpClient: &http.Client{}, Method: "GET", RawUrl: "http://example.com"},
		&Sling{HttpClient: nil, Method: "", RawUrl: "http://example.com"},
		&Sling{queryStructs: make([]interface{}, 0)},
		&Sling{queryStructs: []interface{}{paramsA}},
		&Sling{queryStructs: []interface{}{paramsA, paramsB}},
		&Sling{jsonBody: &FakeModel{Text: "a"}},
		&Sling{jsonBody: FakeModel{Text: "a"}},
		&Sling{jsonBody: nil},
	}
	for _, sling := range cases {
		copy := sling.New()
		if copy.HttpClient != sling.HttpClient {
			t.Errorf("expected %p, got %p", sling.HttpClient, copy.HttpClient)
		}
		if copy.Method != sling.Method {
			t.Errorf("expected %s, got %s", sling.Method, copy.Method)
		}
		if copy.RawUrl != sling.RawUrl {
			t.Errorf("expected %s, got %s", sling.RawUrl, copy.RawUrl)
		}
		// queryStruct slice should be a new slice with a copy of the contents
		if len(sling.queryStructs) > 0 {
			// mutating one slice should not mutate the other
			copy.queryStructs[0] = nil
			if sling.queryStructs[0] == nil {
				t.Errorf("queryStructs was a re-slice, expected slice with copied contents")
			}
		}
		// jsonBody should be copied
		if copy.jsonBody != sling.jsonBody {
			t.Errorf("expected %v, got %v")
		}
	}
}

func TestClientSetter(t *testing.T) {
	developerClient := &http.Client{}
	cases := []struct {
		input    *http.Client
		expected *http.Client
	}{
		{nil, http.DefaultClient},
		{developerClient, developerClient},
	}
	for _, c := range cases {
		sling := New()
		sling.Client(c.input)
		if sling.HttpClient != c.expected {
			t.Errorf("expected %v, got %v", c.expected, sling.HttpClient)
		}
	}
}

func TestBaseSetter(t *testing.T) {
	cases := []string{"http://a.io/", "http://b.io", "/path", "path", ""}
	for _, base := range cases {
		sling := New().Base(base)
		if sling.RawUrl != base {
			t.Errorf("expected %s, got %s", base, sling.RawUrl)
		}
	}
}

func TestPathSetter(t *testing.T) {
	cases := []struct {
		rawUrl         string
		path           string
		expectedRawUrl string
	}{
		{"http://a.io/", "foo", "http://a.io/foo"},
		{"http://a.io/", "/foo", "http://a.io/foo"},
		{"http://a.io", "foo", "http://a.io/foo"},
		{"http://a.io", "/foo", "http://a.io/foo"},
		{"http://a.io/foo/", "bar", "http://a.io/foo/bar"},
		// rawUrl should end in trailing slash if it is to be Path extended
		{"http://a.io/foo", "bar", "http://a.io/bar"},
		{"http://a.io/foo", "/bar", "http://a.io/bar"},
		// path extension is absolute
		{"http://a.io", "http://b.io/", "http://b.io/"},
		{"http://a.io/", "http://b.io/", "http://b.io/"},
		{"http://a.io", "http://b.io", "http://b.io"},
		{"http://a.io/", "http://b.io", "http://b.io"},
		// empty base, empty path
		{"", "http://b.io", "http://b.io"},
		{"http://a.io", "", "http://a.io"},
		{"", "", ""},
	}
	for _, c := range cases {
		sling := New().Base(c.rawUrl).Path(c.path)
		if sling.RawUrl != c.expectedRawUrl {
			t.Errorf("expected %s, got %s", c.expectedRawUrl, sling.RawUrl)
		}
	}
}

func TestMethodSetters(t *testing.T) {
	cases := []struct {
		sling          *Sling
		expectedMethod string
	}{
		{New().Head("http://a.io"), HEAD},
		{New().Get("http://a.io"), GET},
		{New().Post("http://a.io"), POST},
		{New().Put("http://a.io"), PUT},
		{New().Patch("http://a.io"), PATCH},
		{New().Delete("http://a.io"), DELETE},
	}
	for _, c := range cases {
		if c.sling.Method != c.expectedMethod {
			t.Errorf("expected method %s, got %s", c.expectedMethod, c.sling.Method)
		}
	}
}

func TestQueryStructSetter(t *testing.T) {
	cases := []struct {
		sling           *Sling
		expectedStructs []interface{}
	}{
		{New(), []interface{}{}},
		{New().QueryStruct(nil), []interface{}{}},
		{New().QueryStruct(paramsA), []interface{}{paramsA}},
		{New().QueryStruct(paramsA).QueryStruct(paramsA), []interface{}{paramsA, paramsA}},
		{New().QueryStruct(paramsA).QueryStruct(paramsB), []interface{}{paramsA, paramsB}},
		{New().QueryStruct(paramsA).New(), []interface{}{paramsA}},
		{New().QueryStruct(paramsA).New().QueryStruct(paramsB), []interface{}{paramsA, paramsB}},
	}

	for _, c := range cases {
		if count := len(c.sling.queryStructs); count != len(c.expectedStructs) {
			t.Errorf("expected length %d, got %d", len(c.expectedStructs), count)
		}
	check:
		for _, expected := range c.expectedStructs {
			for _, param := range c.sling.queryStructs {
				if param == expected {
					continue check
				}
			}
			t.Errorf("expected to find %v in %v", expected, c.sling.queryStructs)
		}
	}
}

func TestJsonBodySetter(t *testing.T) {
	fakeModel := &FakeModel{}
	cases := []struct {
		initial  interface{}
		input    interface{}
		expected interface{}
	}{
		{fakeModel, nil, fakeModel},
		{nil, fakeModel, fakeModel},
	}
	for _, c := range cases {
		sling := New()
		sling.jsonBody = c.initial
		sling.JsonBody(c.input)
		if sling.jsonBody != c.expected {
			t.Errorf("expected %v, got %v", c.expected, sling.jsonBody)
		}
	}
}

func TestRequest_urlAndMethod(t *testing.T) {
	cases := []struct {
		sling          *Sling
		expectedMethod string
		expectedUrl    string
		expectedErr    error
	}{
		{New().Base("http://a.io"), "", "http://a.io", nil},
		{New().Path("http://a.io"), "", "http://a.io", nil},
		{New().Get("http://a.io"), GET, "http://a.io", nil},
		{New().Put("http://a.io"), PUT, "http://a.io", nil},
		{New().Base("http://a.io/").Path("foo"), "", "http://a.io/foo", nil},
		{New().Base("http://a.io/").Post("foo"), POST, "http://a.io/foo", nil},
		// if relative path is an absolute url, base is ignored
		{New().Base("http://a.io").Path("http://b.io"), "", "http://b.io", nil},
		{New().Path("http://a.io").Path("http://b.io"), "", "http://b.io", nil},
		// last method setter takes priority
		{New().Get("http://b.io").Post("http://a.io"), POST, "http://a.io", nil},
		{New().Post("http://a.io/").Put("foo/").Delete("bar"), DELETE, "http://a.io/foo/bar", nil},
		// last Base setter takes priority
		{New().Base("http://a.io").Base("http://b.io"), "", "http://b.io", nil},
		// Path setters are additive
		{New().Base("http://a.io/").Path("foo/").Path("bar"), "", "http://a.io/foo/bar", nil},
		{New().Path("http://a.io/").Path("foo/").Path("bar"), "", "http://a.io/foo/bar", nil},
		// removes extra '/' between base and ref url
		{New().Base("http://a.io/").Get("/foo"), GET, "http://a.io/foo", nil},
	}
	for _, c := range cases {
		req, err := c.sling.Request()
		if err != c.expectedErr {
			t.Errorf("expected error %v, got %v for %+v", c.expectedErr, err, c.sling)
		}
		if req.URL.String() != c.expectedUrl {
			t.Errorf("expected url %s, got %s for %+v", c.expectedUrl, req.URL.String(), c.sling)
		}
		if req.Method != c.expectedMethod {
			t.Errorf("expected method %s, got %s for %+v", c.expectedMethod, req.Method, c.sling)
		}
	}
}

func TestRequest_queryStructs(t *testing.T) {
	cases := []struct {
		sling       *Sling
		expectedUrl string
	}{
		{New().Base("http://a.io").QueryStruct(paramsA), "http://a.io?limit=30"},
		{New().Base("http://a.io").QueryStruct(paramsA).QueryStruct(paramsB), "http://a.io?count=25&kind_name=recent&limit=30"},
		{New().Base("http://a.io/").Path("foo?path=yes").QueryStruct(paramsA), "http://a.io/foo?limit=30&path=yes"},
		{New().Base("http://a.io").QueryStruct(paramsA).New(), "http://a.io?limit=30"},
		{New().Base("http://a.io").QueryStruct(paramsA).New().QueryStruct(paramsB), "http://a.io?count=25&kind_name=recent&limit=30"},
	}
	for _, c := range cases {
		req, _ := c.sling.Request()
		if req.URL.String() != c.expectedUrl {
			t.Errorf("expected url %s, got %s for %+v", c.expectedUrl, req.URL.String(), c.sling)
		}
	}
}

func TestRequest_jsonBody(t *testing.T) {
	cases := []struct {
		sling        *Sling
		expectedBody string // expected Body io.Reader as a string
	}{
		{New().JsonBody(&FakeModel{Text: "note", FavoriteCount: 12}), "{\"text\":\"note\",\"favorite_count\":12}\n"},
		{New().JsonBody(FakeModel{Text: "note", FavoriteCount: 12}), "{\"text\":\"note\",\"favorite_count\":12}\n"},
		{New().JsonBody(&FakeModel{}), "{}\n"},
		{New().JsonBody(FakeModel{}), "{}\n"},
		// setting the jsonBody overrides existing jsonBody
		{New().JsonBody(&FakeModel{}).JsonBody(&FakeModel{Text: "msg"}), "{\"text\":\"msg\"}\n"},
	}
	for _, c := range cases {
		req, _ := c.sling.Request()
		buf := new(bytes.Buffer)
		buf.ReadFrom(req.Body)
		if value := buf.String(); value != c.expectedBody {
			t.Errorf("expected Request.Body %s, got %s", c.expectedBody, value)
		}
	}

	// test that Body is left nil when no JSON struct is set via JsonBody
	slings := []*Sling{
		New().JsonBody(nil),
		New(),
	}
	for _, sling := range slings {
		req, _ := sling.Request()
		if req.Body != nil {
			t.Errorf("expected nil Request.Body, got %v", req.Body)
		}
	}

	// test that expected jsonBody encoding errors occur, use illegal JSON field
	sling := New().JsonBody(&FakeModel{Temperature: math.Inf(1)})
	req, err := sling.Request()
	expectedErr := errors.New("json: unsupported value: +Inf")
	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
	if req != nil {
		t.Errorf("expected nil Request, got %v", req)
	}
}

func TestAddQueryStructs(t *testing.T) {
	cases := []struct {
		rawurl       string
		queryStructs []interface{}
		expected     string
	}{
		{"http://a.io", []interface{}{}, "http://a.io"},
		{"http://a.io", []interface{}{paramsA}, "http://a.io?limit=30"},
		{"http://a.io", []interface{}{paramsA, paramsA}, "http://a.io?limit=30&limit=30"},
		{"http://a.io", []interface{}{paramsA, paramsB}, "http://a.io?count=25&kind_name=recent&limit=30"},
		// don't blow away query values on the RawUrl (parsed into RawQuery)
		{"http://a.io?initial=7", []interface{}{paramsA}, "http://a.io?initial=7&limit=30"},
	}
	for _, c := range cases {
		reqURL, _ := url.Parse(c.rawurl)
		addQueryStructs(reqURL, c.queryStructs)
		if reqURL.String() != c.expected {
			t.Errorf("expected %s, got %s", c.expected, reqURL.String())
		}
	}
}

func TestEncodeJsonBody(t *testing.T) {
	cases := []struct {
		jsonStruct     interface{}
		expectedReader string // expected io.Reader as a string
		expectedErr    error
	}{
		{&FakeModel{Text: "note", FavoriteCount: 12}, "{\"text\":\"note\",\"favorite_count\":12}\n", nil},
		{FakeModel{Text: "note", FavoriteCount: 12}, "{\"text\":\"note\",\"favorite_count\":12}\n", nil},
		// nil argument should return an empty reader
		{nil, "", nil},
		// zero valued json-tagged should return empty object JSON {}
		{&FakeModel{}, "{}\n", nil},
		{FakeModel{}, "{}\n", nil},
		// check that Encode errors are propagated, illegal JSON field
		{FakeModel{Temperature: math.Inf(1)}, "", errors.New("json: unsupported value: +Inf")},
	}
	for _, c := range cases {
		reader, err := encodeJsonBody(c.jsonStruct)
		if c.expectedErr == nil {
			// err expected to be nil, io.Reader should be readable
			if err != nil {
				t.Errorf("expected error %v, got %v", nil, err)
			}
			buf := new(bytes.Buffer)
			buf.ReadFrom(reader)
			if value := buf.String(); value != c.expectedReader {
				fmt.Println(len(value))
				t.Errorf("expected jsonBody string \"%s\", got \"%s\"", c.expectedReader, value)
			}
		} else {
			// err is non-nil, io.Reader is not readable
			if err.Error() != c.expectedErr.Error() {
				t.Errorf("expected error %s, got %s", c.expectedErr.Error(), err.Error())
			}
			if reader != nil {
				t.Errorf("expected jsonBody nil, got %v", reader)
			}
		}
	}
}

func TestDo(t *testing.T) {
	expectedText := "Some text"
	var expectedFavoriteCount int64 = 24
	client, server := mockServer(`{"text":"Some text","favorite_count":24}`)
	defer server.Close()

	sling := New().Client(client)
	req, _ := http.NewRequest("GET", server.URL, nil)
	var model FakeModel
	resp, err := sling.Do(req, &model)

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected %d, got %d", 200, resp.StatusCode)
	}
	expectedReadError := "http: read on closed response body"
	if _, err = ioutil.ReadAll(resp.Body); err == nil || err.Error() != expectedReadError {
		t.Errorf("expected %s, got %v", expectedReadError, err)
	}
	if model.Text != expectedText {
		t.Errorf("expected %s, got %s", expectedText, model.Text)
	}
	if model.FavoriteCount != expectedFavoriteCount {
		t.Errorf("expected %d, got %d", expectedFavoriteCount, model.FavoriteCount)
	}
}

func TestDo_nilV(t *testing.T) {
	client, server := mockServer("")
	defer server.Close()

	sling := New().Client(client)
	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := sling.Do(req, nil)

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected %d, got %d", 200, resp.StatusCode)
	}
	expectedReadError := "http: read on closed response body"
	if _, err = ioutil.ReadAll(resp.Body); err == nil || err.Error() != expectedReadError {
		t.Errorf("expected %s, got %v", expectedReadError, err)
	}
}

// Testing Utils

// mockServer returns an httptest.Server which always returns Responses with
// the given string as the Body with Content-Type application/json.
// The caller must close the test server.
func mockServer(body string) (*http.Client, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, body)
	}))
	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}
	client := &http.Client{Transport: transport}
	return client, server
}
