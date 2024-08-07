package tui

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Strubbl/wallabago/v7"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Manage update messages on the help view.
func updateHelpView(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			m.CurrentView = "list"
		}
	}
	return m, nil
}

// Manage update messages for the detail entry view.
func updateEntryView(msg tea.Msg, m *model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// A row has been selected, display article detail:
	case walgotSelectRowMsg:
		m.CurrentView = "detail"
		m.Viewport.SetContent(getDetailViewportContent(m.SelectedID, m.Entries, m.TermSize.Width))

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			m.CurrentView = "list"
			// Reset selection.
			m.SelectedID = 0
			// Make sure to scrollback up for other articles:
			m.Viewport.GotoTop()
		case "j", "down":
			m.Viewport.LineDown(1)
		case "k", "up":
			m.Viewport.LineUp(1)
		case "pagedown":
			m.Viewport.HalfViewDown()
		case "pageup":
			m.Viewport.HalfViewUp()
		case "alt+[H":
			m.Viewport.GotoTop()
		case "alt+[F":
			m.Viewport.GotoBottom()

		// Update article (archive, starred, public):
		case "A", "S", "P":
			sID := m.SelectedID
			a, s, p, action := sendEntryUpdate(msg.String(), m.SelectedID, m)
			if m.DebugMode {
				log.Println("Update entry action:", action, a, s)
			}
			m.UpdateMessage = action
			return m, requestWallabagEntryUpdate(
				sID,
				a,
				s,
				p,
			)

		// Open links in entry:
		case "L":
			// Configure textinput:
			m.Dialog.TextInput.Placeholder = "Link number"
			m.Dialog.TextInput.CharLimit = 3
			// Display textinput
			m.Dialog.ShowInput = true
			// Add search button:
			m.Dialog.Action = "open link"
			// Dialog title:
			m.Dialog.Message = "Enter the link number to open:\n"
			// Set current view to dialog:
			m.CurrentView = "dialog"

		// Open or Copy URL:
		case "O", "Y":
			entry := &m.Entries[getSelectedEntryIndex(m.Entries, m.SelectedID)]
			url := entry.URL
			// If entry is public, open the public link:
			if entry.IsPublic {
				url = wallabago.Config.WallabagURL + "/share/" + entry.UID
			}

			if msg.String() == "O" {
				// Open URL in browser:
				if err := openLinkInBrowser(url); err != nil {
					m.Dialog.Message = "Couldn't open link in browser"
					if m.DebugMode {
						log.Println("Error while opening in browser")
						log.Println(err)
					}
					return m, nil
				}
				m.UpdateMessage = "Link opened in browser"
			} else if msg.String() == "Y" {
				// Copy URL:
				if err := copyLinkToClipboard(url); err != nil {
					m.Dialog.Message = "Couldn't copy link"
					if m.DebugMode {
						log.Println("Error while copying link")
						log.Println(err)
					}
					return m, nil
				}
				m.UpdateMessage = "Link copied"
			}

			return m, tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
				return wallabagoResponseClearMsg(true)
			})

		// Delete:
		case "D":
			sID := m.SelectedID
			m.SelectedID = 0
			m.CurrentView = "list"
			return m, requestWallabagEntryDelete(sID)
		}
	}

	m.Viewport, cmd = m.Viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// Manage update messages for the list view.
