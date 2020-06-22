package sling

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

// ResponseDecoder decodes http responses into struct values.
type ResponseDecoder interface {
	// Decode decodes the response into the value pointed to by v.
	Decode(resp *http.Response, v interface{}) error
}

// jsonDecoder decodes http response JSON into a JSON-tagged struct value.
type jsonDecoder struct {
}

// Decode decodes the Response Body into the value pointed to by v.
// Caller must provide a non-nil v and close the resp.Body.
func (d jsonDecoder) Decode(resp *http.Response, v interface{}) error {
	return json.NewDecoder(resp.Body).Decode(v)
}

// byteDecoder decodes http response into a byte slice.
type byteDecoder struct {
}

// Decode decodes the Response Body into a byte slice.
// Caller must provide a non-nil v and close the resp.Body.
func (d byteDecoder) Decode(resp *http.Response, v interface{}) error {
	vBytes, ok := v.(*[]byte)
	if !ok {
		return errors.New("bad v type, must be *[]byte")
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	*vBytes = bytes
	return err
}
