package sling

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ResponseDecoder decodes http responses into struct values.
type ResponseDecoder interface {
	// Decode decodes the response into the value pointed to by v.
	Decode(resp *http.Response, v interface{}) error
}

// jsonDecoder decodes http response JSON into a JSON-tagged struct value.
type jsonDecoder struct{}

// Decode decodes the Response Body into the value pointed to by v.
// Caller must provide a non-nil v and close the resp.Body.
func (d jsonDecoder) Decode(resp *http.Response, v interface{}) error {
	return json.NewDecoder(resp.Body).Decode(v)
}

// ByteStreamer is a [sling.ResponseDecoder] which simply forwards response data 'as is' rather than trying to deocde
// it. This is useful when 1/ response is actually just plain text (like 5XX response from API gateways). 2/ response
// data is a byte stream representing some file or a binary blob.
//
// It leverages existing facilities of automatic discarding and closing of response body so the user does not need to
// care about it.
type ByteStreamer struct{}

// Decode simply tries to copy response data into v assuming its an [io.Writer] instance. Assuming so little about v
// gives consumers a lot of choice about consuming response data. They can wait for all data to be dumped into some
// buffer then act on it or they can read as soon as data gets written.
func (d ByteStreamer) Decode(resp *http.Response, v any) error {
	var w io.Writer
	w, ok := v.(io.Writer)
	if !ok {
		return fmt.Errorf("expected type: %T; got: %T", w, v)
	}

	_, err := io.Copy(w, resp.Body)
	if err != nil {
		return fmt.Errorf("failed copying response data to v: %w", err)
	}

	return nil
}
