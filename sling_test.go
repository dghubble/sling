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
	developerClient := &http.Client{}
	cases := []struct {
		input    *http.Client
		expected *http.Client
	}{
		{nil, http.DefaultClient},
		{developerClient, developerClient},
	}
	for _, c := range cases {
		sling := New(c.input)
		if sling.httpClient != c.expected {
			t.Errorf("expected %v, got %v", c.expected, sling.httpClient)
		}
	}
}

// i.e Sling.Request()
func TestCopy(t *testing.T) {
	cases := []*Sling{
		&Sling{httpClient: &http.Client{}, Method: "GET", BaseUrl: "http://example.com", PathUrl: "/path"},
		&Sling{httpClient: nil, Method: "", BaseUrl: "http://example.com", PathUrl: ""},
	}
	for _, sling := range cases {
		copy := sling.Request()
		if copy.httpClient != sling.httpClient {
			t.Errorf("expected %p, got %p", sling.httpClient, copy.httpClient)
		}
		if copy.Method != sling.Method {
			t.Errorf("expected %s, got %s", sling.Method, copy.Method)
		}
		if copy.BaseUrl != sling.BaseUrl {
			t.Errorf("expected %s, got %s", sling.BaseUrl, copy.BaseUrl)
		}
		if copy.PathUrl != sling.PathUrl {
			t.Errorf("expected %s, got %s", sling.PathUrl, copy.PathUrl)
		}
	}
}

func TestBaseSetter(t *testing.T) {
	cases := []string{"http://example.com", "", "/path", "path"}
	for _, base := range cases {
		sling := New(nil)
		sling.Base(base)
		if sling.BaseUrl != base {
			t.Errorf("expected %s, got %s", base, sling.Base)
		}
	}
}

func TestPathSetter(t *testing.T) {
	cases := []string{"http://example.com", "", "/path", "path"}
	for _, path := range cases {
		sling := New(nil)
		sling.Path(path)
		if sling.PathUrl != path {
			t.Errorf("expected %s, got %s", path, sling.PathUrl)
		}
	}
}

func TestMethodSetters(t *testing.T) {
	cases := []struct {
		sling          *Sling
		expectedMethod string
	}{
		{New(nil).Head("http://a.io"), HEAD},
		{New(nil).Get("http://a.io"), GET},
		{New(nil).Post("http://a.io"), POST},
		{New(nil).Put("http://a.io"), PUT},
		{New(nil).Patch("http://a.io"), PATCH},
		{New(nil).Delete("http://a.io"), DELETE},
	}
	for _, c := range cases {
		if c.sling.Method != c.expectedMethod {
			t.Errorf("expected method %s, got %s", c.expectedMethod, c.sling.Method)
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
		{New(nil).Base("http://a.io"), "", "http://a.io", nil},
		{New(nil).Path("http://a.io"), "", "http://a.io", nil},
		{New(nil).Get("http://a.io"), GET, "http://a.io", nil},
		{New(nil).Put("http://a.io"), PUT, "http://a.io", nil},
		{New(nil).Base("http://a.io").Path("/foo"), "", "http://a.io/foo", nil},
		{New(nil).Base("http://a.io").Post("/foo"), POST, "http://a.io/foo", nil},
		// if relative path is an absolute url, base is ignored
		{New(nil).Base("http://a.io").Path("http://b.io"), "", "http://b.io", nil},
		// last setter takes priority
		{New(nil).Base("http://a.io").Base("http://b.io"), "", "http://b.io", nil},
		{New(nil).Path("http://a.io").Path("http://b.io"), "", "http://b.io", nil},
		{New(nil).Post("http://a.io").Get("http://b.io"), GET, "http://b.io", nil},
		{New(nil).Get("http://b.io").Post("http://a.io"), POST, "http://a.io", nil},
		// removes extra '/' between base and ref url
		{New(nil).Base("http://a.io/").Get("/foo"), GET, "http://a.io/foo", nil},
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

	sling := New(client)
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

	sling := New(client)
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
