// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func generateRawCmd(ctx context.Context, args []string) error {
	// TODO(codyoss): maybe impl this. It does not seem required though...
	slog.Info("generate-raw command invoked", "args", args)

	fs := flag.NewFlagSet("generate-raw", flag.ExitOnError)
	apiRoot := fs.String("api-root", "", "The root directory of the API protos")
	outputDir := fs.String("output", "", "The output directory for generated files")
	apiPath := fs.String("api-path", "", "The path to the API within the api-root")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	slog.Debug("parsed flags",
		"api-root", *apiRoot,
		"output", *outputDir,
		"api-path", *apiPath)

	if *apiRoot == "" {
		return fmt.Errorf("-api-root is required")
	}
	if *outputDir == "" {
		return fmt.Errorf("-output is required")
	}
	if *apiPath == "" {
		return fmt.Errorf("-api-path is required")
	}

	fullAPIPath := filepath.Join(*apiRoot, *apiPath)
	files, err := os.ReadDir(fullAPIPath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", fullAPIPath, err)
	}

	slog.Info("Files in directory:", "path", fullAPIPath)
	for _, file := range files {
		slog.Info("file", "name", file.Name(), "isDir", file.IsDir())
	}

	return nil
}

func generateLibraryCmd(ctx context.Context, args []string) error {
	slog.Info("generate-library command invoked", "args", args)

	fs := flag.NewFlagSet("generate-library", flag.ExitOnError)
	apiRoot := fs.String("api-root", "", "effectively the googleapis directory. Required.")
	generatorInput := fs.String("generator-input", "", "required.")
	outputDir := fs.String("output", "", "root folder for result; required, must exist")
	libraryID := fs.String("library-id", "", "ID of the library to be generated.")
	dryRun := fs.Bool("dry-run", false, "used for testing purposes but doesn’t write anything to the output.")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	slog.Debug("parsed flags",
		"api-root", *apiRoot,
		"generator-input", *generatorInput,
		"output", *outputDir,
		"library-id", *libraryID,
		"dry-run", *dryRun,
	)

	if *apiRoot == "" {
		return fmt.Errorf("-api-root is required")
	}
	if *generatorInput == "" {
		return fmt.Errorf("-generator-input is required")
	}
	if *outputDir == "" {
		return fmt.Errorf("-output is required")
	}

	slog.Info("loading pipeline state", "file", filepath.Join(*generatorInput, "pipeline-state.json"))
	ps, err := NewPipelineState(filepath.Join(*generatorInput, "pipeline-state.json"))
	if err != nil {
		return fmt.Errorf("failed to load pipline state: %w", err)
	}
	// TODO: handle the general case of no id
	ucs, err := ps.UnifiedConfigForID(*libraryID)
	if err != nil {
		return fmt.Errorf("failed to load unified config: %w", err)
	}
	for _, uc := range ucs {
		if err := protoc(uc); err != nil {
			return fmt.Errorf("protoc failed: %w", err)
		}
		// TODO: move the generated snippets
	}

	return nil
}

func protoc(uc *UnifiedConfig) error {
	slog.Info("running protoc", "sourcePath", uc.sourcePath)

	// Gather the proto files
	var protoFiles []string
	dirEntries, err := os.ReadDir(uc.serviceDir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", uc.serviceDir, err)
	}
	for _, entry := range dirEntries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".proto" {
			fullPath := filepath.Join(uc.serviceDir, entry.Name())
			protoFiles = append(protoFiles, fullPath)
			slog.Debug("found proto file", "path", fullPath)
		}
	}

	stubs := "--go_out=/output/"
	args := []string{
		"--experimental_allow_proto3_optional",
		stubs,
		"-I", "/apis",
		"-I", "/include",
	}
	if uc.gapicImportPath != "" {
		args = append(args,
			"--go_gapic_out", "/output",
			"--go_gapic_opt", fmt.Sprintf("go-gapic-package=%s", uc.gapicImportPath),
		)
		if uc.serviceYaml != "" {
			args = append(args, "--go_gapic_opt", fmt.Sprintf("api-service-config=%s", filepath.Join(uc.serviceDir, uc.serviceYaml)))
		}
		if uc.grpcServiceConfig != "" {
			args = append(args, "--go_gapic_opt", fmt.Sprintf("grpc-service-config=%s", filepath.Join(uc.serviceDir, uc.grpcServiceConfig)))
		}
		if uc.transport != "" {
			args = append(args, "--go_gapic_opt", fmt.Sprintf("transport=%s", uc.transport))
		}
		if uc.releaseLevel != "" {
			args = append(args, "--go_gapic_opt", fmt.Sprintf("release-level=%s", uc.releaseLevel))
		}
		if uc.metadata {
			args = append(args, "--go_gapic_opt", "metadata")
		}
		if uc.diregapic {
			args = append(args, "--go_gapic_opt", "diregapic")
		}
		if uc.restNumericEnums {
			args = append(args, "--go_gapic_opt", "rest-numeric-enums")
		}
	}

	args = append(args, protoFiles...)
	c := Command("protoc", args...)
	c.Dir = "/output"
	return c.Run()
}
