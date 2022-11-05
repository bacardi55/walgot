package tui

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"git.bacardi55.io/bacardi55/walgot/internal/api"
	"git.bacardi55.io/bacardi55/walgot/internal/config"

	"github.com/Strubbl/wallabago/v7"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"jaytaylor.com/html2text"
)

// ** Model related Struct ** //

// Terminal physical size:
type termSize struct {
	Width  int
	Height int
}

type walgotTableFilters struct {
	Archived bool
	Starred  bool
	Unread   bool
}

/*
type walgotTableSorts struct {
}
*/

type walgotTableOptions struct {
	Filters walgotTableFilters
	//Sorts walgotTableSorts
}

// Model structure
type model struct {
	Table                table.Model
	Viewport             viewport.Model
	Entries              []wallabago.Item
	Ready                bool
	Reloading            bool
	SelectedID           int
	TermSize             termSize
	CurrentView          string
	Options              walgotTableOptions
	Spinner              spinner.Model
	TotalEntriesOnServer int
	DebugMode            bool
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
		DebugMode:            config.DebugMode,
		// Default start is unread only:
		// TODO: make this configurable.
		Options: walgotTableOptions{
			Filters: walgotTableFilters{
				Unread:  config.DefaultListViewUnread,
				Starred: config.DefaultListViewStarred,
			},
		},
	}
}

// Response message for number of entities from Wallabago
type wallabagoResponseNbEntitiesMsg int

// Response message for all entities from Wallabago
type wallabagoResponseEntitiesMsg []wallabago.Item

// Selected row in table list Message
type walgotSelectRowMsg int

// Callback for requesting the total number of entries via API.
func requestWallabagNbEntries() tea.Msg {
	// Get total number of articles:
	//nbArticles, e := wallabago.GetNumberOfTotalArticles()
	nbArticles, e := api.GetNbTotalEntries()

	// TODO: move error to a walgotErrorMsg:
	if e != nil {
		fmt.Println("Couldn't retrieve entries from wallabag")
		log.Println("Wallabago error:", e.Error())
		os.Exit(1)
	}

	return wallabagoResponseNbEntitiesMsg(nbArticles)
}

// Callback for requesting entries via API.
func requestWallabagEntries(nbArticles int) tea.Cmd {
	// TODO: Make this configurable.
	articleByAPICall := 55

	return func() tea.Msg {
		// Let's not request thousands or article at one, 555 is already big…
		limitArticleByAPICall := articleByAPICall
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
			r, err := api.GetEntries(limitArticleByAPICall, i)
			/*
				r, err := wallabago.GetEntries(
					wallabago.APICall,
					-1,
					-1,
					"updated",
					"desc",
					i,
					limitArticleByAPICall,
					"",
				)
			*/

			// TODO: move to walgotErrorMsg
			if err != nil {
				fmt.Println("Couldn't retrieve some entries from wallabag")
				log.Println("API call number", i)
				log.Println("Wallabago error:", err.Error())
			}

			entries = append(entries, r.Embedded.Items...)
		}

		return wallabagoResponseEntitiesMsg(entries)
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
		log.Println(fmt.Sprintf("Update message received, type: %T", msg))
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		// C-c to kill the app.
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		} else if msg.String() == "?" {
			log.Println("Display help")
			m.CurrentView = "help"
			return m, nil
		}
	}

	// This needs to happen before sending to the sub update function.
	if v, ok := msg.(walgotSelectRowMsg); ok {
		m.SelectedID = int(v)
	}

	if m.CurrentView == "help" {
		return updateHelpView(msg, m)
	}

	// Now send to the right sub-update function:
	if m.SelectedID > 0 {
		return updateEntryView(msg, m)
	}
	return updateListView(msg, m)
}

// View method.
func (m model) View() string {
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.mainView(), m.footerView())
}

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

	if !m.Reloading {
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
		PaddingTop(2).
		Render(text)
}

// Return the main part of the view.
func (m model) mainView() string {
	if !m.Ready {
		// Not initialized yet, let's not style it.
		return "\n   Initializing…"
	}
	if m.Reloading {
		// TODO for MVP: Move to dedicated functions
		text := "Loading all"
		if m.TotalEntriesOnServer > 0 {
			text += " " + strconv.Itoa(m.TotalEntriesOnServer)
		}
		text += " entries from wallabag…"
		return lipgloss.NewStyle().
			Width(m.TermSize.Width).
			Align(lipgloss.Center).
			Render(m.Spinner.View() + text)
	}

	if m.CurrentView == "help" {
		// Return help view:
		return helpView(m)

	} else if m.SelectedID > 0 {
		// Return detail view:
		return entryDetailView(m)
	}
	// Return list view:
	return listView(m)
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

// Manage update for the detail entry view.
func updateEntryView(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// A row has been selected, display article detail:
	case walgotSelectRowMsg:
		m.CurrentView = "detail"

		if c, err := getDetailViewportContent(m.SelectedID, m.Entries); err == nil {
			m.Viewport.SetContent(c)
		} else {
			log.Println("Error retrieving content for entry", m.SelectedID)
		}

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
		case "S":
			log.Println("Star article")
			// TODO for MVP: Star article.
		case "A":
			log.Println("Archived entry")
			// TODO for MVP: Archive article.
		}
	}

	m.Viewport, cmd = m.Viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// Manage updates for the list view.
