package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"git.bacardi55.io/bacardi55/walgot/cmd"
	"git.bacardi55.io/bacardi55/walgot/internal/config"
	"github.com/mitchellh/go-homedir"
)

// TODO: Read from text file to make it easier to update.
const currentVersion = "0.0.1"
const defaultConfigJSON = "~/.config/walgot/walgot.json"
const defaultCredentialsFile = "~/.config/walgot/credentials.json"
const defaultLogFile = "/tmp/walgot.log"

// ** Init related method ** /
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
	log.Println("Setting log file:", logFile)
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	log.SetOutput(file)
	return nil
}

// Main function.
func main() {
	// Manage command line flags:
	configFile, debugMode := handleFlags()

	// Check walgot configuration file path:
	configFilePath, err := homedir.Expand(*configFile)
	if err != nil {
		fmt.Println("Couldn't find configuration file")
		if *debugMode {
			fmt.Println("Couldn't find configuration file")
		}
		os.Exit(1)
	}
	// Load walgot configuration from Json file:
	walgotConfig, err := config.LoadConfig(configFilePath)
	if err != nil {
		fmt.Println("Error reading config")
		if *debugMode {
			fmt.Println("Error loading Walgot configuration", err.Error())
		}
		os.Exit(1)
	}

	// Configure logs before starting:
	if len(walgotConfig.LogFile) == 0 {
		walgotConfig.LogFile = defaultLogFile
	}
	if err := configLogs(walgotConfig.LogFile); err != nil {
		log.Println("Couldn't configure logs")
		if walgotConfig.DebugMode {
			log.Println(err)
		}
		os.Exit(1)
	}

	log.Println(walgotConfig.CredentialsFile)
	// Load credentials file:
	credentialsFilePath := walgotConfig.CredentialsFile
	if len(credentialsFilePath) == 0 {
		log.Println("Empty credentialsFile config, using default", defaultCredentialsFile)
		walgotConfig.CredentialsFile = defaultCredentialsFile
	}

	credentialsFilePath, err = homedir.Expand(walgotConfig.CredentialsFile)
	if err != nil {
		fmt.Println("Couldn't find credentials file. Check your configuration or use the default place: '~/.config/walgot/credentials.json'")
		if walgotConfig.DebugMode {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	// Check if file exists, otherwise API will fail and that's it:
	_, err = os.Stat(credentialsFilePath)
	if err != nil {
		fmt.Println("Credential file (%v) not found, can't continue", credentialsFilePath)
		if walgotConfig.DebugMode {
			log.Println(err)
		}
		os.Exit(1)
	} else if walgotConfig.DebugMode {
		log.Println("Found credentials file", credentialsFilePath)
	}

	walgotConfig.CredentialsFile = credentialsFilePath

	cmd := cmd.New(walgotConfig)
	cmd.Run()
}
