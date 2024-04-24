package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type ConfirmModel struct {
	UnusedVcsRoots []string
	cursor         int
	Confirmed      bool
	Quitting       bool
}

func NewConfirmModel(vcsRoots []string) ConfirmModel {
	return ConfirmModel{
		UnusedVcsRoots: vcsRoots,
		cursor:         0, // Start cursor at the first item
		Confirmed:      false,
		Quitting:       false,
	}
}

func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

func (m ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			m.Confirmed = true
			m.Quitting = true
		case "n", "q", "ctrl+c":
			m.Quitting = true
		case "up", "k": // Handle 'up' arrow or 'k' for moving up
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j": // Handle 'down' arrow or 'j' for moving down
			if m.cursor < len(m.UnusedVcsRoots)-1 {
				m.cursor++
			}
		}
	}
	if m.Quitting {
		return m, tea.Quit
	}
	return m, nil
}

func (m ConfirmModel) View() string {
	s := "Are you sure you want to delete the following VCS roots?\n"
	for i, vcsRoot := range m.UnusedVcsRoots {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, vcsRoot)
	}
	s += "\nPress 'y' to confirm, 'n' to cancel, or 'q' to quit."
	return s
}

func (m ConfirmModel) IsConfirmed() bool {
	return m.Confirmed
}