func updateListView(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	// TODO for MVP: Refactor this function.
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.DebugMode {
				log.Println("Selected row:", m.Table.SelectedRow())
			}
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
			log.Println("Loading entries from API")
			// Status as reloading:
			m.Reloading = true
			// Reset number of entries:
			m.TotalEntriesOnServer = 0
			return m, requestWallabagNbEntries
		// Filters for the table list:
		case "u", "s", "a":
			listViewFiltersUpdate(msg.String(), &m)
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
			requestWallabagEntries(m.TotalEntriesOnServer),
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

// ** View related functions. ** //
// Help view.
func helpView(m model) string {
	text := []byte(`Help:
  Keybinds
	On all screens:
	- ctrl+c: quit
	- h: help (this page)


	On listing page:
	- r: reload article from wallabag via APIs, takes time depending on the number of articles saved
	- u: toggle display only unread articles (disable archived filter)
	- s: toggle display only starred articles
	- a: toggle archived only articles (disable unread filter)
	- h: display help
	- ↑ or k / ↓ or j: move up / down one item in the list
	- page down / page up: move up / down 10 items in the list
	- home: go to the top of the list
	- end: go to bottom of the list
	- enter: select entry to read content
	- q: quit

	On detail page:
	- q: return to list
	- ↑ or k / ↓ or j: go up / down

	On help page:
	- q: return to list
`)

	return lipgloss.
		NewStyle().
		Width(m.TermSize.Width).
		Align(lipgloss.Left).
		Render(string(text))
}

// Get article detail view.
func entryDetailView(m model) string {
	return lipgloss.
		NewStyle().
		Width(m.TermSize.Width).
		Align(lipgloss.Center).
		Render(m.Viewport.View())
}

// Get list view.
func listView(m model) string {
	return m.Table.View()
}

// ** Table related functions ** //
// Create Columns.
func createViewTableColumns(maxWidth int) []table.Column {
	baseWidth := int(maxWidth / 20)

	columns := []table.Column{
		{Title: "ID", Width: baseWidth * 2},
		{Title: "Title", Width: baseWidth * 10},
		{Title: "Domain", Width: baseWidth * 4},
		{Title: "⭐", Width: baseWidth},
		{Title: "✓", Width: baseWidth},
		{Title: "Updated date", Width: baseWidth * 2},
	}

	return columns
}

// Create rows
func getTableRows(items []wallabago.Item, filters walgotTableFilters) []table.Row {
	r := []table.Row{}

	for i := 0; i < len(items); i++ {
		title := items[i].Title

		if filters.Unread && items[i].IsArchived != 0 {
			continue
		}
		if filters.Starred && items[i].IsStarred != 1 {
			continue
		}
		if filters.Archived && items[i].IsArchived != 1 {
			continue
		}

		s := " "
		if items[i].IsStarred == 1 {
			s = "⭐"
		}

		a := " "
		if items[i].IsArchived == 1 {
			a = "✓"
		} else {
			title = lipgloss.NewStyle().Bold(true).Render(items[i].Title)
		}

		r = append(r, table.Row{
			strconv.Itoa(items[i].ID),
			title,
			items[i].DomainName,
			s,
			a,
			items[i].UpdatedAt.Time.Format("2006-02-01"),
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
		Background(lipgloss.Color("57")).
		Bold(false)

	t.SetStyles(s)

	return t
}

// ** Viewport related functions ** //
// Generate content for article detail viewport.
func getDetailViewportContent(selectedID int, entries []wallabago.Item) (string, error) {
	articleTitle := "Title loading…"
	content := "Content loading…"
	if index := getSelectedEntryIndex(entries, selectedID); index >= 0 {
		var err error
		content, err = getSelectedEntryContent(entries, index)
		if err != nil {
			return "", errors.New("error finding selected entry")
		}
		articleTitle = entries[index].Title
	}
	text := lipgloss.
		NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Bold(true).
		Render(articleTitle) +
		"\n\n" +
		content

	return text, nil
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
func getSelectedEntryContent(entries []wallabago.Item, index int) (string, error) {
	contentHTML := entries[index].Content
	content, err := html2text.FromString(contentHTML, html2text.Options{PrettyTables: true})
	if err != nil {
		return "", errors.New("error retrieving article content")
	}
	return wordwrap.String(content, 72), nil
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

// Manage window size changes
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
