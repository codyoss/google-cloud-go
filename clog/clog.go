package clog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// AuthSystemKey is the system key used for auth library logs.
	AuthSystemKey = "auth"
	// ClientSystemKey is the system key used for logs originating from client
	// libraries.
	ClientSystemKey = "client"
)

const (
	// determines if logging is enabled
	enabledEnvVar = "GOOGLE_SDK_DEBUG_LOGGING"
	// determines if sensitive logging is enabled
	enabledSensitiveEnvVar = "GOOGLE_SDK_DEBUG_LOGGING_SENSITIVE"
	// determines which systems
	systemsEnvVar = "GOOGLE_SDK_DEBUG_LOGGING_SYSTEMS"

	googLvlKey    = "severity"
	googMsgKey    = "message"
	googSourceKey = "sourceLocation"
	googTimeKey   = "timestamp"

	redactedValue = "[redacted]"
)

var (
	// configureLoggingOnce is set when the first [slog.Logger] is created from
	// calling [New] or when [SetDefaults] is called. This freezes all variables
	// below:
	configureLoggingOnce    sync.Once
	loggingEnabled          bool
	sensitiveLoggingEnabled bool
	handler                 slog.Handler
	systems                 map[string]bool
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
	// AddSource causes the handler to compute the source code position
	// of the log statement and appends it to each log event. Defaults to
	// false.
	AddSource bool
}

// SetDefaults configures all logging that originates from this package. This
// function must be called before any logger are instantiated with [New].
// This function may be called only once, calling it subsequent times will have
// no effect.
func SetDefaults(opts *DefaultOptions) {
	configureLoggingOnce.Do(func() {
		// Set Logger Defaults
		if opts == nil {
			opts = &DefaultOptions{}
		}
		level := opts.Level
		writer := opts.Writer
		handler = opts.Handler

		if level == nil {
			level = slog.LevelDebug
		}
		if writer == nil {
			writer = os.Stderr
		}
		if handler == nil {
			handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
				AddSource:   opts.AddSource,
				Level:       level,
				ReplaceAttr: replaceAttr,
			})
		}

		// Parse environment variables
		loggingEnabled, _ = strconv.ParseBool(os.Getenv(enabledEnvVar))
		sensitiveLoggingEnabled, _ = strconv.ParseBool(os.Getenv(enabledSensitiveEnvVar))
		ss := strings.Split(strings.TrimSpace(os.Getenv(systemsEnvVar)), ",")
		systems = make(map[string]bool, len(ss))
		for _, s := range ss {
			cleanString := strings.ToLower(strings.TrimSpace(s))
			if cleanString == "" {
				continue
			}
			systems[strings.ToLower(strings.TrimSpace(s))] = true
		}
		if len(systems) == 0 {
			// enable client and auth by default if not configured
			systems[ClientSystemKey] = true
			systems[AuthSystemKey] = true
		}
	})
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
	h := gcHandler{
		system: opts.System,
		h:      handler,
	}
	return slog.New(h)
}

// HTTPRequest returns a lazily evaluated [slog.LogValuer] for a [http.Request].
func HTTPRequest(req *http.Request, payload []byte) slog.LogValuer {
	if req == nil {
		return lazyLogValuer[*http.Request]{}
	}
	return lazyLogValuer[*request]{Value: &request{
		req:     req,
		payload: payload,
	}}
}

// HTTPResponse returns a lazily evaluated [slog.LogValuer] for a
// [http.Response].
func HTTPResponse(resp *http.Response, payload []byte) slog.LogValuer {
	if resp == nil {
		return lazyLogValuer[*http.Request]{}
	}
	return lazyLogValuer[*response]{Value: &response{
		resp:    resp,
		payload: payload,
	}}
}

// SensitiveString returns a string that is only logged if
// GOOGLE_SDK_DEBUG_LOGGING_SENSITIVE and GOOGLE_SDK_DEBUG_LOGGING are both
// true. If not, the returned string is `[redacted]`.
func SensitiveString(s string) string {
	if !sensitiveLoggingEnabled {
		return redactedValue
	}
	return s
}

type request struct {
	req     *http.Request
	payload []byte
}

type response struct {
	resp    *http.Response
	payload []byte
}

type lazyLogValuer[T any] struct {
	Value T
}

func (l lazyLogValuer[T]) LogValue() slog.Value {
	switch v := any(l.Value).(type) {
	case *request:
		if v == nil || v.req == nil {
			return slog.Value{}
		}
		var headerAttr []slog.Attr
		for k, val := range v.req.Header {
			headerAttr = append(headerAttr, slog.String(k, SensitiveString(strings.Join(val, ","))))
		}
		buf := &bytes.Buffer{}
		json.Compact(buf, v.payload)
		return slog.GroupValue(
			slog.String("method", v.req.Method),
			slog.String("url", SensitiveString(v.req.URL.String())),
			slog.Any("headers", headerAttr),
			slog.String("payload", SensitiveString(buf.String())),
		)
	case *response:
		if v == nil {
			return slog.Value{}
		}
		var headerAttr []slog.Attr
		for k, v := range v.resp.Header {
			headerAttr = append(headerAttr, slog.String(k, SensitiveString(strings.Join(v, ","))))
		}
		buf := &bytes.Buffer{}
		json.Compact(buf, v.payload)
		return slog.GroupValue(
			slog.String("status", fmt.Sprint(v.resp.StatusCode)),
			slog.Any("headers", headerAttr),
			slog.String("payload", SensitiveString(buf.String())),
		)
	default:
		return slog.Value{}
	}
}

type gcHandler struct {
	system string
	h      slog.Handler
}

// Enabled determines if logging should be enabled in the Go Cloud SDK by checking
// if:
//   - GOOGLE_SDK_DEBUG_LOGGING` is true
//   - the log level should be logged
//   - the system is configured to log
func (g gcHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return loggingEnabled && g.h.Enabled(ctx, lvl) && systems[g.system]
}

func (g gcHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(slog.String("system", g.system))
	return g.h.Handle(ctx, r)
}

func (g gcHandler) WithAttrs(a []slog.Attr) slog.Handler { return g.h.WithAttrs(a) }

func (g gcHandler) WithGroup(name string) slog.Handler { return g.h.WithGroup(name) }

// replaceAttr remaps default Go logging keys to match what is expected in
// cloud logging.
func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if groups == nil {
		if a.Key == slog.LevelKey {
			a.Key = googLvlKey
			return a
		} else if a.Key == slog.MessageKey {
			a.Key = googMsgKey
			return a
		} else if a.Key == slog.SourceKey {
			a.Key = googSourceKey
			return a
		} else if a.Key == slog.TimeKey {
			a.Key = googTimeKey
			if a.Value.Kind() == slog.KindTime {
				a.Value = slog.StringValue(a.Value.Time().Format(time.RFC3339))
			}
			return a
		}
	}
	return a
}
