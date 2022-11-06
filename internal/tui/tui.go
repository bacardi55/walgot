package tui

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"git.bacardi55.io/bacardi55/walgot/internal/api"
	"git.bacardi55.io/bacardi55/walgot/internal/config"

	"github.com/Strubbl/wallabago/v7"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"

	"github.com/k3a/html2text"
)

// ** Model related Struct ** //

// Terminal physical size:
type termSize struct {
	Width  int
	Height int
}

// TableView filter options
type walgotTableFilters struct {
	Archived bool
	Starred  bool
	Unread   bool
}

// TableView Sort options
type walgotTableSorts struct {
	Field string
	Order string
}

type walgotTableOptions struct {
	Filters walgotTableFilters
	Sorts   walgotTableSorts
}

// Model structure
type model struct {
	// Sub models related:
	Table         table.Model
	Viewport      viewport.Model
	DialogMessage string
	Spinner       spinner.Model
	UpdateMessage string
	// Tui Status related
	Ready       bool
	Reloading   bool
	CurrentView string
	Options     walgotTableOptions
	// Wallabag(o) related:
	Entries              []wallabago.Item
	SelectedID           int
	TotalEntriesOnServer int
	// Configs
	NbEntriesPerAPICall int
	TermSize            termSize
	DebugMode           bool
}

// NewModel returns default model for walgot.
func NewModel(config config.WalgotConfig) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.
		NewStyle().
		Foreground(lipgloss.Color("205"))

	return model{
		SelectedID:           0,
		Ready:                false,
		Reloading:            true,
		CurrentView:          "list",
		TotalEntriesOnServer: 0,
		Spinner:              s,
		NbEntriesPerAPICall:  config.NbEntriesPerAPICall,
		DebugMode:            config.DebugMode,
		Options: walgotTableOptions{
			Filters: walgotTableFilters{
				Unread:  config.DefaultListViewUnread,
				Starred: config.DefaultListViewStarred,
			},
			Sorts: walgotTableSorts{
				Field: "created",
				Order: "desc",
			},
		},
	}
}

// Response message for number of entities from Wallabago
type wallabagoResponseNbEntitiesMsg int

// Response message for all entities from Wallabago.
type wallabagoResponseEntitiesMsg []wallabago.Item

// Response message for entity update.
type wallabagoResponseEntityUpdateMsg struct {
	UpdatedEntry wallabago.Item
}

// After update message has been displayed enough time.
type wallabagoResponseClearMsg bool

type wallabagoResponseErrorMsg struct {
	message        string
	wallabagoError error
}

// Selected row in table list Message.
type walgotSelectRowMsg int

// Callback for requesting the total number of entries via API.
func requestWallabagNbEntries() tea.Msg {
	// Get total number of articles:
	nbArticles, e := api.GetNbTotalEntries()

	if e != nil {
		return wallabagoResponseErrorMsg{
			message:        "Error:\n couldn't retrieve the total number of entries from wallabag API",
			wallabagoError: e,
		}
	}

	return wallabagoResponseNbEntitiesMsg(nbArticles)
}

// Callback for requesting entries via API.
func requestWallabagEntries(nbArticles, nbEntriesPerAPICall int, sortField, sortOrder string) tea.Cmd {
	return func() tea.Msg {
		limitArticleByAPICall := nbEntriesPerAPICall
		nbCalls := 1
		if nbArticles > limitArticleByAPICall {
			nbCalls = nbArticles / limitArticleByAPICall
			if float64(nbCalls) < float64(nbArticles)/float64(limitArticleByAPICall) {
				nbCalls++
			}
		}

		// TODO: Move this to async channel?
		// Might not be a good idea with the ELM architecture?
		var entries []wallabago.Item
		for i := 1; i < nbCalls+1; i++ {
			r, err := api.GetEntries(limitArticleByAPICall, i, sortField, sortOrder)

			if err != nil {
				return wallabagoResponseErrorMsg{
					message:        "Error:\n couldn't retrieve the entries from wallabag API",
					wallabagoError: err,
				}
			}

			entries = append(entries, r.Embedded.Items...)
		}

		return wallabagoResponseEntitiesMsg(entries)
	}
}

