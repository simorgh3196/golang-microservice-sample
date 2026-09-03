package connectutil

import (
	"errors"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// NewJSONCodec はゼロ値（falseや空文字）も省略せずに出力する JSON コーデックを生成します。
func NewJSONCodec() connect.Codec {
	return &protoJSONCodec{
		name: "json",
		marshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true, // ゼロ値(false, "", 0など)もキーを省略せず出力する
		},
		unmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true, // 道のフィールドがあっても無視する
		},
	}
}

type protoJSONCodec struct {
	name             string
	marshalOptions   protojson.MarshalOptions
	unmarshalOptions protojson.UnmarshalOptions
}

func (c *protoJSONCodec) Name() string {
	return c.name
}

func (c *protoJSONCodec) Marshal(message any) ([]byte, error) {
	msg, ok := message.(proto.Message)
	if !ok {
		return nil, errors.New("message does not implement proto.Message")
	}
	return c.marshalOptions.Marshal(msg)
}

func (c *protoJSONCodec) Unmarshal(data []byte, message any) error {
	msg, ok := message.(proto.Message)
	if !ok {
		return errors.New("message does not implement proto.Message")
	}
	return c.unmarshalOptions.Unmarshal(data, msg)
}
