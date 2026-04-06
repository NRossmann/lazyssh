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
	"strings"
	"time"

	"github.com/Adembc/lazyssh/internal/core/domain"
	"github.com/Adembc/lazyssh/internal/core/ports"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

// AgentManager is a full-screen component for managing pre-defined SSH agent
// configurations stored in ~/.lazyssh/agents.json. It is opened from the main
// screen via the 'A' key and provides its own add / edit / delete key bindings.
type AgentManager struct {
	*tview.Flex

	app     *tview.Application
	logger  *zap.SugaredLogger
	repo    ports.AgentConfigRepository
	vcsRepo ports.VCSRepository

	list      *AgentList
	statusBar *tview.TextView
	emptyText *tview.TextView

	agents  []domain.AgentConfig
	onClose func()
}

// NewAgentManager creates a new agent management screen.
func NewAgentManager(
	app *tview.Application,
	logger *zap.SugaredLogger,
	repo ports.AgentConfigRepository,
	vcsRepo ports.VCSRepository,
) *AgentManager {
	am := &AgentManager{
		Flex:    tview.NewFlex().SetDirection(tview.FlexRow),
		app:     app,
		logger:  logger,
		repo:    repo,
		vcsRepo: vcsRepo,
		list:    NewAgentList(),
	}
	am.build()
	return am
}

// OnClose sets the callback invoked when the user leaves the agent manager.
func (am *AgentManager) OnClose(fn func()) *AgentManager {
	am.onClose = fn
	return am
}

func (am *AgentManager) build() {
	// Title bar
	title := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	title.SetBackgroundColor(tcell.Color234)
	title.SetText("[#FFFFFF::b]SSH Agent Manager[-]")

	separator := tview.NewTextView().SetDynamicColors(true)
	separator.SetBackgroundColor(tcell.Color235)
	separator.SetText("[#444444]" + strings.Repeat("─", 200) + "[-]")

	// Empty-state text (shown when there are no agents)
	am.emptyText = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetWordWrap(true)
	am.emptyText.SetText("\n\n[#888888]No SSH agents configured.\n\nPress [white]a[-][#888888] to add your first agent.[-]")

	// Status bar
	am.statusBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	am.statusBar.SetBackgroundColor(tcell.Color235)
	am.statusBar.SetText(agentManagerStatusText())

	// Layout
	am.Flex.
		AddItem(title, 1, 0, false).
		AddItem(separator, 1, 0, false).
		AddItem(am.list, 0, 1, true).
		AddItem(am.statusBar, 1, 0, false)

	// Key handling
	am.Flex.SetInputCapture(am.handleKeys)
}

func agentManagerStatusText() string {
	return "[white]a[-] Add  " +
		"\u2022 [white]e[-] Edit  " +
		"\u2022 [white]d[-] Delete  " +
		"\u2022 [white]Esc[-] Back"
}

func (am *AgentManager) handleKeys(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEscape {
		am.close()
		return nil
	}

	switch event.Rune() {
	case 'q':
		am.close()
		return nil
	case 'a':
		am.showAddForm()
		return nil
	case 'e':
		am.showEditForm()
		return nil
	case 'd':
		am.showDeleteConfirm()
		return nil
	}

	return event
}

// Load reads agents from the repository and refreshes the list.
func (am *AgentManager) Load() {
	agents, err := am.repo.ListAgents()
	if err != nil {
		am.logger.Errorw("failed to load agents", "error", err)
		agents = []domain.AgentConfig{}
	}
	am.agents = agents
	am.refreshList()
}

func (am *AgentManager) refreshList() {
	am.list.UpdateAgents(am.agents)

	// Toggle between list and empty state
	am.Flex.RemoveItem(am.list)
	am.Flex.RemoveItem(am.emptyText)

	if len(am.agents) == 0 {
		// Insert empty text before the status bar (index 2 after title+separator)
		am.Flex.AddItem(am.emptyText, 0, 1, true)
	} else {
		am.Flex.AddItem(am.list, 0, 1, true)
	}

	// Re-add status bar at the bottom (remove first to avoid duplicates)
	am.Flex.RemoveItem(am.statusBar)
	am.Flex.AddItem(am.statusBar, 1, 0, false)
}

func (am *AgentManager) save() error {
	if err := am.repo.SaveAgents(am.agents); err != nil {
		am.logger.Errorw("failed to save agents", "error", err)
		return err
	}
	return nil
}

// vcsCommit attempts to commit lazyssh-managed files via the VCS repository.
// Errors are logged but never propagated.
func (am *AgentManager) vcsCommit(message string) {
	if am.vcsRepo == nil {
		return
	}
	if err := am.vcsRepo.Commit(message); err != nil {
		am.logger.Warnw("vcs commit failed", "message", message, "error", err)
	}
}

func (am *AgentManager) close() {
	if am.onClose != nil {
		am.onClose()
	}
}

func (am *AgentManager) showStatusTemp(msg string) {
	am.statusBar.SetText("[#A0FFA0]" + msg + "[-]")
	time.AfterFunc(2*time.Second, func() {
		am.app.QueueUpdateDraw(func() {
			am.statusBar.SetText(agentManagerStatusText())
		})
	})
}

// ─── Add ────────────────────────────────────────────────────────────────────

