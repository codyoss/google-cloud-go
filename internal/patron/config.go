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
	"encoding/json"
	"fmt"
	"iter"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/bazelbuild/buildtools/build"
	"github.com/ghodss/yaml"
	"google.golang.org/genproto/googleapis/api/serviceconfig"
	"google.golang.org/protobuf/encoding/protojson"
)

type UnifiedConfig struct {
	serviceConfig *serviceconfig.Service
	serviceDir    string
	sourcePath    string
	// From BUILD.bazel
	protoImportPath   string
	gapicImportPath   string
	grpcServiceConfig string
	metadata          bool
	releaseLevel      string
	restNumericEnums  bool
	serviceYaml       string
	transport         string
	diregapic         bool
}

func NewUnifiedConfig(serviceDir, sourcePath string) (*UnifiedConfig, error) {
	uc := &UnifiedConfig{
		serviceDir: serviceDir,
		sourcePath: sourcePath,
	}
	if err := uc.loadBazelRules(serviceDir); err != nil {
		return nil, fmt.Errorf("failed to load bazel rules: %w", err)
	}
	if err := uc.loadServiceConfig(serviceDir); err != nil {
		return nil, fmt.Errorf("failed to load service config: %w", err)
	}

	slog.Info("unified config loaded", "conf", fmt.Sprintf("%+v", uc))
	return uc, nil
}

func (c *UnifiedConfig) loadBazelRules(dir string) error {
	fp := filepath.Join(dir, "BUILD.bazel")
	data, err := os.ReadFile(fp)
	if err != nil {
		return fmt.Errorf("failed to read BUILD.bazel file %s: %w", fp, err)
	}
	f, err := build.Parse(fp, data)
	if err != nil {
		return fmt.Errorf("failed to parse BUILD.bazel file %s: %w", fp, err)
	}

	// Gapic build target
	for _, rule := range f.Rules("go_gapic_library") {
		if v := rule.AttrString("grpc_service_config"); v != "" {
			c.grpcServiceConfig = v
		}
		if v := rule.AttrString("importpath"); v != "" {
			c.gapicImportPath = v
		}
		if v := rule.AttrLiteral("metadata"); v != "" {
			if b, err := strconv.ParseBool(v); err == nil {
				c.metadata = b
			} else {
				slog.Warn("failed to parse metadata", "error", err, "input", v)
			}
		}
		if v := rule.AttrString("release_level"); v != "" {
			c.releaseLevel = v
		}
		if v := rule.AttrLiteral("rest_numeric_enums"); v != "" {
			if b, err := strconv.ParseBool(v); err == nil {
				c.restNumericEnums = b
			} else {
				slog.Warn("failed to parse rest_numeric_enums", "error", err, "input", v)
			}
		}
		if v := rule.AttrString("service_yaml"); v != "" {
			c.serviceYaml = v
		}
		if v := rule.AttrString("transport"); v != "" {
			c.transport = v
		}
		if v := rule.AttrLiteral("diregapic"); v != "" {
			if b, err := strconv.ParseBool(v); err == nil {
				c.diregapic = b
			} else {
				slog.Warn("failed to parse diregapic", "error", err, "input", v)
			}
		}
	}

	// We are currently migrating go_proto_library to go_grpc_library, so check
	// both for now.
	for _, rule := range f.Rules("go_proto_library") {
		if v := rule.AttrString("importpath"); v != "" {
			c.protoImportPath = v
		}
	}
	for _, rule := range f.Rules("go_grpc_library") {
		if v := rule.AttrString("importpath"); v != "" {
			c.protoImportPath = v
		}
	}
	return nil
}

func (c *UnifiedConfig) loadServiceConfig(dir string) error {
	if c.serviceYaml == "" {
		return fmt.Errorf("service config is empty")
	}
	fp := filepath.Join(dir, c.serviceYaml)
	yb, err := os.ReadFile(fp)
	if err != nil {
		return fmt.Errorf("failed to read service config file %s: %w", fp, err)
	}
	jb, err := yaml.YAMLToJSON(yb)
	if err != nil {
		return fmt.Errorf("failed to convert yaml to json for %s: %w", fp, err)
	}
	sc := &serviceconfig.Service{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(jb, sc); err != nil {
		return fmt.Errorf("failed to unmarshal service config file %s: %w", fp, err)
	}
	c.serviceConfig = sc
	return nil
}

// PipelineState corresponds to the PipelineState proto message.
// It represents the overall state of the generation and release pipeline.
type PipelineState struct {
	ImageTag                 string         `json:"imageTag,omitempty"`
	Libraries                []LibraryState `json:"libraries,omitempty"`
	CommonLibrarySourcePaths []string       `json:"commonLibrarySourcePaths,omitempty"`
	IgnoredAPIPaths          []string       `json:"ignoredApiPaths,omitempty"`
}

func NewPipelineState(pipelineStateFile string) (*PipelineState, error) {
	data, err := os.ReadFile(pipelineStateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read pipeline state file %s: %w", pipelineStateFile, err)
	}
	var ps *PipelineState
	if err := json.Unmarshal(data, &ps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pipeline state file %s: %w", pipelineStateFile, err)
	}
	return ps, nil
}

func (p *PipelineState) UnifiedConfigForID(id string) ([]*UnifiedConfig, error) {
	var ucs []*UnifiedConfig
	for _, lib := range p.Libraries {
		if lib.ID == id {
			for i, v := range lib.APIPaths {
				uc, err := NewUnifiedConfig(filepath.Join("/apis", v), lib.SourcePaths[i])
				if err != nil {
					return nil, fmt.Errorf("failed to load unified config for %s: %w", v, err)
				}
				ucs = append(ucs, uc)
			}
		}
	}
	if len(ucs) == 0 {
		return nil, fmt.Errorf("library with id %s not found", id)
	}
	return ucs, nil
}

func (p *PipelineState) LibraryIDs() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, lib := range p.Libraries {
			if !yield(lib.ID) {
				return
			}
		}
	}
}

// LibraryState corresponds to the LibraryState proto message.
// It represents the generation state of a single library.
type LibraryState struct {
	ID                        string   `json:"id,omitempty"`
	CurrentVersion            string   `json:"currentVersion,omitempty"`
	NextVersion               string   `json:"nextVersion,omitempty"`
	GenerationAutomationLevel string   `json:"generationAutomationLevel,omitempty"` // Corresponds to AutomationLevel enum
	ReleaseAutomationLevel    string   `json:"releaseAutomationLevel,omitempty"`    // Corresponds to AutomationLevel enum
	ReleaseTimestamp          string   `json:"releaseTimestamp,omitempty"`          // google.protobuf.Timestamp as string
	LastGeneratedCommit       string   `json:"lastGeneratedCommit,omitempty"`
	LastReleasedCommit        string   `json:"lastReleasedCommit,omitempty"`
	APIPaths                  []string `json:"apiPaths,omitempty"`
	SourcePaths               []string `json:"sourcePaths,omitempty"`
}

// AutomationLevel constants (string representations of the enum)
const (
	AutomationLevelNone         = "AUTOMATION_LEVEL_NONE"
	AutomationLevelBlocked      = "AUTOMATION_LEVEL_BLOCKED"
	AutomationLevelManualReview = "AUTOMATION_LEVEL_MANUAL_REVIEW"
	AutomationLevelAutomatic    = "AUTOMATION_LEVEL_AUTOMATIC"
)
