package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

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

	switch keyMsg.String() {
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < len(m.UnusedVcsRoots)-1 {
			m.Cursor++
		}
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
