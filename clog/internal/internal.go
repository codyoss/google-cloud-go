package internal

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// determines if logging is enabled
	enabledEnvVar = "GOOGLE_SDK_DEBUG_LOGGING"
	// determines which systems
	systemsEnvVar = "GOOGLE_SDK_DEBUG_LOGGING_SYSTEMS"
	// logging level
	levelEnvVar = "GOOGLE_SDK_DEBUG_LOGGING_LEVEL"

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
	lvl                     slog.Level
	loggingEnabled          bool
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
	// EnableLogging turns on logging.
	EnableLogging bool
}

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
			sLevel := os.Getenv(levelEnvVar)
			switch sLevel {
			case "debug":
				level = slog.LevelDebug
			case "info":
				level = slog.LevelInfo
			case "warn":
				level = slog.LevelWarn
			case "error":
				level = slog.LevelError
			default:
				level = slog.LevelInfo
			}
		}

		if writer == nil {
			writer = os.Stderr
		}
		if handler == nil {
			handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
				AddSource:   true,
				Level:       level,
				ReplaceAttr: replaceAttr,
			})
		}

		// Parse environment variables
		loggingEnabled, _ = strconv.ParseBool(os.Getenv(enabledEnvVar))
		// Also honor code settings
		loggingEnabled = loggingEnabled || opts.EnableLogging

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
			systems["client"] = true
			systems["auth"] = true
		}
		lvl = level.Level()
	})
}

func NewHandler(system string) slog.Handler {
	return gcHandler{
		system: system,
		h:      handler,
	}
}

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

func SensitiveString(s string) string {
	if IsDebugLoggingEnabled() {
		return s
	}
	// Redact vales if not set to debug or lower
	return redactedValue
}

func IsDebugLoggingEnabled() bool {
	return lvl <= slog.LevelDebug
}

func DynamicLevel() slog.Level {
	if lvl <= slog.LevelDebug {
		return slog.LevelDebug
	}
	return slog.LevelInfo
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
func (g gcHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return loggingEnabled && g.h.Enabled(ctx, level) && systems[g.system]
}

func (g gcHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(slog.String("system", g.system))
	return g.h.Handle(ctx, r)
}

func (g gcHandler) WithAttrs(a []slog.Attr) slog.Handler { return g.h.WithAttrs(a) }

func (g gcHandler) WithGroup(name string) slog.Handler { return g.h.WithGroup(name) }
