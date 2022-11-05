package cmd

import (
	"fmt"
	"os"

	"git.bacardi55.io/bacardi55/walgot/internal/api"
	"git.bacardi55.io/bacardi55/walgot/internal/config"
	"git.bacardi55.io/bacardi55/walgot/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

// WalgotCmd contains command data.
type WalgotCmd struct {
	config     config.WalgotConfig
	teaProgram *tea.Program
}

// Run starts the application.
func (cmd WalgotCmd) Run() {
	if err := cmd.teaProgram.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// New returns a WalgotCmd.
func New(config config.WalgotConfig) WalgotCmd {
	api.InitWallabagoAPI(config.CredentialsFile)
	p := tea.NewProgram(
		tui.NewModel(config),
		//initialModel(config),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	return WalgotCmd{
		config:     config,
		teaProgram: p,
	}
}
