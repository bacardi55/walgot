package tui

import (
	"log"
	"strconv"
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
		case "q":
			m.CurrentView = "list"
		}
	}
	return m, nil
}

// Manage update messages for the detail entry view.
func updateEntryView(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// A row has been selected, display article detail:
	case walgotSelectRowMsg:
		m.CurrentView = "detail"
		m.Viewport.SetContent(getDetailViewportContent(m.SelectedID, m.Entries))

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			m.CurrentView = "list"
			// Reset selection.
			m.SelectedID = 0
			// Make sure to scrollback up for other articles:
			m.Viewport.GotoTop()
		case "j", "down":
			//m.Viewport.HalfViewDown()
			m.Viewport.LineDown(1)
		case "k", "up":
			//m.Viewport.HalfViewUp()
			m.Viewport.LineUp(1)
		case "pagedown":
			m.Viewport.HalfViewDown()
		case "pageup":
			m.Viewport.HalfViewUp()
		case "alt+[H":
			m.Viewport.GotoTop()
		case "alt+[F":
			m.Viewport.GotoBottom()

		case "A", "S", "P":
			sID := m.SelectedID
			a, s, p, action := sendEntryUpdate(msg.String(), m.SelectedID, &m)
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
		case "O":
			entry := m.Entries[getSelectedEntryIndex(m.Entries, m.SelectedID)]
			if err := openLinkInBrowser(entry.URL); err != nil {
				m.DialogMessage = "Couldn't open link in browser"
				if m.DebugMode {
					log.Println("Error while opening in browser")
					log.Println(err)
				}
				return m, nil
			}
			m.UpdateMessage = "Link opened in browser"
			return m, tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
				return wallabagoResponseClearMsg(true)
			})
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
		case "u", "s", "a":
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

		// Open in default browser:
		case "O":
			sID, _ := strconv.Atoi(m.Table.SelectedRow()[0])
			entry := m.Entries[getSelectedEntryIndex(m.Entries, sID)]
			url := entry.URL
			// If entry is public, open the public link:
			if entry.IsPublic {
				url = wallabago.Config.WallabagURL + "/share/" + entry.UID
			}

			if err := openLinkInBrowser(url); err != nil {
				m.DialogMessage = "Couldn't open link in browser"
				if m.DebugMode {
					log.Println("Error while opening in browser")
					log.Println(err)
				}
				return m, nil
			}
			m.UpdateMessage = "Link opened in browser"
			return m, tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
				return wallabagoResponseClearMsg(true)
			})
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
		m.Table.SetRows(getTableRows(m.Entries, m.Options.Filters))

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
func updateDialogView(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "esc":
			// Validates dialog, so close it by resetting message:
			m.DialogMessage = ""
		}
	}

	return m, nil
}

// Manage update message for updated entry via API.
func updatedEntryInModel(m *model, updatedEntry wallabago.Item) {
	// Add a message update. No need for a popup here.
	m.UpdateMessage = "Entry has been updated"
	// The entry in the model needs to be updated to avoid refreshing all via API
	m.Entries[getSelectedEntryIndex(m.Entries, updatedEntry.ID)] = updatedEntry
	// Update the table rows so that's it udpated in the list view:
	m.Table.SetRows(getTableRows(m.Entries, m.Options.Filters))
}

// Manage keybinds changing filters on listView.
func listViewFiltersUpdate(msg string, m *model) {
	if msg == "u" {
		m.Options.Filters.Unread = !m.Options.Filters.Unread
		// Unread and Archived can't be selected at the same time:
		if m.Options.Filters.Unread {
			m.Options.Filters.Archived = false
		}
	}
	if msg == "s" {
		m.Options.Filters.Starred = !m.Options.Filters.Starred
	}
	if msg == "a" {
		m.Options.Filters.Archived = !m.Options.Filters.Archived
		// Unread and Archived can't be selected at the same time:
		if m.Options.Filters.Archived {
			m.Options.Filters.Unread = false
		}
	}
	m.Table.SetRows(getTableRows(m.Entries, m.Options.Filters))
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
