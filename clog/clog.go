package clog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"cloud.google.com/go/clog/internal"
)

const (
	// AuthSystemKey is the system key used for auth library logs.
	AuthSystemKey = "auth"
	// ClientSystemKey is the system key used for logs originating from client
	// libraries.
	ClientSystemKey = "client"
)

// DefaultOptions used to configure global log settings.
type DefaultOptions struct {
	// Level configures what the default log level is. Defaults to
	// [slog.LevelDebug]
	Level slog.Leveler
	// Writer configures where logs are written to. Defaults to [os.Stderr].
	Writer io.Writer
	// Handler configure the underlying handler used to format the logs. If
	// specified all other options are ignored. Defaults to a
	// [slog.JSONHandler].
	Handler slog.Handler
	// EnableSourceInfo causes the handler to compute the source code position
	// of the log statement and appends it to each log event. Defaults to
	// false.
	EnableSourceInfo bool
	// EnableLogging turns on logging.
	EnableLogging bool
	// EnableSensitiveLogging turns on logging of more sensitive data.
	EnableSensitiveLogging bool
}

func (o *DefaultOptions) toInternal() *internal.DefaultOptions {
	io := &internal.DefaultOptions{}
	if o == nil {
		return io
	}
	io.Level = o.Level
	io.Writer = o.Writer
	io.Handler = o.Handler
	io.EnableSourceInfo = o.EnableSourceInfo
	io.EnableLogging = o.EnableLogging
	io.EnableSensitiveLogging = o.EnableSensitiveLogging
	return io
}

// SetDefaults configures all logging that originates from this package. This
// function must be called before any logger are instantiated with [New].
// This function may be called only once, calling it subsequent times will have
// no effect.
func SetDefaults(opts *DefaultOptions) {
	internal.SetDefaults(opts.toInternal())
}

// Options used to configure [slog.Logger] returned by [New].
type Options struct {
	// System the logger is associated with. Defaults to [ClientSystemKey].
	System string
}

// New returns a new [slog.Logger] configured with the provided [Options]. The
// returned logger will be a noop logger unless the environment variable
// GOOGLE_SDK_DEBUG_LOGGING is set to true. See package documentation for more
// details.
func New(opts *Options) *slog.Logger {
	// configures package defaults
	SetDefaults(nil)
	// configure logger defaults if not provided
	if opts == nil {
		opts = &Options{
			System: ClientSystemKey,
		}
	}
	sys := opts.System
	if sys == "" {
		sys = ClientSystemKey
	}
	return slog.New(internal.NewHandler(sys))
}

// HTTPRequest returns a lazily evaluated [slog.LogValuer] for a [http.Request].
func HTTPRequest(req *http.Request, payload []byte) slog.LogValuer {
	return &request{
		req:     req,
		payload: payload,
	}
}

// HTTPResponse returns a lazily evaluated [slog.LogValuer] for a
// [http.Response].
func HTTPResponse(resp *http.Response, payload []byte) slog.LogValuer {
	return &response{
		resp:    resp,
		payload: payload,
	}
}

// SensitiveString returns a string that is only logged if
// GOOGLE_SDK_DEBUG_LOGGING_SENSITIVE and GOOGLE_SDK_DEBUG_LOGGING are both
// true. If not, the returned string is `[redacted]`.
func SensitiveString(s string) string {
	return internal.SensitiveString(s)
}

type request struct {
	req     *http.Request
	payload []byte
}

func (r *request) LogValue() slog.Value {
	if r == nil || r.req == nil {
		return slog.Value{}
	}
	var headerAttr []slog.Attr
	for k, val := range r.req.Header {
		headerAttr = append(headerAttr, slog.String(k, SensitiveString(strings.Join(val, ","))))
	}
	buf := &bytes.Buffer{}
	json.Compact(buf, r.payload)
	return slog.GroupValue(
		slog.String("method", r.req.Method),
		slog.String("url", SensitiveString(r.req.URL.String())),
		slog.Any("headers", headerAttr),
		slog.String("payload", SensitiveString(buf.String())),
	)
}

type response struct {
	resp    *http.Response
	payload []byte
}

func (r *response) LogValue() slog.Value {
	if r == nil {
		return slog.Value{}
	}
	var headerAttr []slog.Attr
	for k, v := range r.resp.Header {
		headerAttr = append(headerAttr, slog.String(k, SensitiveString(strings.Join(v, ","))))
	}
	buf := &bytes.Buffer{}
	if err := json.Compact(buf, r.payload); err != nil {
		buf.Reset()
		buf.Write(r.payload)
	}
	return slog.GroupValue(
		slog.String("status", fmt.Sprint(r.resp.StatusCode)),
		slog.Any("headers", headerAttr),
		slog.String("payload", SensitiveString(buf.String())),
	)
}
