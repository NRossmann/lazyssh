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

package yadm

import (
	"os/exec"
	"testing"

	"github.com/Adembc/lazyssh/internal/core/ports"
	"go.uber.org/zap"
)

// Compile-time interface checks.
var (
	_ ports.VCSRepository = (*Repository)(nil)
	_ ports.VCSRepository = (*NoopRepository)(nil)
)

func TestNoopRepository_IsAvailable(t *testing.T) {
	r := NewNoopRepository()
	if r.IsAvailable() {
		t.Error("NoopRepository.IsAvailable() should return false")
	}
}

func TestNoopRepository_Commit(t *testing.T) {
	r := NewNoopRepository()
	if err := r.Commit("test message"); err != nil {
		t.Errorf("NoopRepository.Commit() should return nil, got %v", err)
	}
}

func TestNewRepository_YadmNotInstalled(t *testing.T) {
	// This test verifies behavior when yadm is not in PATH.
	// If yadm happens to be installed in the test environment, skip it.
	if _, err := exec.LookPath("yadm"); err == nil {
		t.Skip("yadm is installed; skipping not-installed test")
	}

	logger := zap.NewNop().Sugar()
	r := NewRepository(logger, []string{"/tmp/nonexistent-file"})

	if r.IsAvailable() {
		t.Error("expected IsAvailable() == false when yadm is not installed")
	}
	if err := r.Commit("should be no-op"); err != nil {
		t.Errorf("expected Commit to return nil when unavailable, got %v", err)
	}
}

func TestNewRepository_NoTrackedFiles(t *testing.T) {
	// If yadm is installed but the file is not tracked, IsAvailable should be false.
	if _, err := exec.LookPath("yadm"); err != nil {
		t.Skip("yadm is not installed; skipping tracked-files test")
	}

	logger := zap.NewNop().Sugar()
	r := NewRepository(logger, []string{"/tmp/definitely-not-tracked-by-yadm-" + t.Name()})

	if r.IsAvailable() {
		t.Error("expected IsAvailable() == false for untracked file")
	}
}
