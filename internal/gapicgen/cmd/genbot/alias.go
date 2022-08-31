package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"cloud.google.com/go/internal/aliasfix"
	"cloud.google.com/go/internal/aliasgen"
	"cloud.google.com/go/internal/gapicgen/execv"
	"cloud.google.com/go/internal/gapicgen/execv/gocmd"
	"golang.org/x/sync/errgroup"
)

type aliasConfig struct {
	googleapisDir   string
	gocloudDir      string
	genprotoDir     string
	gapicToGenerate string
}

func genAlias(ctx context.Context, c aliasConfig) error {
	log.Println("creating temp dir")
	tmpDir, err := ioutil.TempDir("", "genalias")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("temp dir created at %s\n", tmpDir)
	tmpGenprotoDir := filepath.Join(tmpDir, "genproto")
	tmpGocloudDir := filepath.Join(tmpDir, "gocloud")
	grp, _ := errgroup.WithContext(ctx)
	gitShallowClone(grp, "https://github.com/googleapis/go-genproto", c.genprotoDir, tmpGenprotoDir)
	gitShallowClone(grp, "https://github.com/googleapis/google-cloud-go", c.gocloudDir, tmpGocloudDir)
	return nil
}

func generateAliases(goCloudDir, genprotoDir string) error {
	// Make changes in genproto
	for genprotoImport, pkg := range aliasfix.GenprotoPkgMigration {
		if pkg.Status != aliasfix.StatusMigrated {
			continue
		}
		log.Printf("generating alias for: %q", pkg.ImportPath)
		cloudSrcDir := filepath.Join(goCloudDir, strings.TrimPrefix(pkg.ImportPath, "cloud.google.com/go/"))
		genprotoSrcDir := filepath.Join(genprotoDir, strings.TrimPrefix(genprotoImport, "google.golang.org/genproto/"))
		// Find the latest version of the cloud module.
		cloudMod, cloudModVersion, err := gocmd.ListModVersion(cloudSrcDir)
		if err != nil {
			return err
		}
		// Checkout said version.
		if err := checkoutRef(cloudSrcDir, cloudMod, cloudModVersion); err != nil {
			return err
		}
		// Have genproto depend on this version.
		if err := gocmd.Get(genprotoSrcDir, cloudMod, cloudModVersion); err != nil {
			return err
		}
		// Generate new aliases
		if err := aliasgen.Run(cloudSrcDir, genprotoSrcDir); err != nil {
			log.Printf("error while generating aliases for: %q", pkg.ImportPath)
			return err
		}
		// Restore checkout to main for cloud.
		if err := checkoutRef(cloudSrcDir, "", ""); err != nil {
			return err
		}
	}
	// Open genproto PR
	return nil
}

// checkoutRef checks out the ref that is constructed by using the module
// name and version. If the version provided is empty the main branch will be
// checked out.
func checkoutRef(dir string, modName string, version string) error {
	var ref string
	if version == "" {
		ref = "main"
	} else {
		// Transform cloud.google.com/storage/v2 into storage for example
		vPrefix := strings.Split(strings.TrimPrefix(modName, "cloud.google.com/go/"), "/")[0]
		ref = fmt.Sprintf("%s/%s", vPrefix, version)

	}
	cmd := execv.Command("git", "checkout", ref)
	cmd.Dir = dir
	return cmd.Run()
}
