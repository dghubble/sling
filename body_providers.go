package sling

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	goquery "github.com/google/go-querystring/query"
)

// BodyProvider creates request bodies. Slings comes with a couple of providers (form, json, ...)
// and you can also implement yours to handle custom body types.
type BodyProvider interface {
	// ContentType controls the request's content type header
	ContentType() string

	// Body creates the io.Reader that will be used as the request's body
	Body() (io.Reader, error)
}

// ReaderBody creates a BodyProvider from an io.Reader
func ReaderBody(body io.Reader) BodyProvider {
	return readerBodyProvider{reader: body}
}

// JSONBody creates a BodyProvider that encodes the provided value as JSON.
// The bodyJSON argument should be a pointer to a JSON tagged struct. See
// https://golang.org/pkg/encoding/json/#MarshalIndent for details.
func JSONBody(bodyJSON interface{}) BodyProvider {
	if bodyJSON == nil {
		return nil
	}

	return jsonBodyProvider{payload: bodyJSON}
}

// FormBody creates a BodyProvider that encodes the provided value as URL form encoded body.
// The bodyForm argument should be a pointer to a url tagged struct. See
// https://godoc.org/github.com/google/go-querystring/query for details.
func FormBody(bodyForm interface{}) BodyProvider {
	if bodyForm == nil {
		return nil
	}

	return formBodyProvider{payload: bodyForm}
}

// Implementations

// JSON

type jsonBodyProvider struct {
	payload interface{}
}

func (p jsonBodyProvider) ContentType() string {
	return jsonContentType
}

func (p jsonBodyProvider) Body() (io.Reader, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(p.payload)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// Form
type formBodyProvider struct {
	payload interface{}
}

func (p formBodyProvider) ContentType() string {
	return formContentType
}

func (p formBodyProvider) Body() (io.Reader, error) {
	values, err := goquery.Values(p.payload)
	if err != nil {
		return nil, err
	}
	return strings.NewReader(values.Encode()), nil
}

// Reader

type readerBodyProvider struct {
	reader io.Reader
}

func (p readerBodyProvider) ContentType() string {
	return ""
}

func (p readerBodyProvider) Body() (io.Reader, error) {
	return p.reader, nil
}
