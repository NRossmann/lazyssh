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

// Package agent_config implements the AgentConfigRepository port,
// reading and writing pre-defined SSH agent socket configurations
// from ~/.lazyssh/agents.json.
package agent_config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Adembc/lazyssh/internal/core/domain"
	"go.uber.org/zap"
)

// Repository implements ports.AgentConfigRepository backed by a JSON file.
type Repository struct {
	filePath string
	logger   *zap.SugaredLogger
}

// NewRepository creates a new agent config repository.
// filePath should point to something like ~/.lazyssh/agents.json.
func NewRepository(logger *zap.SugaredLogger, filePath string) *Repository {
	return &Repository{
		filePath: filePath,
		logger:   logger,
	}
}

// ListAgents reads all agent configs from the JSON file.
// Returns an empty slice (not an error) if the file does not exist yet.
func (r *Repository) ListAgents() ([]domain.AgentConfig, error) {
	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
		return []domain.AgentConfig{}, nil
	}

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("read agents config '%s': %w", r.filePath, err)
	}

	if len(data) == 0 {
		return []domain.AgentConfig{}, nil
	}

	var agents []domain.AgentConfig
	if err := json.Unmarshal(data, &agents); err != nil {
		r.logger.Errorw("failed to parse agents config", "path", r.filePath, "error", err)
		return nil, fmt.Errorf("parse agents config JSON '%s': %w", r.filePath, err)
	}

	return agents, nil
}

// SaveAgents writes the given agent configs to the JSON file,
// creating the parent directory if needed.
func (r *Repository) SaveAgents(agents []domain.AgentConfig) error {
	if err := r.ensureDirectory(); err != nil {
		return fmt.Errorf("ensure agents config directory: %w", err)
	}

	data, err := json.MarshalIndent(agents, "", "  ")
	if err != nil {
		r.logger.Errorw("failed to marshal agents config", "error", err)
		return fmt.Errorf("marshal agents config: %w", err)
	}

	if err := os.WriteFile(r.filePath, data, 0o600); err != nil {
		r.logger.Errorw("failed to write agents config", "path", r.filePath, "error", err)
		return fmt.Errorf("write agents config '%s': %w", r.filePath, err)
	}

	return nil
}

func (r *Repository) ensureDirectory() error {
	dir := filepath.Dir(r.filePath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("mkdir '%s': %w", dir, err)
	}
	return nil
}
