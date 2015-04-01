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
