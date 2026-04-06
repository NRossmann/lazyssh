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

package agent_config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Adembc/lazyssh/internal/core/domain"
	"go.uber.org/zap"
)

func newTestRepo(t *testing.T) (*Repository, string) {
	t.Helper()
	dir := t.TempDir()
	fp := filepath.Join(dir, "agents.json")
	logger := zap.NewNop().Sugar()
	return NewRepository(logger, fp), fp
}

func TestListAgents_FileDoesNotExist(t *testing.T) {
	repo, _ := newTestRepo(t)

	agents, err := repo.ListAgents()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(agents) != 0 {
		t.Fatalf("expected empty slice, got %d items", len(agents))
	}
}

func TestListAgents_EmptyFile(t *testing.T) {
	repo, fp := newTestRepo(t)

	if err := os.WriteFile(fp, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	agents, err := repo.ListAgents()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(agents) != 0 {
		t.Fatalf("expected empty slice, got %d items", len(agents))
	}
}

func TestListAgents_ValidJSON(t *testing.T) {
	repo, fp := newTestRepo(t)

	want := []domain.AgentConfig{
		{Name: "gpg-agent", Path: "/run/user/1000/gnupg/S.gpg-agent.ssh"},
		{Name: "1password", Path: "~/.1password/agent.sock"},
	}
	data, _ := json.Marshal(want)
	if err := os.WriteFile(fp, data, 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := repo.ListAgents()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d agents, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i].Name != want[i].Name || got[i].Path != want[i].Path {
			t.Errorf("agent[%d]: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestListAgents_InvalidJSON(t *testing.T) {
	repo, fp := newTestRepo(t)

	if err := os.WriteFile(fp, []byte("{bad json}"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := repo.ListAgents()
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestSaveAgents_CreatesFileAndDirectory(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "subdir", "agents.json")
	logger := zap.NewNop().Sugar()
	repo := NewRepository(logger, fp)

	agents := []domain.AgentConfig{
		{Name: "test", Path: "/tmp/agent.sock"},
	}
	if err := repo.SaveAgents(agents); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	data, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("expected file to exist, got %v", err)
	}

	var got []domain.AgentConfig
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("expected valid JSON, got %v", err)
	}
	if len(got) != 1 || got[0].Name != "test" || got[0].Path != "/tmp/agent.sock" {
		t.Errorf("unexpected content: %+v", got)
	}
}

func TestSaveAgents_OverwritesExisting(t *testing.T) {
	repo, _ := newTestRepo(t)

	initial := []domain.AgentConfig{{Name: "old", Path: "/old"}}
	if err := repo.SaveAgents(initial); err != nil {
		t.Fatal(err)
	}

	updated := []domain.AgentConfig{{Name: "new", Path: "/new"}}
	if err := repo.SaveAgents(updated); err != nil {
		t.Fatal(err)
	}

	got, err := repo.ListAgents()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "new" {
		t.Errorf("expected overwritten data, got %+v", got)
	}
}

func TestSaveAgents_EmptySlice(t *testing.T) {
	repo, fp := newTestRepo(t)

	if err := repo.SaveAgents([]domain.AgentConfig{}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(fp)
	if err != nil {
		t.Fatal(err)
	}

	var got []domain.AgentConfig
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("expected valid JSON, got %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty array, got %+v", got)
	}
}

func TestRoundTrip(t *testing.T) {
	repo, _ := newTestRepo(t)

	want := []domain.AgentConfig{
		{Name: "gpg", Path: "/run/user/1000/gnupg/S.gpg-agent.ssh"},
		{Name: "1password", Path: "~/.1password/agent.sock"},
		{Name: "yubikey", Path: "/tmp/yubikey-agent.sock"},
	}

	if err := repo.SaveAgents(want); err != nil {
		t.Fatal(err)
	}

	got, err := repo.ListAgents()
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d agents, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("agent[%d]: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestSaveAgents_FilePermissions(t *testing.T) {
	repo, fp := newTestRepo(t)

	if err := repo.SaveAgents([]domain.AgentConfig{{Name: "a", Path: "/a"}}); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(fp)
	if err != nil {
		t.Fatal(err)
	}

	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("expected file permissions 0600, got %o", perm)
	}
}
