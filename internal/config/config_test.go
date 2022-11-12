package config

import (
	"testing"
)

func TestReadJson(t *testing.T) {
	var tests = []struct {
		input                          string
		expectedCredentialsFile        string
		expectedLogFile                string
		expectedDefaultSorting         string
		expectedDefaultOrder           string
		expectedDefaultListViewUnread  bool
		expectedDefaultListViewStarred bool
		expectedNbEntriesPerAPICall    int
		expectedDebugMode              bool
		expectedIsErrNil               bool
	}{
		{
			"{\"CredentialsFile\": \"~/.config/walgot/credentials.json\", \"DefaultListViewUnread\": true, \"DefaultListViewStarred\": false, \"DebugMode\": true, \"LogFile\": \"/tmp/walgot.log\", \"NbEntriesPerAPICall\": 255, \"DefaultSorting\": \"created\", \"DefaultOrder\": \"desc\"}",
			"~/.config/walgot/credentials.json",
			"/tmp/walgot.log",
			"created",
			"desc",
			true,
			false,
			255,
			true,
			true,
		},
		{
			"{\"CredentialsFile\": \"~/.config/walgot/credentials.json\"}",
			"~/.config/walgot/credentials.json",
			"",
			"",
			"",
			false,
			false,
			0,
			false,
			true,
		},
		{"", "", "", "", "", false, false, 0, false, false},
	}
	for _, test := range tests {
		var raw = []byte(test.input)
		c, e := readJSON(raw)
		if c.CredentialsFile != test.expectedCredentialsFile {
			t.Errorf("readJson(%v): expectedCredentialsFile %v, got %v", test.input, test.expectedCredentialsFile, c.CredentialsFile)
		}
		if c.LogFile != test.expectedLogFile {
			t.Errorf("readJson(%v): expectedClientId %v, got %v", test.input, test.expectedLogFile, c.LogFile)
		}
		if c.DefaultSorting != test.expectedDefaultSorting {
			t.Errorf("readJson(%v): expectedDefaultSorting %v, got %v", test.input, test.expectedDefaultSorting, c.DefaultSorting)
		}
		if c.DefaultOrder != test.expectedDefaultOrder {
			t.Errorf("readJson(%v): expectedDefaultOrder %v, got %v", test.input, test.expectedDefaultOrder, c.DefaultOrder)
		}
		if c.DefaultListViewUnread != test.expectedDefaultListViewUnread {
			t.Errorf("readJson(%v): expectedDefaultListViewUnread %v, got %v", test.input, test.expectedDefaultListViewUnread, c.DefaultListViewUnread)
		}
		if c.DefaultListViewStarred != test.expectedDefaultListViewStarred {
			t.Errorf("readJson(%v): expectedDefaultListViewStarred %v, got %v", test.input, test.expectedDefaultListViewStarred, c.DefaultListViewStarred)
		}
		if c.NbEntriesPerAPICall != test.expectedNbEntriesPerAPICall {
			t.Errorf("readJson(%v): expectedNbEntriesPerAPICall %v, got %v", test.input, test.expectedNbEntriesPerAPICall, c.NbEntriesPerAPICall)
		}
		if c.DebugMode != test.expectedDebugMode {
			t.Errorf("readJson(%v): expectedDebugMode %v, got %v", test.input, test.expectedDebugMode, c.DebugMode)
		}
		isErrNil := (e == nil)
		if isErrNil != test.expectedIsErrNil {
			t.Errorf("readJson(%v): expectedIsErrNil %v, got %v", test.input, test.expectedIsErrNil, isErrNil)
		}
	}
}
