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
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "y" {
			m.Confirmed = true
			m.Quitting = true
		} else if keyMsg.String() == "n" || keyMsg.String() == "q" || keyMsg.String() == "ctrl+c" {
			m.Quitting = true
		}

	}
	return m, nil
}

func (m ConfirmActionModel) View() string {
	return "Press 'y' to confirm, 'n' to cancel, or 'q' to quit.\n"
}

func (m ConfirmActionModel) IsConfirmed() bool {
	return m.Confirmed
}

// ListModel handles listing and navigation.
type ListModel struct {
	UnusedVcsRoots []string
	ListMsg        string
	Cursor         int
}

func NewListModel(vcsRoots []string, listMsg string) ListModel {
	return ListModel{
		UnusedVcsRoots: vcsRoots,
		ListMsg:        listMsg,
		Cursor:         0,
	}
}

func (m ListModel) Init() tea.Cmd {
	return nil
}

func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil // Early return if msg is not a KeyMsg
	}

	if keyMsg.String() == "up" || keyMsg.String() == "k" {
		if m.Cursor > 0 {
			m.Cursor--
		}
		return m, nil
	}

	if keyMsg.String() == "down" || keyMsg.String() == "j" {
		if m.Cursor < len(m.UnusedVcsRoots)-1 {
			m.Cursor++
		}

		return m, nil
	}

	return m, nil
}

func (m ListModel) View() string {
	s := "\n" + m.ListMsg + "\n\n"
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

func NewUnusedVcsRootsModel(vcsRoots []string, listMsg string) UnusedVcsRootsModel {
	return UnusedVcsRootsModel{
		ActionModel: NewConfirmActionModel(),
		ListModel:   NewListModel(vcsRoots, listMsg),
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

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "y" || keyMsg.String() == "n" || keyMsg.String() == "q" || keyMsg.String() == "ctrl+c" {
			m.ActionModel, cmd = m.ActionModel.Update(keyMsg)
		} else {
			m.ListModel, cmd = m.ListModel.Update(keyMsg)
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

	return m.ListModel.View() + "\n" + m.ActionModel.View()
}

func (m UnusedVcsRootsModel) IsConfirmed() bool {

	return m.ActionModel.IsConfirmed()
}
