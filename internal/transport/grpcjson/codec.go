package grpcjson

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

type Codec struct{}

func (Codec) Name() string {
	return "json"
}

func (Codec) Marshal(value any) ([]byte, error) {
	return json.Marshal(value)
}

func (Codec) Unmarshal(data []byte, value any) error {
	return json.Unmarshal(data, value)
}

func init() {
	encoding.RegisterCodec(Codec{})
}
