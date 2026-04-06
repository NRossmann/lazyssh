// Copyright 2025.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package yadm implements the VCSRepository port using the yadm dotfile
// manager. It auto-detects whether yadm is installed and whether it tracks
// any lazyssh-managed files, then stages and commits changes after every
// mutation.
package yadm

import (
	"fmt"
	"os/exec"

	"go.uber.org/zap"
)

// Repository implements ports.VCSRepository backed by the yadm CLI.
type Repository struct {
	logger       *zap.SugaredLogger
	filePaths    []string // absolute paths of files to stage (e.g. ~/.ssh/config)
	trackedPaths []string // subset of filePaths actually tracked by yadm
	available    bool
}

// NewRepository creates a yadm VCS repository.
// filePaths should contain the absolute paths of every lazyssh-managed file
// that might be committed (e.g. ~/.ssh/config, ~/.lazyssh/metadata.json,
// ~/.lazyssh/agents.json). The constructor probes yadm once to determine
// availability.
func NewRepository(logger *zap.SugaredLogger, filePaths []string) *Repository {
	r := &Repository{
		logger:    logger,
		filePaths: filePaths,
	}
	r.detect()
	return r
}

// detect checks whether yadm is installed and which of the configured file
// paths are tracked by it.
func (r *Repository) detect() {
	if _, err := exec.LookPath("yadm"); err != nil {
		r.logger.Debugw("yadm not found in PATH, VCS integration disabled")
		return
	}

	for _, fp := range r.filePaths {
		if isTracked(fp) {
			r.trackedPaths = append(r.trackedPaths, fp)
		}
	}

	if len(r.trackedPaths) == 0 {
		r.logger.Debugw("yadm installed but no lazyssh files are tracked")
		return
	}

	r.available = true
	r.logger.Infow("yadm VCS integration enabled", "tracked", r.trackedPaths)
}

// isTracked returns true if the given absolute path is tracked by yadm.
func isTracked(absPath string) bool {
	// #nosec G204 -- absPath comes from hardcoded config paths, not user input
	cmd := exec.Command("yadm", "ls-files", "--error-unmatch", absPath)
	return cmd.Run() == nil
}

// IsAvailable reports whether yadm is installed and tracks at least one
// lazyssh-managed file.
func (r *Repository) IsAvailable() bool {
	return r.available
}

// Commit stages all tracked lazyssh-managed files and creates a commit.
// It silently returns nil when there are no staged changes (nothing to commit).
func (r *Repository) Commit(message string) error {
	if !r.available {
		return nil
	}

	// Stage each tracked file. yadm add is a no-op for unmodified files.
	for _, fp := range r.trackedPaths {
		// #nosec G204 -- fp comes from hardcoded config paths
		if out, err := exec.Command("yadm", "add", fp).CombinedOutput(); err != nil {
			r.logger.Warnw("yadm add failed", "path", fp, "error", err, "output", string(out))
			return fmt.Errorf("yadm add %s: %w", fp, err)
		}
	}

	// Check if there are any staged changes. yadm diff --cached --quiet
	// exits 0 if clean, 1 if there are staged diffs.
	if err := exec.Command("yadm", "diff", "--cached", "--quiet").Run(); err == nil {
		// Exit code 0 means no staged changes — nothing to commit.
		r.logger.Debugw("yadm: nothing to commit", "message", message)
		return nil
	}

	// Commit the staged changes.
	// #nosec G204 -- message is generated internally, not user-supplied
	if out, err := exec.Command("yadm", "commit", "-m", message).CombinedOutput(); err != nil {
		r.logger.Warnw("yadm commit failed", "error", err, "output", string(out))
		return fmt.Errorf("yadm commit: %w", err)
	}

	r.logger.Infow("yadm commit created", "message", message)
	return nil
}
