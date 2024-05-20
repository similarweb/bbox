package models

import (
	"strings"

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

func (m ConfirmActionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		userInput := strings.ToLower(keyMsg.String())
		if userInput == "y" {
			m.Confirmed = true
			m.Quitting = true
		} else if userInput == "n" || userInput == "ctrl+c" {
			m.Quitting = true
		}
	}

	if m.Quitting {
		return m, tea.Quit
	}

	return m, nil
}

func (m ConfirmActionModel) View() string {
	if !m.Quitting {
		return "Press 'y' to confirm or 'n' to cancel.\n"
	}

	return ""
}

func (m ConfirmActionModel) IsConfirmed() bool {
	return m.Confirmed
}
