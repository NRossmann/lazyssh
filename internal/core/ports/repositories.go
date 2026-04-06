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

package ports

import "github.com/Adembc/lazyssh/internal/core/domain"

type ServerRepository interface {
	ListServers(query string) ([]domain.Server, error)
	UpdateServer(server domain.Server, newServer domain.Server) error
	AddServer(server domain.Server) error
	DeleteServer(server domain.Server) error
	SetPinned(alias string, pinned bool) error
	RecordSSH(alias string) error
}

// AgentConfigRepository provides access to pre-defined SSH agent configurations.
type AgentConfigRepository interface {
	// ListAgents returns all configured SSH agent entries.
	ListAgents() ([]domain.AgentConfig, error)
	// SaveAgents persists the given SSH agent entries, replacing all existing ones.
	SaveAgents(agents []domain.AgentConfig) error
}

// VCSRepository provides version control integration for dotfile managers
// such as yadm. Implementations must be safe to call even when the underlying
// tool is not installed — in that case every method is a silent no-op.
type VCSRepository interface {
	// IsAvailable reports whether the VCS tool is installed and manages
	// at least one lazyssh-related file.
	IsAvailable() bool
	// Commit stages all tracked lazyssh-managed files and creates a commit
	// with the given message. It silently returns nil when nothing changed.
	Commit(message string) error
}
