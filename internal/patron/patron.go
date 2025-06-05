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
	"fmt"
	"log/slog"
	"os"
)

func main() {
	slog.Info("patron invoked")
	if err := run(context.Background()); err != nil {
		slog.Error("patron failed to run", "error", err)
		os.Exit(1)
	}
	slog.Info("patron finished")
}

func run(ctx context.Context) error {
	if len(os.Args) < 2 {
		return fmt.Errorf("expected at least one argument, got %d", len(os.Args)-1)
	}
	cmd := os.Args[1]
	switch cmd {
	case "generate-raw":
		return generateRawCmd(ctx, os.Args[2:])
	case "generate-library":
		return generateLibraryCmd(ctx, os.Args[2:])
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}