func updateListView(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if !m.Reloading {
				sID, _ := strconv.Atoi(m.Table.SelectedRow()[0])
				return m, selectEntryCommand(sID)
			}
		case "j", "down":
			m.Table.MoveDown(1)
		case "pgdown":
			m.Table.MoveDown(10)
		case "k", "up":
			m.Table.MoveUp(1)
		case "pgup":
			m.Table.MoveUp(10)
		case "alt+[H":
			m.Table.GotoTop()
		case "alt+[F":
			m.Table.GotoBottom()
		case "q":
			// If search active, clean it and don't quit:
			if m.Options.Filters.Search != "" {
				return m, func() tea.Msg {
					return walgotSearchEntryMsg("")
				}
			}
			return m, tea.Quit
		case "r":
			// If already reloading, do nothing
			if m.Reloading {
				return m, nil
			}
			// Status as reloading:
			m.Reloading = true
			// Reset number of entries:
			m.TotalEntriesOnServer = 0
			return m, requestWallabagNbEntries

		// Filters for the table list:
		case "u", "s", "a", "p":
			listViewFiltersUpdate(msg.String(), &m)

		// Update entry status:
		case "A", "S", "P":
			sID, _ := strconv.Atoi(m.Table.SelectedRow()[0])
			a, s, p, action := sendEntryUpdate(msg.String(), sID, &m)
			if m.DebugMode {
				log.Println("Update entry action:", action, a, s)
			}
			m.UpdateMessage = action
			return m, requestWallabagEntryUpdate(
				sID,
				a,
				s,
				p,
			)

		// Open or Copy URL:
		case "O", "Y":
			sID, _ := strconv.Atoi(m.Table.SelectedRow()[0])
			entry := m.Entries[getSelectedEntryIndex(m.Entries, sID)]
			url := entry.URL
			// If entry is public, open the public link:
			if entry.IsPublic {
				url = wallabago.Config.WallabagURL + "/share/" + entry.UID
			}

			if msg.String() == "O" {
				// Open URL in browser:
				if err := openLinkInBrowser(url); err != nil {
					m.Dialog.Message = "Couldn't open link in browser"
					if m.DebugMode {
						log.Println("Error while opening in browser")
						log.Println(err)
					}
					return m, nil
				}
				m.UpdateMessage = "Link opened in browser"
			} else if msg.String() == "Y" {
				// Copy URL:
				if err := copyLinkToClipboard(url); err != nil {
					m.Dialog.Message = "Couldn't copy link"
					if m.DebugMode {
						log.Println("Error while copying link")
						log.Println(err)
					}
					return m, nil
				}
				m.UpdateMessage = "Link copied"
			}

			return m, tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
				return wallabagoResponseClearMsg(true)
			})

		// Delete:
		case "D":
			if m.Reloading {
				return m, nil
			}
			sID, _ := strconv.Atoi(m.Table.SelectedRow()[0])
			return m, requestWallabagEntryDelete(sID)

		// Search:
		case "/":
			if m.Reloading {
				return m, nil
			}
			// Configure textinput:
			m.Dialog.TextInput.Placeholder = "Search"
			m.Dialog.TextInput.CharLimit = 55
			// Display textinput
			m.Dialog.ShowInput = true
			// Add search button:
			m.Dialog.Action = "search"
			// Dialog title:
			m.Dialog.Message = "Filter by article's title:\n"
			// Set current view to dialog:
			m.CurrentView = "dialog"

		// Add an entry:
		case "N":
			if m.Reloading {
				return m, nil
			}
			// Configure textinput:
			m.Dialog.TextInput.Placeholder = "URL"
			m.Dialog.TextInput.CharLimit = 0
			// Display textinput
			m.Dialog.ShowInput = true
			// Add search button:
			m.Dialog.Action = "add"
			// Dialog title:
			m.Dialog.Message = "Add a URL to wallabag:\n"
			// Set current view to dialog:
			m.CurrentView = "dialog"

		// Clean, if needed:
		case "esc":
			if m.Options.Filters.Search != "" {
				// Cleaning a search.
				m.Options.Filters.Search = ""
				// Table needs to be refreshed:
				return m, func() tea.Msg {
					return walgotSearchEntryMsg("")
				}
			}
		}

	// When resizing the window, sizes needs to change everywhere…
	case tea.WindowSizeMsg:
		m.TermSize = termSize{msg.Width, msg.Height}
		// TODO: Seems to bug when resizing though:
		windowSizeUpdate(&m)

	// Retrieved total number of entities from API:
	case wallabagoResponseNbEntitiesMsg:
		m.TotalEntriesOnServer = int(msg)
		// We now have the number of entries, we can trigger
		// the process to retrieve all these entries
		return m, tea.Batch(
			requestWallabagEntries(
				m.TotalEntriesOnServer,
				m.NbEntriesPerAPICall,
				m.Options.Sorts.Field,
				m.Options.Sorts.Order,
			),
			m.Spinner.Tick,
		)

	// Retrieved entities from API, data has changed:
	case wallabagoResponseEntitiesMsg:
		// Response received, we are not reloading anymore:
		m.Reloading = false
		m.Entries = msg
		if m.DebugMode {
			log.Println("wallabagoResponseEntityMsg", len(msg))
		}
		m.Table.SetRows(getTableRows(m.Entries, m.Options.Filters, m.TermSize.Width))

	// Added entry response:
	case wallabagoResponseAddEntryMsg:
		// Add new entry at the top.
		m.Entries = append([]wallabago.Item{msg.Entry}, m.Entries...)
		// Recalculate table rows:
		m.Table.SetRows(getTableRows(m.Entries, m.Options.Filters, m.TermSize.Width))
		// Wallabag API send a 200 even if the URL isn't good.
		// Unfortunately, it means checking the content of the entry…
		if strings.Contains(
			msg.Entry.Content,
			"wallabag can't retrieve contents for this article",
		) {
			m.Dialog.Message = "Wallabag couldn't retrieve content, empty entry created."
		} else {
			// Letting user know:
			m.UpdateMessage = "Entry has been added successfully"
			return m, tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
				return wallabagoResponseClearMsg(true)
			})
		}

	// Deleted entry response:
	case wallabagoResponseDeleteEntryMsg:
		// Remove deleted entry from model:
		index := getSelectedEntryIndex(m.Entries, int(msg))
		if index == 0 {
			m.Entries = m.Entries[1:]
		} else if index == len(m.Entries)-1 {
			m.Entries = m.Entries[:len(m.Entries)-2]
		} else {
			m.Entries = append(m.Entries[:index], m.Entries[index+1:]...)
		}
		m.Table.SetRows(getTableRows(m.Entries, m.Options.Filters, m.TermSize.Width))
		// Letting user know:
		m.UpdateMessage = "Entry has been deleted successfully"
		return m, tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
			return wallabagoResponseClearMsg(true)
		})

	// Search request:
	case walgotSearchEntryMsg:
		m.Options.Filters.Search = string(msg)
		// Recalculate table rows:
		m.Table.SetRows(getTableRows(m.Entries, m.Options.Filters, m.TermSize.Width))

	case spinner.TickMsg:
		// Spin only if it is still displaying the reload screen:
		if m.Reloading {
			m.Spinner, cmd = m.Spinner.Update(msg)
			return m, cmd
		}
	}

	return m, cmd
}

