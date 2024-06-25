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
	// [slog.LevelInfo]
	Level slog.Leveler
	// Writer configures where logs are written to. Defaults to [os.Stderr].
	Writer io.Writer
	// Handler configure the underlying handler used to format the logs. If
	// specified all other options are ignored. Defaults to a
	// [slog.JSONHandler].
	Handler slog.Handler
	// EnableLogging turns on logging.
	EnableLogging bool
}

func (o *DefaultOptions) toInternal() *internal.DefaultOptions {
	io := &internal.DefaultOptions{}
	if o == nil {
		return io
	}
	io.Level = o.Level
	io.Writer = o.Writer
	io.Handler = o.Handler
	io.EnableLogging = o.EnableLogging
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
func HTTPRequest(req *http.Request, body []byte) slog.LogValuer {
	return &request{
		req:     req,
		payload: body,
	}
}

// HTTPResponse returns a lazily evaluated [slog.LogValuer] for a
// [http.Response].
func HTTPResponse(resp *http.Response, body []byte) slog.LogValuer {
	return &response{
		resp:    resp,
		payload: body,
	}
}

// SensitiveString returns a string that is only logged if:
//   - GOOGLE_SDK_DEBUG_LOGGING is true
//   - log level is at or below [slog.LevelDebug]
//
// If not, the returned string is `[redacted]`.
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
	var groupValueAtts []slog.Attr
	groupValueAtts = append(groupValueAtts, slog.String("method", r.req.Method))

	if internal.IsDebugLoggingEnabled() {
		groupValueAtts = append(groupValueAtts, slog.String("url", internal.SensitiveString(r.req.URL.String())))

		var headerAttr []slog.Attr
		for k, val := range r.req.Header {
			headerAttr = append(headerAttr, slog.String(k, internal.SensitiveString(strings.Join(val, ","))))
		}
		groupValueAtts = append(groupValueAtts, slog.Any("headers", headerAttr))

		buf := &bytes.Buffer{}
		json.Compact(buf, r.payload)
		groupValueAtts = append(groupValueAtts, slog.String("payload", internal.SensitiveString(buf.String())))
	}
	return slog.GroupValue(groupValueAtts...)
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
		headerAttr = append(headerAttr, slog.String(k, internal.SensitiveString(strings.Join(v, ","))))
	}
	buf := &bytes.Buffer{}
	if err := json.Compact(buf, r.payload); err != nil {
		buf.Reset()
		buf.Write(r.payload)
	}
	return slog.GroupValue(
		slog.String("status", fmt.Sprint(r.resp.StatusCode)),
		slog.Any("headers", headerAttr),
		slog.String("payload", internal.SensitiveString(buf.String())),
	)
}

// DynamicLevel returns the level things should be logged at in client libraries.
// This is only meant to be used when using logging helpers like [HTTPRequest]
// as they redact certain info at certain levels.
func DynamicLevel() slog.Level {
	return internal.DynamicLevel()
}