func (am *AgentManager) showAddForm() {
	form := tview.NewForm()
	form.SetBorder(true).
		SetTitle(" Add SSH Agent ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.Color238).
		SetTitleColor(tcell.Color250)

	form.AddInputField("Name:", "", 30, nil, nil)
	form.AddInputField("Path:", "", 50, nil, nil)

	form.AddButton("Save", func() {
		name := strings.TrimSpace(form.GetFormItem(0).(*tview.InputField).GetText())
		path := strings.TrimSpace(form.GetFormItem(1).(*tview.InputField).GetText())

		if name == "" || path == "" {
			am.showError("Name and Path are required.")
			return
		}

		am.agents = append(am.agents, domain.AgentConfig{Name: name, Path: path})
		if err := am.save(); err != nil {
			am.showError(fmt.Sprintf("Save failed: %v", err))
			return
		}
		am.vcsCommit(fmt.Sprintf("lazyssh: add agent %s", name))

		am.refreshList()
		am.app.SetRoot(am, true)
		am.app.SetFocus(am.list)
		am.showStatusTemp("Agent added: " + name)
	})
	form.AddButton("Cancel", func() {
		am.app.SetRoot(am, true)
		am.app.SetFocus(am.list)
	})
	form.SetCancelFunc(func() {
		am.app.SetRoot(am, true)
		am.app.SetFocus(am.list)
	})

	am.app.SetRoot(form, true)
	am.app.SetFocus(form)
}

// ─── Edit ───────────────────────────────────────────────────────────────────

func (am *AgentManager) showEditForm() {
	agent, idx, ok := am.list.GetSelectedAgent()
	if !ok {
		return
	}

	form := tview.NewForm()
	form.SetBorder(true).
		SetTitle(fmt.Sprintf(" Edit SSH Agent: %s ", agent.Name)).
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.Color238).
		SetTitleColor(tcell.Color250)

	form.AddInputField("Name:", agent.Name, 30, nil, nil)
	form.AddInputField("Path:", agent.Path, 50, nil, nil)

	form.AddButton("Save", func() {
		name := strings.TrimSpace(form.GetFormItem(0).(*tview.InputField).GetText())
		path := strings.TrimSpace(form.GetFormItem(1).(*tview.InputField).GetText())

		if name == "" || path == "" {
			am.showError("Name and Path are required.")
			return
		}

		am.agents[idx] = domain.AgentConfig{Name: name, Path: path}
		if err := am.save(); err != nil {
			am.showError(fmt.Sprintf("Save failed: %v", err))
			return
		}
		am.vcsCommit(fmt.Sprintf("lazyssh: update agent %s", name))

		am.refreshList()
		am.app.SetRoot(am, true)
		am.app.SetFocus(am.list)
		am.showStatusTemp("Agent updated: " + name)
	})
	form.AddButton("Cancel", func() {
		am.app.SetRoot(am, true)
		am.app.SetFocus(am.list)
	})
	form.SetCancelFunc(func() {
		am.app.SetRoot(am, true)
		am.app.SetFocus(am.list)
	})

	am.app.SetRoot(form, true)
	am.app.SetFocus(form)
}

// ─── Delete ─────────────────────────────────────────────────────────────────

func (am *AgentManager) showDeleteConfirm() {
	agent, idx, ok := am.list.GetSelectedAgent()
	if !ok {
		return
	}

	msg := fmt.Sprintf("Delete agent %q?\n(%s)\n\nThis action cannot be undone.", agent.Name, agent.Path)

	modal := tview.NewModal().
		SetText(msg).
		AddButtons([]string{"[yellow]C[-]ancel", "[yellow]D[-]elete"}).
		SetDoneFunc(func(buttonIndex int, _ string) {
			if buttonIndex == 1 {
				am.agents = append(am.agents[:idx], am.agents[idx+1:]...)
				if err := am.save(); err != nil {
					am.showError(fmt.Sprintf("Delete failed: %v", err))
					return
				}
				am.vcsCommit(fmt.Sprintf("lazyssh: delete agent %s", agent.Name))
				am.refreshList()
				am.showStatusTemp("Agent deleted: " + agent.Name)
			}
			am.app.SetRoot(am, true)
			am.app.SetFocus(am.list)
		})

	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'c', 'C':
			am.app.SetRoot(am, true)
			am.app.SetFocus(am.list)
			return nil
		case 'd', 'D':
			am.agents = append(am.agents[:idx], am.agents[idx+1:]...)
			if err := am.save(); err != nil {
				am.showError(fmt.Sprintf("Delete failed: %v", err))
				return nil
			}
			am.vcsCommit(fmt.Sprintf("lazyssh: delete agent %s", agent.Name))
			am.refreshList()
			am.app.SetRoot(am, true)
			am.app.SetFocus(am.list)
			am.showStatusTemp("Agent deleted: " + agent.Name)
			return nil
		}
		return event
	})

	am.app.SetRoot(modal, true)
}

// ─── Error Modal ────────────────────────────────────────────────────────────

func (am *AgentManager) showError(msg string) {
	modal := tview.NewModal().
		SetText(msg).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(_ int, _ string) {
			am.app.SetRoot(am, true)
			am.app.SetFocus(am.list)
		})
	am.app.SetRoot(modal, true)
}
