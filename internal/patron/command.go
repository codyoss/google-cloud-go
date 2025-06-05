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
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

// CmdWrapper is a wrapper around exec.Cmd with logging.
type CmdWrapper struct {
	*exec.Cmd
}

// Command wraps a exec.Command to add some logging about commands being run.
// The commands stdout/stderr default to os.Stdout/os.Stderr respectfully.
func Command(name string, arg ...string) *CmdWrapper {
	c := &CmdWrapper{exec.Command(name, arg...)}
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return &CmdWrapper{exec.Command(name, arg...)}
}

// Run a command.
func (c *CmdWrapper) Run() error {
	b, err := c.Output()
	if len(b) > 0 {
		slog.Info("Command Output", "output", string(b))
	}
	return err
}

// Output a command.
func (c *CmdWrapper) Output() ([]byte, error) {
	c.Env = append(c.Env,
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
	)
	slog.Info("Running Command", "command", strings.Join(c.Args, " "), "dir", c.Dir, "env", c.Env)
	b, err := c.Cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			slog.Error("Command failed", "error", string(ee.Stderr))
		}
	}
	return b, err
}
