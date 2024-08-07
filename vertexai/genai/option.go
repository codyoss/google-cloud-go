// Copyright 2024 Google LLC
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

package genai

import (
	"google.golang.org/api/option"
	"google.golang.org/api/option/internaloption"
)

// WithREST is an option that enables REST transport for the client.
// The default transport (if this option isn't provided) is gRPC.
func WithREST() option.ClientOption {
	return &withREST{}
}

type withREST struct {
	internaloption.EmbeddableAdapter
}

func (w *withREST) applyVertexaiOpt(c *config) {
	c.withREST = true
}

// WithClientInfo is an option that sets request information
// identifying the product that is calling this client.
func WithClientInfo(key, value string) option.ClientOption {
	return &withClientInfo{key: key, value: value}
}

type withClientInfo struct {
	internaloption.EmbeddableAdapter
	key, value string
}

func (w *withClientInfo) applyVertexaiOpt(c *config) {
	c.ciKey = w.key
	c.ciValue = w.value
}

type config struct {
	// withREST tells the client to use REST as the underlying transport.
	withREST bool
	// key-value pair to add to the Google client info header.
	ciKey   string
	ciValue string
}

// newConfig generates a new config with all the given
// vertexaiClientOptions applied.
func newConfig(opts ...option.ClientOption) config {
	var conf config
	for _, opt := range opts {
		if vOpt, ok := opt.(vertexaiClientOption); ok {
			vOpt.applyVertexaiOpt(&conf)
		}
	}
	return conf
}

// A vertexaiClientOption is an option for a vertexai client.
type vertexaiClientOption interface {
	option.ClientOption
	applyVertexaiOpt(*config)
}