// Callback for updating an entry status via API.
func requestWallabagEntryUpdate(entryID, archive, starred int) tea.Cmd {
	return func() tea.Msg {
		// Send PATCH via API:
		r, err := api.UpdateEntry(entryID, archive, starred)
		if err != nil {
			return wallabagoResponseErrorMsg{
				message:        "Error:\n Couldn't update the selected entry",
				wallabagoError: err,
			}
		}

		var item wallabago.Item
		err = json.Unmarshal(r, &item)
		if err != nil {
			return wallabagoResponseErrorMsg{
				message:        "Error:\n Response from wallabago is not valid",
				wallabagoError: err,
			}
		}

		return wallabagoResponseEntityUpdateMsg{item}
	}
}

// Callback for selecting entry in list:
func selectEntryCommand(selectedRowID int) tea.Cmd {
	return func() tea.Msg {
		return walgotSelectRowMsg(selectedRowID)
	}
}

// ** Model related methods ** //
// Init method.
func (m model) Init() tea.Cmd {
	//wallabago.ReadConfig(m.WallabagConfig )

	return tea.Batch(
		requestWallabagNbEntries,
		m.Spinner.Tick,
	)
}

// Update method.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.DebugMode {
		log.Println(fmt.Sprintf("Update message received, type: %T", msg, m.CurrentView))
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		// C-c to kill the app.
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		} else if msg.String() == "?" && !m.Reloading {
			m.CurrentView = "help"
			return m, nil
		}
	}

	// Priority: Error > updates > entrySelection:
	if v, ok := msg.(wallabagoResponseErrorMsg); ok {
		m.Reloading = false
		if m.DebugMode {
			log.Println("Wallabago error:")
			log.Println(v.wallabagoError)
		}
		m.DialogMessage = v.message
	} else if v, ok := msg.(wallabagoResponseEntityUpdateMsg); ok {
		// If received an entry update response message,
		// the model needs to be updated with refreshed entry:
		updatedEntryInModel(&m, v.UpdatedEntry)
		// To remove the update message after 3 seconds:
		return m, tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
			return wallabagoResponseClearMsg(true)
		})
	} else if v, ok := msg.(wallabagoResponseClearMsg); ok && bool(v) {
		// Clear update message
		m.UpdateMessage = ""
	} else if v, ok := msg.(walgotSelectRowMsg); ok {
		// This needs to happen before sending to the sub update function.
		m.SelectedID = int(v)
	}

	// Priority order: dialog > help > detail > list.
	if m.DialogMessage != "" {
		return updateDialogView(msg, m)
	} else if m.CurrentView == "help" {
		return updateHelpView(msg, m)
	}

	// Now send to the right sub-update function:
	if m.SelectedID > 0 {
		return updateEntryView(msg, m)
	}
	return updateListView(msg, m)
}

// ** Update related functions ** //
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
			m.Viewport.HalfViewDown()
		case "k", "up":
			m.Viewport.HalfViewUp()
		case "A", "S":
			sID := m.SelectedID
			a, s, action := sendEntryUpdate(msg.String(), m.SelectedID, &m)
			if m.DebugMode {
				log.Println("Update entry action:", action, a, s)
			}
			m.UpdateMessage = action
			return m, requestWallabagEntryUpdate(sID, a, s)
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
			sID, _ := strconv.Atoi(m.Table.SelectedRow()[0])
			return m, selectEntryCommand(sID)
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
		case "A", "S":
			sID, _ := strconv.Atoi(m.Table.SelectedRow()[0])
			a, s, action := sendEntryUpdate(msg.String(), sID, &m)
			if m.DebugMode {
				log.Println("Update entry action:", action, a, s)
			}
			m.UpdateMessage = action
			return m, requestWallabagEntryUpdate(sID, a, s)

		case "O":
			sID, _ := strconv.Atoi(m.Table.SelectedRow()[0])
			entry := m.Entries[getSelectedEntryIndex(m.Entries, sID)]
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

// View method.
func (m model) View() string {
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.mainView(), m.footerView())
}

// ** View related functions. ** //
// Return the header part of the view.
func (m model) headerView() string {
	titleStyle := lipgloss.
		NewStyle().
		Width(m.TermSize.Width).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Align(lipgloss.Center)

	subtitle := ""
	if !m.Reloading && m.Ready {
		if m.Options.Filters.Unread {
			subtitle += " - Unread"
		}
		if m.Options.Filters.Starred {
			subtitle += " - Starred"
		}
		if m.Options.Filters.Archived {
			subtitle += " - Archived"
		}
	}

	if len(subtitle) == 0 && !m.Reloading {
		subtitle = " - All"
	}

	t := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Bold(true).Render("Walgot"),
		lipgloss.NewStyle().Render(subtitle),
	)

	return titleStyle.Render(t)
}

