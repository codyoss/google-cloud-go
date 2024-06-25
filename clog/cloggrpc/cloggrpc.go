package cloggrpc

import (
	"context"
	"log/slog"
	"strings"

	"cloud.google.com/go/clog/internal"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ProtoMessageRequest returns a lazily evaluated [slog.LogValuer] for
// the provided message. The context is used to extract outgoing headers.
func ProtoMessageRequest(ctx context.Context, msg proto.Message) slog.LogValuer {
	return &protoMessage{ctx: ctx, msg: msg}
}

// ProtoMessageResponse returns a lazily evaluated [slog.LogValuer] for
// the provided message.
func ProtoMessageResponse(msg proto.Message) slog.LogValuer {
	return &protoMessage{msg: msg}
}

type protoMessage struct {
	ctx context.Context
	msg proto.Message
}

func (m *protoMessage) LogValue() slog.Value {
	if m == nil || m.msg == nil {
		return slog.Value{}
	}

	var groupValueAtts []slog.Attr

	if m.ctx != nil && internal.IsDebugLoggingEnabled() {
		var headerAttr []slog.Attr
		if m, ok := metadata.FromOutgoingContext(m.ctx); ok {
			for k, v := range m {
				headerAttr = append(headerAttr, slog.String(k, internal.SensitiveString(strings.Join(v, ","))))
			}
		}
		groupValueAtts = append(groupValueAtts, slog.Any("headers", headerAttr))
	}
	b, _ := protojson.MarshalOptions{AllowPartial: true, UseEnumNumbers: true}.Marshal(m.msg)
	groupValueAtts = append(groupValueAtts, slog.String("payload", internal.SensitiveString(string(b))))
	return slog.GroupValue(groupValueAtts...)
}
