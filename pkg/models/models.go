package models

import (
	tea "github.com/charmbracelet/bubbletea"
)

type ConfirmModel struct {
	Confirmed bool
	Quitting  bool
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
		}
	}
	if m.Quitting {
		return m, tea.Quit
	}
	return m, nil
}

func (m ConfirmModel) View() string {
	if m.Quitting {
		return ""
	}
	return "Are you sure you want to delete all unused VCS roots? (y/n): "
}

func (m ConfirmModel) IsConfirmed() bool {
	return m.Confirmed
}
