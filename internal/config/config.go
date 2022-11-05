package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
)

// WalgotConfig contains all configuration data.
type WalgotConfig struct {
	CredentialsFile        string
	DefaultListViewUnread  bool
	DefaultListViewStarred bool
	DebugMode              bool
	LogFile                string
	NbEntriesPerAPICall    int
}

// LoadConfig will read a given configJSON file and parses the result, returning a parsed config object
// Stolen from Wallabago code:
// https://github.com/Strubbl/wallabago/blob/master/config.go#L40
func LoadConfig(configJSON string) (WalgotConfig, error) {
	config, err := getConfig(configJSON)
	return config, err
}

// Stolen from Wallabago code:
// https://github.com/Strubbl/wallabago/blob/master/config.go#L40
func getConfig(configJSON string) (config WalgotConfig, err error) {
	raw, err := ioutil.ReadFile(configJSON)
	if err != nil {
		return
	}
	config, err = readJSON(raw)
	return
}

// Stolen from Wallabago code:
// https://github.com/Strubbl/wallabago/blob/master/config.go#L49
// readJSON parses a byte stream into a Walgot Config object
func readJSON(raw []byte) (config WalgotConfig, err error) {
	// trim BOM bytes that make the JSON parser crash
	raw = bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf"))
	err = json.Unmarshal(raw, &config)
	return
}
