package gensnippets

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/internal/godocfx/pkgload"
	"github.com/jhump/protoreflect/desc/protoparse"
)

type MetadataConfig struct {
	GoogleapisDir   string
	GenprotoDir     string
	CloudDir        string
	GapicImportPath string
}

func GenerateMetadata(configs []*MetadataConfig, outDir string) error {
	for _, c := range configs {
		pis, err := pkgload.Load("./...", c.CloudDir, nil)
		if err != nil {
			return fmt.Errorf("failed to load packages: %v", err)
		}
		types := pis[0].Doc.Types
		meths := types[1].Methods
		meth := meths[0]
		log.Println(meth.Doc)
	}
	return nil
}

func Gen2(dir string) error {
	var protoFiles []string
	if err := fs.WalkDir(os.DirFS("/Users/codyoss/oss/googleapis/google/cloud/secretmanager/v1"), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(d.Name(), ".proto") {
			protoFiles = append(protoFiles, filepath.Join(dir, path))
		}
		return nil
	}); err != nil {
		return err
	}
	p := protoparse.Parser{
		ImportPaths: []string{"/Users/codyoss/oss/protobuf", "/Users/codyoss/oss/googleapis"},
		IncludeSourceCodeInfo: true,
	}
	fd, err := p.ParseFiles(protoFiles...)
	if err != nil {
		return err
	}
	meth := fd[1].GetServices()[0].GetMethods()[0]
	si := meth.GetSourceInfo()
	log.Println(si.GetLeadingComments())
	return nil
}
