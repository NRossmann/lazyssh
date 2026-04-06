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

package ui

import (
	"fmt"

	"github.com/Adembc/lazyssh/internal/core/domain"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// AgentList is a tview.List wrapper that displays pre-defined SSH agent
// configurations. It mirrors the ServerList component pattern.
type AgentList struct {
	*tview.List
	agents            []domain.AgentConfig
	onSelectionChange func(domain.AgentConfig)
}

// NewAgentList creates and initializes a new AgentList widget.
func NewAgentList() *AgentList {
	al := &AgentList{
		List: tview.NewList(),
	}
	al.build()
	return al
}

func (al *AgentList) build() {
	al.List.ShowSecondaryText(true)
	al.List.SetBorder(true).
		SetTitle(" SSH Agents ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.Color238).
		SetTitleColor(tcell.Color250)
	al.List.
		SetSelectedBackgroundColor(tcell.Color24).
		SetSelectedTextColor(tcell.Color255).
		SetHighlightFullLine(true).
		SetSecondaryTextColor(tcell.Color245)

	al.List.SetChangedFunc(func(index int, _, _ string, _ rune) {
		if index >= 0 && index < len(al.agents) && al.onSelectionChange != nil {
			al.onSelectionChange(al.agents[index])
		}
	})
}

// UpdateAgents replaces the displayed list with the given agents.
func (al *AgentList) UpdateAgents(agents []domain.AgentConfig) {
	al.agents = agents
	al.List.Clear()

	for i, agent := range agents {
		primary := fmt.Sprintf("  🔑  [white::b]%s[-]", agent.Name)
		secondary := fmt.Sprintf("       [#888888]%s[-]", agent.Path)
		idx := i
		al.List.AddItem(primary, secondary, 0, func() {
			// select on enter — no-op for now, edit is via 'e'
			_ = idx
		})
	}

	if al.List.GetItemCount() > 0 {
		al.List.SetCurrentItem(0)
		if al.onSelectionChange != nil {
			al.onSelectionChange(al.agents[0])
		}
	}
}

// GetSelectedAgent returns the currently highlighted agent and its index.
func (al *AgentList) GetSelectedAgent() (domain.AgentConfig, int, bool) {
	idx := al.List.GetCurrentItem()
	if idx >= 0 && idx < len(al.agents) {
		return al.agents[idx], idx, true
	}
	return domain.AgentConfig{}, -1, false
}

// OnSelectionChange sets the callback invoked when the highlighted item changes.
func (al *AgentList) OnSelectionChange(fn func(domain.AgentConfig)) *AgentList {
	al.onSelectionChange = fn
	return al
}
