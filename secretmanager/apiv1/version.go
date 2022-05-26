package secretmanager

import "cloud.google.com/go/secretmanager/internal"

func init() {
	versionClient = internal.Version
}
