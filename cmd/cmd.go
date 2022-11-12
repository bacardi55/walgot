package cmd

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"git.bacardi55.io/bacardi55/walgot/internal/api"
	"git.bacardi55.io/bacardi55/walgot/internal/config"
	"git.bacardi55.io/bacardi55/walgot/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mitchellh/go-homedir"
)

// Default config:
const currentVersion = "0.0.1"
const defaultConfigJSON = "~/.config/walgot/walgot.json"
const defaultCredentialsFile = "~/.config/walgot/credentials.json"
const defaultLogFile = "/tmp/walgot.log"
const defaultNbEntriesPerAPICall = 250

// WalgotCmd contains command data.
type WalgotCmd struct {
	config     config.WalgotConfig
	teaProgram *tea.Program
}

// New returns a WalgotCmd.
func New() *WalgotCmd {
	return &WalgotCmd{}
}

// Init initialize the application.
func Init() (*WalgotCmd, error) {
	// Manage command line flags:
	configFile, debugMode := handleFlags()

	// Check walgot configuration file path:
	configFilePath, err := homedir.Expand(*configFile)
	if err != nil {
		if *debugMode {
			fmt.Println("Couldn't find configuration file")
		}
		return New(), errors.New("couldn't find configuration file")
	}

	// Load walgot configuration from Json file:
	walgotConfig, err := config.LoadConfig(configFilePath)
	if err != nil {
		if *debugMode {
			fmt.Println("Error loading Walgot configuration", err.Error())
		}
		return &WalgotCmd{}, errors.New("couldn't load walgot configuration")
	}

	// Configure logs before starting:
	if len(walgotConfig.LogFile) == 0 {
		walgotConfig.LogFile = defaultLogFile
	}
	if err := configLogs(walgotConfig.LogFile); err != nil {
		if walgotConfig.DebugMode {
			log.Println(err)
		}
		return &WalgotCmd{}, errors.New("error configuring logs")
	}

	// Load credentials file:
	credentialsFilePath := walgotConfig.CredentialsFile
	if len(credentialsFilePath) == 0 {
		log.Println("Empty credentialsFile config, using default", defaultCredentialsFile)
		walgotConfig.CredentialsFile = defaultCredentialsFile
	}
	credentialsFilePath, err = homedir.Expand(walgotConfig.CredentialsFile)
	if err != nil {
		if walgotConfig.DebugMode {
			fmt.Println(err)
		}
		return &WalgotCmd{}, errors.New("couldn't determine path for credentials file")
	}
	// Check if file exists, otherwise API will fail and that's it:
	_, err = os.Stat(credentialsFilePath)
	if err != nil {
		if walgotConfig.DebugMode {
			log.Println(err)
		}
		return &WalgotCmd{}, errors.New("couldn't find credentials file")
	} else if walgotConfig.DebugMode {
		log.Println("Found credentials file", credentialsFilePath)
	}

	walgotConfig.CredentialsFile = credentialsFilePath

	// If NbEntriesPerAPICall is not set:
	if walgotConfig.NbEntriesPerAPICall <= 0 {
		walgotConfig.NbEntriesPerAPICall = defaultNbEntriesPerAPICall
	}

	// Initialize wallabago:
	api.InitWallabagoAPI(walgotConfig.CredentialsFile)

	// Create bubbletea program:
	p := tea.NewProgram(
		tui.NewModel(walgotConfig),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	return &WalgotCmd{
		config:     walgotConfig,
		teaProgram: p,
	}, nil
}

// Run starts the application.
func (cmd WalgotCmd) Run() {
	if err := cmd.teaProgram.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// Manage debug flags.
func handleFlags() (*string, *bool) {
	var (
		version    = flag.Bool("version", false, "get walgot version")
		debug      = flag.Bool("d", false, "enable debug output")
		configJSON = flag.String("config", defaultConfigJSON, "file name of config JSON file")
	)
	flag.Parse()
	if *version {
		fmt.Println("Walgot version:", currentVersion)
		os.Exit(1)
	}
	if *debug {
		fmt.Println("handleFlags: debug mode")
	}

	return configJSON, debug
}

// Manage log configuration.
func configLogs(logFile string) error {
	fmt.Println("Setting log file:", logFile)
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	log.SetOutput(file)
	return nil
}
