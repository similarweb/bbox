package models

import (
	"github.com/charmbracelet/bubbletea"
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