// Manage update messages for dialog view.
func updateDialogView(msg tea.Msg, m *model) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Close and reset dialog box:
			m.Dialog.Message = ""
			m.Dialog.ShowInput = false
			m.Dialog.Action = ""
			m.Dialog.TextInput.Blur()
			// Search input is not resetted though, just in case.
			return m, nil

		case "enter":
			input := m.Dialog.TextInput.Value()
			action := m.Dialog.Action
			// Cleaning dialog box:
			m.Dialog.Message = ""
			m.Dialog.ShowInput = false
			m.Dialog.Action = ""
			m.Dialog.TextInput.Blur()
			m.Dialog.TextInput.Reset()
			// Next screen should be on filtered list:
			m.CurrentView = "list"

			// Per action command:
			switch action {
			case "search":
				// Start search, value needs to be copied.
				cmds = append(cmds, func() tea.Msg {
					return walgotSearchEntryMsg(input)
				})

			// Save entry:
			case "add":
				return m, requestWallabagAddEntry(input)

			case "open link":
				_, links := getCleanedContentAndLinks(
					m.Entries[getSelectedEntryIndex(m.Entries, m.SelectedID)].Content,
				)
				selected, err := strconv.Atoi(input)
				if err != nil {
					m.Dialog.Message = "Couldn't find link number " + input
					if m.DebugMode {
						log.Println("Error while opening link")
						log.Println(err)
					}
					return m, nil
				}

				if selected < 1 || selected > len(links) {
					m.Dialog.Message = "No link numbered " + input
					if m.DebugMode {
						log.Println("Error while opening link")
						log.Println(err)
					}
					return m, nil
				}

				if err := openLinkInBrowser(links[selected-1]); err != nil {
					m.Dialog.Message = "Couldn't open link"
					if m.DebugMode {
						log.Println("Error while opening link")
						log.Println(err)
					}
					return m, nil
				}
				m.UpdateMessage = "Link opened"

				cmds = append(
					cmds,
					tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
						return wallabagoResponseClearMsg(true)
					}),
				)
			}
		}
	}

	m.Dialog.TextInput.Focus()
	var cmd tea.Cmd
	m.Dialog.TextInput, cmd = m.Dialog.TextInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// Manage update message for updated entry via API.
func updatedEntryInModel(m *model, updatedEntry wallabago.Item) {
	// Add a message update. No need for a popup here.
	m.UpdateMessage = "Entry has been updated"
	// The entry in the model needs to be updated to avoid refreshing all via API
	m.Entries[getSelectedEntryIndex(m.Entries, updatedEntry.ID)] = updatedEntry
	// Update the table rows so that's it udpated in the list view:
	m.Table.SetRows(getTableRows(m.Entries, m.Options.Filters, m.TermSize.Width))
}

// Manage keybinds changing filters on listView.
func listViewFiltersUpdate(msg string, m *model) {
	if msg == "u" {
		m.Options.Filters.Unread = !m.Options.Filters.Unread
		// Unread and Archived can't be selected at the same time:
		if m.Options.Filters.Unread {
			m.Options.Filters.Archived = false
		}
	} else if msg == "a" {
		m.Options.Filters.Archived = !m.Options.Filters.Archived
		// Unread and Archived can't be selected at the same time:
		if m.Options.Filters.Archived {
			m.Options.Filters.Unread = false
		}
	} else if msg == "s" {
		m.Options.Filters.Starred = !m.Options.Filters.Starred
	} else if msg == "p" {
		m.Options.Filters.Public = !m.Options.Filters.Public
	}

	m.Table.SetRows(getTableRows(m.Entries, m.Options.Filters, m.TermSize.Width))
}

// Retrieve updates variable.
func sendEntryUpdate(msg string, sID int, m *model) (int, int, int, string) {
	entry := m.Entries[getSelectedEntryIndex(m.Entries, sID)]
	action := "Toggled entry status: "
	a := entry.IsArchived
	s := entry.IsStarred
	p := 0
	if entry.IsPublic {
		p = 1
	}

	if msg == "A" {
		if entry.IsArchived == 0 {
			action = "archive"
			a = 1
		} else {
			action = "read"
			a = 0
		}
	} else if msg == "S" {
		if entry.IsStarred == 0 {
			action = "starred"
			s = 1
		} else {
			action = "unstarred"
			s = 0
		}
	} else if msg == "P" {
		if !entry.IsPublic {
			action = "publish"
			p = 1
		} else {
			action = "unpublish"
			p = 0
		}
	}

	return a, s, p, action
}
