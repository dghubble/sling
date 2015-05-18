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
	"reflect"
	"testing"
)

type FakeParams struct {
	KindName string `url:"kind_name"`
	Count    int    `url:"count"`
}

// Url-tagged query struct
var paramsA = struct {
	Limit int `url:"limit"`
}{
	30,
}
var paramsB = FakeParams{KindName: "recent", Count: 25}

// Json-tagged model struct
type FakeModel struct {
	Text          string  `json:"text,omitempty"`
	FavoriteCount int64   `json:"favorite_count,omitempty"`
	Temperature   float64 `json:"temperature,omitempty"`
}

var modelA = FakeModel{Text: "note", FavoriteCount: 12}

func TestNew(t *testing.T) {
	sling := New()
	if sling.HttpClient != http.DefaultClient {
		t.Errorf("expected %v, got %v", http.DefaultClient, sling.HttpClient)
	}
	if sling.Header == nil {
		t.Errorf("Header map not initialized with make")
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
		New().Add("Content-Type", "application/json"),
		New().Add("A", "B").Add("a", "c").New(),
		New().Add("A", "B").New().Add("a", "c"),
		New().BodyStruct(paramsB),
		New().BodyStruct(paramsB).New(),
	}
	for _, sling := range cases {
		child := sling.New()
		if child.HttpClient != sling.HttpClient {
			t.Errorf("expected %p, got %p", sling.HttpClient, child.HttpClient)
		}
		if child.Method != sling.Method {
			t.Errorf("expected %s, got %s", sling.Method, child.Method)
		}
		if child.RawUrl != sling.RawUrl {
			t.Errorf("expected %s, got %s", sling.RawUrl, child.RawUrl)
		}
		// Header should be a copy of parent Sling Header. For example, calling
		// baseSling.Add("k","v") should not mutate previously created child Slings
		if sling.Header != nil {
			// struct literal cases don't init Header in usual way, skip Header check
			if !reflect.DeepEqual(sling.Header, child.Header) {
				t.Errorf("not DeepEqual: expected %v, got %v", sling.Header, child.Header)
			}
			sling.Header.Add("K", "V")
			if child.Header.Get("K") != "" {
				t.Errorf("child.Header was a reference to original map, should be copy")
			}
		}
		// queryStruct slice should be a new slice with a copy of the contents
		if len(sling.queryStructs) > 0 {
			// mutating one slice should not mutate the other
			child.queryStructs[0] = nil
			if sling.queryStructs[0] == nil {
				t.Errorf("child.queryStructs was a re-slice, expected slice with copied contents")
			}
		}
		// jsonBody should be copied
		if child.jsonBody != sling.jsonBody {
			t.Errorf("expected %v, got %v", sling.jsonBody, child.jsonBody)
		}
		// bodyStruct should be copied
		if child.bodyStruct != sling.bodyStruct {
			t.Errorf("expected %v, got %v", sling.bodyStruct, child.bodyStruct)
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
		rawURL         string
		path           string
		expectedRawURL string
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
		sling := New().Base(c.rawURL).Path(c.path)
		if sling.RawUrl != c.expectedRawURL {
			t.Errorf("expected %s, got %s", c.expectedRawURL, sling.RawUrl)
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

func TestAddHeader(t *testing.T) {
	cases := []struct {
		sling          *Sling
		expectedHeader map[string][]string
	}{
		{New().Add("authorization", "OAuth key=\"value\""), map[string][]string{"Authorization": []string{"OAuth key=\"value\""}}},
		// header keys should be canonicalized
		{New().Add("content-tYPE", "application/json").Add("User-AGENT", "sling"), map[string][]string{"Content-Type": []string{"application/json"}, "User-Agent": []string{"sling"}}},
		// values for existing keys should be appended
		{New().Add("A", "B").Add("a", "c"), map[string][]string{"A": []string{"B", "c"}}},
		// Add should add to values for keys added by parent Slings
		{New().Add("A", "B").Add("a", "c").New(), map[string][]string{"A": []string{"B", "c"}}},
		{New().Add("A", "B").New().Add("a", "c"), map[string][]string{"A": []string{"B", "c"}}},
	}
	for _, c := range cases {
		// type conversion from Header to alias'd map for deep equality comparison
		headerMap := map[string][]string(c.sling.Header)
		if !reflect.DeepEqual(c.expectedHeader, headerMap) {
			t.Errorf("not DeepEqual: expected %v, got %v", c.expectedHeader, headerMap)
		}
	}
}

func TestSetHeader(t *testing.T) {
	cases := []struct {
		sling          *Sling
		expectedHeader map[string][]string
	}{
		// should replace existing values associated with key
		{New().Add("A", "B").Set("a", "c"), map[string][]string{"A": []string{"c"}}},
		{New().Set("content-type", "A").Set("Content-Type", "B"), map[string][]string{"Content-Type": []string{"B"}}},
		// Set should replace values received by copying parent Slings
		{New().Set("A", "B").Add("a", "c").New(), map[string][]string{"A": []string{"B", "c"}}},
		{New().Add("A", "B").New().Set("a", "c"), map[string][]string{"A": []string{"c"}}},
	}
	for _, c := range cases {
		// type conversion from Header to alias'd map for deep equality comparison
		headerMap := map[string][]string(c.sling.Header)
		if !reflect.DeepEqual(c.expectedHeader, headerMap) {
			t.Errorf("not DeepEqual: expected %v, got %v", c.expectedHeader, headerMap)
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
		// json tagged struct is set as jsonBody
		{nil, fakeModel, fakeModel},
		// nil argument to jsonBody does not replace existing jsonBody
		{fakeModel, nil, fakeModel},
		// nil jsonBody remains nil
		{nil, nil, nil},
	}
	for _, c := range cases {
		sling := New()
		sling.jsonBody = c.initial
		sling.JsonBody(c.input)
		if sling.jsonBody != c.expected {
			t.Errorf("expected %v, got %v", c.expected, sling.jsonBody)
		}
		// Header Content-Type should be application/json if jsonBody arg was non-nil
		if c.input != nil && sling.Header.Get(contentType) != jsonContentType {
			t.Errorf("Incorrect or missing header, expected %s, got %s", jsonContentType, sling.Header.Get(contentType))
		} else if c.input == nil && sling.Header.Get(contentType) != "" {
			t.Errorf("did not expect a Content-Type header, got %s", sling.Header.Get(contentType))
		}
	}
}

func TestBodyStructSetter(t *testing.T) {
	cases := []struct {
		initial  interface{}
		input    interface{}
		expected interface{}
	}{
		// url tagged struct is set as bodyStruct
		{nil, paramsB, paramsB},
		// nil argument to bodyStruct does not replace existing bodyStruct
		{paramsB, nil, paramsB},
		// nil bodyStruct remains nil
		{nil, nil, nil},
	}
	for _, c := range cases {
		sling := New()
		sling.bodyStruct = c.initial
		sling.BodyStruct(c.input)
		if sling.bodyStruct != c.expected {
			t.Errorf("expected %v, got %v", c.expected, sling.bodyStruct)
		}
		// Content-Type should be application/x-www-form-urlencoded if bodyStruct was non-nil
		if c.input != nil && sling.Header.Get(contentType) != formContentType {
			t.Errorf("Incorrect or missing header, expected %s, got %s", formContentType, sling.Header.Get(contentType))
		} else if c.input == nil && sling.Header.Get(contentType) != "" {
			t.Errorf("did not expect a Content-Type header, got %s", sling.Header.Get(contentType))
		}
	}

}

func TestRequest_urlAndMethod(t *testing.T) {
	cases := []struct {
		sling          *Sling
		expectedMethod string
		expectedURL    string
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
		if req.URL.String() != c.expectedURL {
			t.Errorf("expected url %s, got %s for %+v", c.expectedURL, req.URL.String(), c.sling)
		}
		if req.Method != c.expectedMethod {
			t.Errorf("expected method %s, got %s for %+v", c.expectedMethod, req.Method, c.sling)
		}
	}
}

func TestRequest_queryStructs(t *testing.T) {
	cases := []struct {
		sling       *Sling
		expectedURL string
	}{
		{New().Base("http://a.io").QueryStruct(paramsA), "http://a.io?limit=30"},
		{New().Base("http://a.io").QueryStruct(paramsA).QueryStruct(paramsB), "http://a.io?count=25&kind_name=recent&limit=30"},
		{New().Base("http://a.io/").Path("foo?path=yes").QueryStruct(paramsA), "http://a.io/foo?limit=30&path=yes"},
		{New().Base("http://a.io").QueryStruct(paramsA).New(), "http://a.io?limit=30"},
		{New().Base("http://a.io").QueryStruct(paramsA).New().QueryStruct(paramsB), "http://a.io?count=25&kind_name=recent&limit=30"},
	}
	for _, c := range cases {
		req, _ := c.sling.Request()
		if req.URL.String() != c.expectedURL {
			t.Errorf("expected url %s, got %s for %+v", c.expectedURL, req.URL.String(), c.sling)
		}
	}
}

func TestRequest_body(t *testing.T) {
	cases := []struct {
		sling               *Sling
		expectedBody        string // expected Body io.Reader as a string
		expectedContentType string
	}{
		// JsonBody
		{New().JsonBody(modelA), "{\"text\":\"note\",\"favorite_count\":12}\n", jsonContentType},
		{New().JsonBody(&modelA), "{\"text\":\"note\",\"favorite_count\":12}\n", jsonContentType},
		{New().JsonBody(&FakeModel{}), "{}\n", jsonContentType},
		{New().JsonBody(FakeModel{}), "{}\n", jsonContentType},
		// JsonBody overrides existing values
		{New().JsonBody(&FakeModel{}).JsonBody(&FakeModel{Text: "msg"}), "{\"text\":\"msg\"}\n", jsonContentType},
		// BodyStruct (form)
		{New().BodyStruct(paramsA), "limit=30", formContentType},
		{New().BodyStruct(paramsB), "count=25&kind_name=recent", formContentType},
		{New().BodyStruct(&paramsB), "count=25&kind_name=recent", formContentType},
		// BodyStruct overrides existing values
		{New().BodyStruct(paramsA).New().BodyStruct(paramsB), "count=25&kind_name=recent", formContentType},
		// Mixture of JsonBody and BodyStruct prefers body setter called last with a non-nil argument
		{New().BodyStruct(paramsB).New().JsonBody(modelA), "{\"text\":\"note\",\"favorite_count\":12}\n", jsonContentType},
		{New().JsonBody(modelA).New().BodyStruct(paramsB), "count=25&kind_name=recent", formContentType},
		{New().BodyStruct(paramsB).New().JsonBody(nil), "count=25&kind_name=recent", formContentType},
		{New().JsonBody(modelA).New().BodyStruct(nil), "{\"text\":\"note\",\"favorite_count\":12}\n", jsonContentType},
	}
	for _, c := range cases {
		req, _ := c.sling.Request()
		buf := new(bytes.Buffer)
		buf.ReadFrom(req.Body)
		// req.Body should have contained the expectedBody string
		if value := buf.String(); value != c.expectedBody {
			t.Errorf("expected Request.Body %s, got %s", c.expectedBody, value)
		}
		// Header Content-Type should be application/json
		if actualHeader := req.Header.Get(contentType); actualHeader != c.expectedContentType {
			t.Errorf("Incorrect or missing header, expected %s, got %s", c.expectedContentType, actualHeader)
		}
	}
}

func TestRequest_bodyNoData(t *testing.T) {
	// test that Body is left nil when no jsonBody or bodyStruct set
	slings := []*Sling{
		New(),
		New().JsonBody(nil),
		New().BodyStruct(nil),
	}
	for _, sling := range slings {
		req, _ := sling.Request()
		if req.Body != nil {
			t.Errorf("expected nil Request.Body, got %v", req.Body)
		}
		// Header Content-Type should not be set when jsonBody argument was nil or never called
		if actualHeader := req.Header.Get(contentType); actualHeader != "" {
			t.Errorf("did not expect a Content-Type header, got %s", actualHeader)
		}
	}
}

func TestRequest_bodyEncodeErrors(t *testing.T) {
	cases := []struct {
		sling       *Sling
		expectedErr error
	}{
		// check that Encode errors are propagated, illegal JSON field
		{New().JsonBody(FakeModel{Temperature: math.Inf(1)}), errors.New("json: unsupported value: +Inf")},
	}
	for _, c := range cases {
		req, err := c.sling.Request()
		if err == nil || err.Error() != c.expectedErr.Error() {
			t.Errorf("expected error %v, got %v", c.expectedErr, err)
		}
		if req != nil {
			t.Errorf("expected nil Request, got %+v", req)
		}
	}
}

func TestRequest_headers(t *testing.T) {
	cases := []struct {
		sling          *Sling
		expectedHeader map[string][]string
	}{
		{New().Add("authorization", "OAuth key=\"value\""), map[string][]string{"Authorization": []string{"OAuth key=\"value\""}}},
		// header keys should be canonicalized
		{New().Add("content-tYPE", "application/json").Add("User-AGENT", "sling"), map[string][]string{"Content-Type": []string{"application/json"}, "User-Agent": []string{"sling"}}},
		// values for existing keys should be appended
		{New().Add("A", "B").Add("a", "c"), map[string][]string{"A": []string{"B", "c"}}},
		// Add should add to values for keys added by parent Slings
		{New().Add("A", "B").Add("a", "c").New(), map[string][]string{"A": []string{"B", "c"}}},
		{New().Add("A", "B").New().Add("a", "c"), map[string][]string{"A": []string{"B", "c"}}},
		// Add and Set
		{New().Add("A", "B").Set("a", "c"), map[string][]string{"A": []string{"c"}}},
		{New().Set("content-type", "A").Set("Content-Type", "B"), map[string][]string{"Content-Type": []string{"B"}}},
		// Set should replace values received by copying parent Slings
		{New().Set("A", "B").Add("a", "c").New(), map[string][]string{"A": []string{"B", "c"}}},
		{New().Add("A", "B").New().Set("a", "c"), map[string][]string{"A": []string{"c"}}},
	}
	for _, c := range cases {
		req, _ := c.sling.Request()
		// type conversion from Header to alias'd map for deep equality comparison
		headerMap := map[string][]string(req.Header)
		if !reflect.DeepEqual(c.expectedHeader, headerMap) {
			t.Errorf("not DeepEqual: expected %v, got %v", c.expectedHeader, headerMap)
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

func TestDo(t *testing.T) {
	expectedText := "Some text"
	var expectedFavoriteCount int64 = 24
	client, server := mockServer(`{"text":"Some text","favorite_count":24}`)
	defer server.Close()

	sling := New().Client(client)
	req, _ := http.NewRequest("GET", server.URL, nil)
	var model FakeModel
	resp, err := sling.Do(req, &model, nil)

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
	resp, err := sling.Do(req, nil, nil)

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
