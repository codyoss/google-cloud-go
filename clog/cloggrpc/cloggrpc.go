package cloggrpc

import (
	"log/slog"

	"cloud.google.com/go/clog/internal"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func ProtoMessage(msg proto.Message) slog.LogValuer {
	return &protoMessage{msg: msg}
}

type protoMessage struct {
	msg proto.Message
}

func (m *protoMessage) LogValue() slog.Value {
	if m == nil || m.msg == nil {
		return slog.Value{}
	}
	b, err := protojson.MarshalOptions{AllowPartial: true, UseEnumNumbers: true}.Marshal(m.msg)
	if err != nil {
		return slog.Value{}
	}
	return slog.StringValue(internal.SensitiveString(string(b)))
}
