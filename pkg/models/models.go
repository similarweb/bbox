package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// ConfirmActionModel handles confirmation of actions.
type ConfirmActionModel struct {
	Confirmed bool
	Quitting  bool
}

// NewConfirmActionModel creates a new instance of ConfirmActionModel.
func NewConfirmActionModel() ConfirmActionModel {
	return ConfirmActionModel{
		Confirmed: false,
		Quitting:  false,
	}
}

func (m ConfirmActionModel) Init() tea.Cmd {
	return nil
}

func (m ConfirmActionModel) Update(msg tea.Msg) (ConfirmActionModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			m.Confirmed = true
			m.Quitting = true
		case "n", "q", "ctrl+c":
			m.Quitting = true
		}
	}
	return m, nil
}

func (m ConfirmActionModel) View() string {
	return "Press 'y' to confirm, 'n' to cancel, or 'q' to quit."
}

func (m ConfirmActionModel) IsConfirmed() bool {
	return m.Confirmed
}

// ListModel handles listing and navigation
type ListModel struct {
	UnusedVcsRoots []string
	Cursor         int
}

func NewListModel(vcsRoots []string) ListModel {
	return ListModel{
		UnusedVcsRoots: vcsRoots,
		Cursor:         0,
	}
}

func (m ListModel) Init() tea.Cmd {
	return nil
}

func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.UnusedVcsRoots)-1 {
				m.Cursor++
			}
		}
	}
	return m, nil
}

func (m ListModel) View() string {
	s := "The list contains the following objects:\n" //-----------------------------------------------------
	for i, vcsRoot := range m.UnusedVcsRoots {
		cursor := " "
		if m.Cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, vcsRoot)
	}
	return s
}

// UnusedVcsRootsModel manages both ConfirmActionModel and ListModel.
type UnusedVcsRootsModel struct {
	ActionModel ConfirmActionModel
	ListModel   ListModel
}

func NewUnusedVcsRootsModel(vcsRoots []string) UnusedVcsRootsModel {
	return UnusedVcsRootsModel{
		ActionModel: NewConfirmActionModel(),
		ListModel:   NewListModel(vcsRoots),
	}
}

func (m UnusedVcsRootsModel) Init() tea.Cmd {
	return nil
}

func (m UnusedVcsRootsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.ActionModel.Quitting {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "y" || msg.String() == "n" || msg.String() == "q" || msg.String() == "ctrl+c" {
			m.ActionModel, cmd = m.ActionModel.Update(msg)
		} else {
			m.ListModel, cmd = m.ListModel.Update(msg)
		}
	}

	if m.ActionModel.Quitting {
		return m, tea.Quit
	}
	return m, cmd
}

func (m UnusedVcsRootsModel) View() string {
	if m.ActionModel.Quitting {
		return "Operation complete. Exiting..."
	}
	// Combine views from both models
	return m.ListModel.View() + "\n" + m.ActionModel.View()
}

func (m UnusedVcsRootsModel) IsConfirmed() bool {
	return m.ActionModel.IsConfirmed()
}
