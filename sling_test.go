package sling

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestNew(t *testing.T) {
	sling := New()
	if sling.HttpClient != http.DefaultClient {
		t.Errorf("expected %v, got %v", http.DefaultClient, sling.HttpClient)
	}
	if sling.queryStructs == nil {
		t.Errorf("queryStructs not initialized with make")
	}
}

// i.e Sling.Request()
func TestCopy(t *testing.T) {
	cases := []*Sling{
		&Sling{HttpClient: &http.Client{}, Method: "GET", RawUrl: "http://example.com"},
		&Sling{HttpClient: nil, Method: "", RawUrl: "http://example.com"},
		&Sling{queryStructs: make([]interface{}, 0)},
		&Sling{queryStructs: []interface{}{paramsA}},
		&Sling{queryStructs: []interface{}{paramsA, paramsB}},
	}
	for _, sling := range cases {
		copy := sling.Request()
		if copy.HttpClient != sling.HttpClient {
			t.Errorf("expected %p, got %p", sling.HttpClient, copy.HttpClient)
		}
		if copy.Method != sling.Method {
			t.Errorf("expected %s, got %s", sling.Method, copy.Method)
		}
		if copy.RawUrl != sling.RawUrl {
			t.Errorf("expected %s, got %s", sling.RawUrl, copy.RawUrl)
		}
		if len(sling.queryStructs) > 0 {
			// mutating the queryStructs must not mutate original
			copy.queryStructs[0] = nil
			if sling.queryStructs[0] == nil {
				t.Errorf("copy's queryStructs is a re-slice, not a copy")
			}
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
		{New().QueryStruct(paramsA).Request(), []interface{}{paramsA}},
		{New().QueryStruct(paramsA).Request().QueryStruct(paramsB), []interface{}{paramsA, paramsB}},
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

func TestHttpRequest_urlAndMethod(t *testing.T) {
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
		req, err := c.sling.HttpRequest()
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

func TestHttpRequest_queryStructs(t *testing.T) {
	cases := []struct {
		sling       *Sling
		expectedUrl string
	}{
		{New().Base("http://a.io").QueryStruct(paramsA), "http://a.io?limit=30"},
		{New().Base("http://a.io").QueryStruct(paramsA).QueryStruct(paramsB), "http://a.io?count=25&kind_name=recent&limit=30"},
		{New().Base("http://a.io/").Path("foo?path=yes").QueryStruct(paramsA), "http://a.io/foo?limit=30&path=yes"},
		{New().Base("http://a.io").QueryStruct(paramsA).Request(), "http://a.io?limit=30"},
		{New().Base("http://a.io").QueryStruct(paramsA).Request().QueryStruct(paramsB), "http://a.io?count=25&kind_name=recent&limit=30"},
	}
	for _, c := range cases {
		req, _ := c.sling.HttpRequest()
		if req.URL.String() != c.expectedUrl {
			t.Errorf("expected url %s, got %s for %+v", c.expectedUrl, req.URL.String(), c.sling)
		}
	}
}

// mockServer returns an httptest.Server which always returns the given body.
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

type FakeModel struct {
	Text          string `json:"text"`
	FavoriteCount int64  `json:"favorite_count"`
}

func TestFire(t *testing.T) {
	expectedText := "Some text"
	var expectedFavoriteCount int64 = 24
	client, server := mockServer(`{"text": "Some text", "favorite_count": 24}`)
	defer server.Close()

	sling := New().Client(client)
	req, _ := http.NewRequest("GET", server.URL, nil)
	var model FakeModel
	resp, err := sling.Fire(req, &model)

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

func TestFire_nilV(t *testing.T) {
	client, server := mockServer("")
	defer server.Close()

	sling := New().Client(client)
	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := sling.Fire(req, nil)

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
