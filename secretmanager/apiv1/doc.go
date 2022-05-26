
package secretmanager // import "cloud.google.com/go/secretmanager/apiv1"

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"google.golang.org/api/option"
	"google.golang.org/grpc/metadata"
)

// For more information on implementing a client constructor hook, see
// https://github.com/googleapis/google-cloud-go/wiki/Customizing-constructors.
type clientHookParams struct{}
type clientHook func(context.Context, clientHookParams) ([]option.ClientOption, error)

var versionClient string

func getVersionClient() string {
	if versionClient == "" {
		return "UNKNOWN"
	}
	return versionClient
}

func insertMetadata(ctx context.Context, mds ...metadata.MD) context.Context {
	out, _ := metadata.FromOutgoingContext(ctx)
	out = out.Copy()
	for _, md := range mds {
		for k, v := range md {
			out[k] = append(out[k], v...)
		}
	}
	return metadata.NewOutgoingContext(ctx, out)
}

func checkDisableDeadlines() (bool, error) {
	raw, ok := os.LookupEnv("GOOGLE_API_GO_EXPERIMENTAL_DISABLE_DEFAULT_DEADLINE")
	if !ok {
		return false, nil
	}

	b, err := strconv.ParseBool(raw)
	return b, err
}

// DefaultAuthScopes reports the default set of authentication scopes to use with this package.
func DefaultAuthScopes() []string {
	return []string{
		"https://www.googleapis.com/auth/cloud-platform",
	}
}

// versionGo returns the Go runtime version. The returned string
// has no whitespace, suitable for reporting in header.
func versionGo() string {
	const develPrefix = "devel +"

	s := runtime.Version()
	if strings.HasPrefix(s, develPrefix) {
		s = s[len(develPrefix):]
		if p := strings.IndexFunc(s, unicode.IsSpace); p >= 0 {
			s = s[:p]
		}
		return s
	}

	notSemverRune := func(r rune) bool {
		return !strings.ContainsRune("0123456789.", r)
	}

	if strings.HasPrefix(s, "go1") {
		s = s[2:]
		var prerelease string
		if p := strings.IndexFunc(s, notSemverRune); p >= 0 {
			s, prerelease = s[:p], s[p:]
		}
		if strings.HasSuffix(s, ".") {
			s += "0"
		} else if strings.Count(s, ".") < 2 {
			s += ".0"
		}
		if prerelease != "" {
			s += "-" + prerelease
		}
		return s
	}
	return "UNKNOWN"
}
