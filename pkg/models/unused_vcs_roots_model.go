package models

import (
	tea "github.com/charmbracelet/bubbletea"
)

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