// Return the footer part of the view.
func (m model) footerView() string {
	var text string

	if len(m.UpdateMessage) > 0 {
		text += lipgloss.NewStyle().Italic(true).Render(m.UpdateMessage)
	} else if !m.Reloading {
		text += lipgloss.
			NewStyle().
			Bold(true).
			Render(strconv.Itoa(m.TotalEntriesOnServer))
		text += " articles loaded from wallabag"
	}

	text += "\n[r]eload -- Toggles: [u]nread, [s]tarred, [a]rchived -- [h]elp"

	return lipgloss.
		NewStyle().
		Width(m.TermSize.Width).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		Align(lipgloss.Center).
		PaddingTop(0).
		Render(text)
}

// Return the main part of the view.
func (m model) mainView() string {
	if !m.Ready {
		// Not initialized yet, let's not style it.
		return "\n   Initializing…"
	}
	if m.Reloading {
		return reloadingView(m)
	}

	// Priority: dialog > help > detail > list.
	if m.DialogMessage != "" {
		return dialogView(m)
	} else if m.CurrentView == "help" {
		return helpView(m)
	} else if m.SelectedID > 0 {
		return entryDetailView(m)
	}
	return listView(m)
}

// Manage window size changes.
func windowSizeUpdate(m *model) {
	h := m.TermSize.Height - lipgloss.Height(m.headerView()) - lipgloss.Height(m.footerView())
	// Regenerate the table based on new size:
	t := createViewTable(m.TermSize.Width, h-5)
	if m.Ready {
		m.Table.SetRows(getTableRows(m.Entries, m.Options.Filters))
	}
	m.Table = t
	// Generate viewport based on screen size
	contentWidth := 80
	if m.TermSize.Width < 80 {
		contentWidth = m.TermSize.Width
	}
	v := viewport.New(contentWidth, h-5)

	// We recieved terminal size, we are ready:
	m.Ready = true
	// Saving viewport in model:
	m.Viewport = v
}

// Manage reloading view.
func reloadingView(m model) string {
	text := "Loading all"
	if m.TotalEntriesOnServer > 0 {
		text += " " + strconv.Itoa(m.TotalEntriesOnServer)
	}
	text += " entries from wallabag…"
	if m.TotalEntriesOnServer > 250 {
		text += " (This can take a few moment…)"
	}

	return lipgloss.NewStyle().
		Width(m.TermSize.Width).
		Align(lipgloss.Center).
		Render(m.Spinner.View() + text)
}

// Help view.
func helpView(m model) string {
	text := []byte(`Help:
  On all screens:
  - ctrl+c: Quit
  - h: Help (this page)

  On listing page:
  - r: Reload article from wallabag via APIs, takes time depending on the number of articles saved
  - u: Toggle display only unread articles (disable archived filter)
  - s: Toggle display only starred articles
  - a: Toggle archived only articles (disable unread filter)
  - A: Toggle Archive / Unread for the current article (and update wallabag backend)
  - S: Toggle Starred / Unstarred for the current article (and update wallabag backend)
  - O: Open article link url in default browser
  - h: Display help
  - ↑ or k / ↓ or j: Move up / down one item in the list
  - page down / page up: Move up / down 10 items in the list
  - home: Go to the top of the list
  - end: Go to bottom of the list
  - enter: Select entry to read content
  - q: quit

  On detail page:
  - A: Toggle Archive / Unread for the current article (and update wallabag backend)
  - S: Toggle Starred / Unstarred for the current article (and update wallabag backend)
  - O: Open article link url in default browser
  - q: Return to list
  - ↑ or k / ↓ or j: Go up / down

  On dialog (modal) view:
  - "enter" or "esc": Close the dialog

  On help page:
  - q: Return to list
`)

	return lipgloss.
		NewStyle().
		Width(m.TermSize.Width).
		Align(lipgloss.Left).
		Render(string(text))
}

// Get article detail view.
func entryDetailView(m model) string {
	header := entryDetailViewTitle(
		m.Entries[getSelectedEntryIndex(m.Entries, m.SelectedID)],
	)
	footer := entryDetailViewFooter(m.Viewport)

	return lipgloss.
		NewStyle().
		Width(m.TermSize.Width).
		Align(lipgloss.Center).
		Render(header + "\n" + m.Viewport.View() + "\n" + footer)
}

