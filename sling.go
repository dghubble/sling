package sling

import (
	"encoding/json"
	"net/http"
)

type Sling struct {
	// http Client for doing requests
	httpClient *http.Client
}

// New returns a new Sling.
func New(httpClient *http.Client) *Sling {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Sling{httpClient: httpClient}
}

// Fire sends the HTTP request and decodes the response into the value pointed
// to by 'success'. It wraps http.Client.Do, but handles closing the Response
// Body. The Response and any error making the request are returned.
//
// Note that non-2xx StatusCodes are valid responses, not errors.
func (s *Sling) Fire(req *http.Request, value interface{}) (*http.Response, error) {
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return resp, err
	}
	// when err is nil, resp contains a non-nil resp.Body which must be closed
	defer resp.Body.Close()
	err = decodeResponse(resp, value)
	return resp, err
}

// decodeResponse decodes Response Body encoded as JSON into the value pointed
// to by v. Caller should call resp.Body.Close().
func decodeResponse(resp *http.Response, v interface{}) error {
	return json.NewDecoder(resp.Body).Decode(v)
}
