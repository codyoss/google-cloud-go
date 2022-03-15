package gensnippets

import (
	"os"
	"testing"
)

func TestGenerateMetadata(t *testing.T) {
	confs := []*MetadataConfig{
		&MetadataConfig{
			GoogleapisDir:   "/Users/codyoss/oss/googleapis/google/cloud/secretmanager/v1",
			GenprotoDir:     "/Users/codyoss/oss/go-genproto/googleapis/cloud/secretmanager/v1",
			CloudDir:        "/Users/codyoss/oss/google-cloud-go/secretmanager/apiv1",
			GapicImportPath: "cloud.google.com/go/secretmanager/apiv1",
		},
	}

	GenerateMetadata(confs, os.Stdout.Name())
}

func TestGen2(t *testing.T) {
	confs := []*MetadataConfig{
		&MetadataConfig{
			GoogleapisDir:   "google/cloud/secretmanager/v1",
			GenprotoDir:     "/Users/codyoss/oss/go-genproto/googleapis/cloud/secretmanager/v1",
			CloudDir:        "/Users/codyoss/oss/google-cloud-go/secretmanager/apiv1",
			GapicImportPath: "cloud.google.com/go/secretmanager/apiv1",
		},
	}

	if err := Gen2(confs[0].GoogleapisDir); err != nil {
		t.Fatal(err)
	}
}