// Retrieve title for detail view.
func entryDetailViewTitle(entry wallabago.Item) string {
	return lipgloss.
		NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Bold(true).
		Width(80).
		Align(lipgloss.Left).
		Render(wordwrap.String(entry.Title, 72))
}

// Retrieve footer for detail view.
func entryDetailViewFooter(viewport viewport.Model) string {
	info := lipgloss.
		NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderRight(true).
		Render(fmt.Sprintf("%3.f%%", viewport.ScrollPercent()*100))

	width := viewport.Width - lipgloss.Width(info)
	if width < 0 {
		width = 0
	}
	line := strings.Repeat("─", width)

	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

// Get list view.
func listView(m model) string {
	return m.Table.View()
}

// Get dialog view.
func dialogView(m model) string {
	dialogBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 0).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)

	okButton := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFF7DB")).
		Background(lipgloss.Color("#888B7E")).
		Padding(0, 3).
		MarginTop(1).
		Underline(true).
		Render("Ok")

	question := lipgloss.
		NewStyle().
		Width(50).
		Align(lipgloss.Center).
		Render(m.DialogMessage)

	ui := lipgloss.JoinVertical(lipgloss.Center, question, okButton)

	return lipgloss.
		NewStyle().
		Width(m.TermSize.Width).
		Align(lipgloss.Center).
		Render(dialogBoxStyle.Render(ui))
}

// ** Table related functions ** //
// Create Columns.
func createViewTableColumns(maxWidth int) []table.Column {
	baseWidth := int(maxWidth / 20)

	columns := []table.Column{
		{Title: "ID", Width: baseWidth},
		{Title: "Status", Width: baseWidth},
		{Title: "Title", Width: baseWidth * 12},
		{Title: "Domain", Width: baseWidth * 4},
		{Title: "Created", Width: baseWidth * 2},
	}

	return columns
}

// Create rows
func getTableRows(items []wallabago.Item, filters walgotTableFilters) []table.Row {
	r := []table.Row{}

	for i := 0; i < len(items); i++ {
		title := items[i].Title
		id := strconv.Itoa(items[i].ID)
		domainName := items[i].DomainName
		status := "  "
		createdAt := items[i].CreatedAt.Time.Format("2006-02-01")

		if filters.Unread && items[i].IsArchived != 0 {
			continue
		}
		if filters.Starred && items[i].IsStarred != 1 {
			continue
		}
		if filters.Archived && items[i].IsArchived != 1 {
			continue
		}

		archivedEntry := true
		if items[i].IsArchived == 0 {
			status = "🆕"
			archivedEntry = false
		}
		if items[i].IsStarred == 1 {
			status += "⭐"

		}

		if archivedEntry {
			// This create a bug in the selected row,
			// where it stops the selected style (blue background).
			// TODO: Create an issue on bubble bugtracker
			title = lipgloss.NewStyle().Faint(true).Render(title)
		}

		r = append(r, table.Row{
			id,
			status,
			title,
			domainName,
			createdAt,
		})

	}

	return r
}

// Generate the bubbletea table.
func createViewTable(maxWidth int, maxHeight int) table.Model {
	t := table.New(
		table.WithColumns(createViewTableColumns(maxWidth)),
		table.WithHeight(maxHeight),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		BorderTop(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57"))

	t.SetStyles(s)

	return t
}

// ** Viewport related functions ** //
// Generate content for article detail viewport.
func getDetailViewportContent(selectedID int, entries []wallabago.Item) string {
	content := "…"
	if index := getSelectedEntryIndex(entries, selectedID); index >= 0 {
		content = getSelectedEntryContent(entries, index)
	}

	return content
}

// Retrieve index of the selected entry in model.Entries
func getSelectedEntryIndex(entries []wallabago.Item, id int) int {
	entryIndex := -1
	for i := 0; i < len(entries); i++ {
		if entries[i].ID == id {
			return i
		}
	}

	return entryIndex
}

// Retrieve the article content, in clean and wrap text.
func getSelectedEntryContent(entries []wallabago.Item, index int) string {
	contentHTML := entries[index].Content
	content := html2text.HTML2Text(contentHTML)
	return wordwrap.String(content, 72)
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
func sendEntryUpdate(msg string, sID int, m *model) (int, int, string) {
	entry := m.Entries[getSelectedEntryIndex(m.Entries, sID)]
	action := "Toggled entry status: "
	a := entry.IsArchived
	s := entry.IsStarred
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
	}

	return a, s, action
}
