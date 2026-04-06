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

// NoopRepository is a VCSRepository that does nothing. It is used when yadm
// is not installed or does not track any lazyssh-managed files.
type NoopRepository struct{}

// NewNoopRepository returns a no-op VCS repository.
func NewNoopRepository() *NoopRepository { return &NoopRepository{} }

// IsAvailable always returns false.
func (n *NoopRepository) IsAvailable() bool { return false }

// Commit is a no-op and always returns nil.
func (n *NoopRepository) Commit(_ string) error { return nil }
