package commands

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shrijan00003/restler/bin/svc"
	"github.com/shrijan00003/restler/core/env"
)

// init restler project
// init command should be able to set the RESTLER_PATH in .env file which will be loaded by restler.
// it should be creating default files and folders in the path.
type textInputModel struct {
	textInput textinput.Model
	err       error
}

type (
	errMsg error
)

func initialTextInputModel() textInputModel {
	ti := textinput.New()
	ti.Placeholder = "restler"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return textInputModel{
		textInput: ti,
		err:       nil,
	}
}

func (m textInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m textInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			executeInitCommand(m.textInput.Value())
			return m, tea.Quit
		}
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m textInputModel) View() string {
	return fmt.Sprintf(
		"Where do you want to initialize your restler project? \n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}

func initRestlerProject() error {
	p := tea.NewProgram(initialTextInputModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error occurred while initializing restler project: ", err)
		return err
	}
	return nil
}

// TODO: Will download the sample folder from github repo instead of creating each one one by one
func executeInitCommand(path string) error {
	// if path exists, thats it, otherwise ask if user wants to create it
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("[log]: Path doesn't exist, creating restler project in: ", path)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			fmt.Println("[error]: Error occurred while creating restler project: ", err)
			return err
		}
		err = svc.CreateDefaultFile(path)
		if err != nil {
			fmt.Println("[error]: Error occurred while creating default files: ", err)
			return err
		}
		env.UpdateEnv(path)
		return nil
	} else {
		fmt.Println("[info]: path exists, updating RESTLER_PATH env: ")
		env.UpdateEnv(path)
		return nil
	}

}
