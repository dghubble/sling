package sling

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

// Decoder decodes http response bodies.  It is implemented by JSONDecoder.
// You can provide alternative decoders if you expect a different content type.
type Decoder interface {
	Decode(r io.Reader, v interface{}) error
}

// JSONDecoder is the default Decoder implementation. It uses
// the JSON decoder from the standard encoding/json package.
type JSONDecoder struct{}

// Decode reads the next value from the reader and stores it in the value pointed to by v.
func (d JSONDecoder) Decode(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

// assert this implements the Decoder interface
var _ Decoder = JSONDecoder{}

// JSONPBDecoder returns a decoder which can unmarshal JSON-encoded protobuf messages.
type JSONPBDecoder struct{}

// Decode reads the next value from the reader and stores it in the value pointed to by v.
func (d JSONPBDecoder) Decode(r io.Reader, v interface{}) error {
	if msg, ok := v.(proto.Message); ok {
		return jsonpb.Unmarshal(r, msg)
	}
	return fmt.Errorf("non-protobuf interface v given")
}

// assert this implements the Decoder interface
var _ Decoder = JSONPBDecoder{}
