package sling

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

// ResponseDecoder decodes http responses into struct values.
type ResponseDecoder interface {
	// Decode decodes the response into the value pointed to by v.
	Decode(resp *http.Response, v interface{}) error
}

// jsonResponseDecoder decodes http response JSON into a JSON-tagged struct value.
type jsonResponseDecoder struct {
}

// Decode decodes the Response Body into the value pointed to by v.
// Caller must provide a non-nil v and close the resp.Body.
func (d jsonResponseDecoder) Decode(resp *http.Response, v interface{}) error {
	return json.NewDecoder(resp.Body).Decode(v)
}

// xmlDecoder decodes http response XML into a XML-tagged struct value.
type xmlResponseDecoder struct{}

// Decode decodes the Response Body into the value pointed to by v.
// Caller must provide a non-nil v and close the resp.Body.
func (d xmlResponseDecoder) Decode(resp *http.Response, v interface{}) error {
	return xml.NewDecoder(resp.Body).Decode(v)
}
