package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	// Library for Wallabag API:
	"github.com/Strubbl/wallabago/v7"

	// Library for creating the TUI:
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	// Library for converting HTML to readable format:
	"jaytaylor.com/html2text"

	// Library for wraping word:
	"github.com/muesli/reflow/wordwrap"
)

// TODO: Read from text file to make it easier to update.
const currentVersion = "0.0.1"

// TODO:
// - Make it configurable
// - Allow "short path" like "~/"
const defaultConfigJSON = "/home/bacardi55/.config/walgot/config.json"

var (
	version    = flag.Bool("v", false, "print version")
	verbose    = flag.Bool("verbose", false, "verbose mode")
	debug      = flag.Bool("d", false, "get debug output (implies verbose mode)")
	configJSON = flag.String("config", defaultConfigJSON, "file name of config JSON file")

	// TODO: Move to configuration:
	itemsPerPage = 50
)

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
	Table       table.Model
	Viewport    viewport.Model
	Entries     []wallabago.Item
	Ready       bool
	Reloading   bool
	SelectedID  int
	TermSize    termSize
	CurrentView string
	Options     walgotTableOptions
}

// Response message from Wallabago API call
type wallabagoResponseMsg []wallabago.Item

// Selected row in table list Message
type walgotSelectRowMsg int

// Update list message
type walgotUpdateListMsg string

// Callback for requesting entries via API:
func requestWallabagEntries() tea.Msg {
	// Get total number of articles:
	nbArticles, e := wallabago.GetNumberOfTotalArticles()
	if e != nil {
		fmt.Println("Couldn't retrieve entries from wallabag")
		if *debug {
			log.Println("Wallabago error:", e.Error())
		}
		os.Exit(1)
	}
	if *verbose {
		log.Println("Found", nbArticles, "in wallabag")
	}

	// Let's not request thousands or article at one, 555 is already big…
	// TODO: Make this configurable.
	limitArticleByAPICall := 100
	nbCalls := 1
	if nbArticles > limitArticleByAPICall {
		nbCalls = nbArticles / limitArticleByAPICall
		if float64(nbCalls) < float64(nbArticles)/float64(limitArticleByAPICall) {
			nbCalls++
		}
	}
	if *debug {
		log.Println(nbCalls, "API call will be needed to wallabag API")
	}

	// TODO: Move this to async channel?
	var entries []wallabago.Item
	for i := 1; i < nbCalls+1; i++ {
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

		if err != nil {
			if *verbose {
				fmt.Println("Couldn't retrieve some entries from wallabag")
				if *debug {
					log.Println("API call number", i)
					log.Println("Wallabago error:", err.Error())
				}
			}
		}

		if *debug {
			log.Println("Entries, batch n°", i)
			log.Println("Adding", len(r.Embedded.Items), "entries")
		}
		entries = append(entries, r.Embedded.Items...)

		// TODO:
		break
	}

	if *verbose {
		log.Println("Entries have been correctly retrieved")
	}

	return wallabagoResponseMsg(entries)
}

// Callback for selecting entry in list:
func selectEntryCommand(selectedRowID int) tea.Cmd {
	return func() tea.Msg {
		return walgotSelectRowMsg(selectedRowID)
	}
}

func (m model) Init() tea.Cmd {
	return requestWallabagEntries
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	if msg, ok := msg.(tea.KeyMsg); ok {
		// C-c to kill the app.
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		} else if msg.String() == "?" {
			log.Println("Display help")
			// TODO
			return m, nil
		}
	}

	// This needs to happen before sending to the sub update function.
	if v, ok := msg.(walgotSelectRowMsg); ok {
		m.SelectedID = int(v)
	}

	// Now send to the right sub-update function:
	if m.SelectedID > 0 {
		return updateEntryView(msg, m)
	}
	return updateListView(msg, m)
}

// Manage update for the detail entry view.
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
		case "S":
			log.Println("Star article")
			// TODO
		case "A":
			log.Println("Archived entry")
			// TODO
		}
	}

	m.Viewport, cmd = m.Viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// Manage updates for the list view.
func updateListView(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if *debug {
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
			m.Reloading = true
			return m, requestWallabagEntries
		// Filters for the table list:
		case "u", "s", "a":
			if msg.String() == "u" {
				m.Options.Filters.Unread = !m.Options.Filters.Unread
				// Unread and Archived can't be selected at the same time:
				if m.Options.Filters.Unread {
					m.Options.Filters.Archived = false
				}
			}
			if msg.String() == "s" {
				m.Options.Filters.Starred = !m.Options.Filters.Starred
			}
			if msg.String() == "a" {
				m.Options.Filters.Archived = !m.Options.Filters.Archived
				// Unread and Archived can't be selected at the same time:
				if m.Options.Filters.Archived {
					m.Options.Filters.Unread = false
				}
			}
			m.Table.SetRows(getTableRows(m.Entries, m.Options.Filters))
		}

	// When resizing the window, sizes needs to change everywhere…
	case tea.WindowSizeMsg:
		m.TermSize = termSize{msg.Width, msg.Height}
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

		// Response from the API, data has changed:
	case wallabagoResponseMsg:
		// Response received, we are not reloading anymore:
		m.Reloading = false
		m.Entries = msg
		log.Println("wallabagoResponseMsg", len(msg))
		m.Table.SetRows(getTableRows(m.Entries, m.Options.Filters))
	}

	return m, cmd
}

// Return the view to display.
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
		subtitle = " - All articles"
	}

	t := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Bold(true).Render("Walgot"),
		//lipgloss.NewStyle().Render(" - a WALlabag GO Tui client"),
		lipgloss.NewStyle().Render(subtitle),
	)
	return titleStyle.Render(t)
}

// Return the footer part of the view.
func (m model) footerView() string {
	return lipgloss.
		NewStyle().
		Width(m.TermSize.Width).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		Align(lipgloss.Center).
		PaddingTop(2).
		Render("[r]eload -- Toggles: [u]nread, [s]tarred, [a]rchived -- [h]elp (todo)")
}

// Return the main part of the view.
func (m model) mainView() string {
	if !m.Ready {
		// Not initialized yet, let's not style it.
		return "\n   Initializing…"
	}
	if m.Reloading {
		return lipgloss.NewStyle().
			Width(m.TermSize.Width).
			Align(lipgloss.Center).
			Render("Loading entries from wallabag")
	}

	// Return detail view:
	if m.CurrentView == "detail" {
		return lipgloss.
			NewStyle().
			Width(m.TermSize.Width).
			Align(lipgloss.Center).
			Render(entryDetailView(m))
	}

	// Return list view:
	return listView(m)
}

// Get article detail view.
func entryDetailView(m model) string {
	return m.Viewport.View()
}

// Get list view.
func listView(m model) string {
	return m.Table.View()
}

// Manage debug flags.
func handleFlags() {
	flag.Parse()
	if *debug && len(flag.Args()) > 0 {
		log.Printf("handleFlags: non-flag args=%v", strings.Join(flag.Args(), " "))
	}
	// version first, because it directly exits here
	if *version {
		fmt.Printf("version %v\n", currentVersion)
		os.Exit(0)
	}
	// test verbose before debug because debug implies verbose
	if *verbose && !*debug {
		log.Printf("verbose mode")
	}
	if *debug {
		log.Printf("handleFlags: debug mode")
		// debug implies verbose
		*verbose = true
	}
}

// Manage log configuration.
func configLogs() error {
	// TODO: make configurable.
	file, err := os.OpenFile("/tmp/walgot.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Couldn't open log file, exiting")
		return err
	}

	log.SetOutput(file)
	return nil
}

// Main function.
func main() {
	handleFlags()

	if err := configLogs(); err != nil {
		log.Println("Couldn't configure logs")
		if *debug {
			log.Println(err)
		}
		os.Exit(1)
	}

	// check for config
	if *verbose {
		log.Println("reading config", *configJSON)
	}

	err := wallabago.ReadConfig(*configJSON)
	if err != nil {
		fmt.Println("Error reading config")
		if *debug {
			fmt.Println("Wallabago error while loading config", err.Error())
		}
		os.Exit(1)
	}

	m := model{
		SelectedID:  0,
		Ready:       false,
		Reloading:   true,
		CurrentView: "list",
		// Default start is unread only:
		// TODO: make this configurable.
		Options: walgotTableOptions{
			Filters: walgotTableFilters{
				Unread: true,
			},
		},
	}

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// Retrieve index of the selected entry in model.Entries
func getSelectedEntryIndex(entries []wallabago.Item, id int) int {
	if *debug {
		log.Println("Selected entry:", id)
	}

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
		return "", errors.New("Error retrieving article content")
	}
	return wordwrap.String(content, 72), nil
}

// Table related functions:
// Create Columns.
func createViewTableColumns(maxWidth int) []table.Column {
	baseWidth := int(math.Floor(float64(maxWidth / 20)))

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

// Generate content for article detail viewport.
func getDetailViewportContent(selectedID int, entries []wallabago.Item) string {
	articleTitle := "Title loading…"
	content := "Content loading…"
	if index := getSelectedEntryIndex(entries, selectedID); index >= 0 {
		var err error
		content, err = getSelectedEntryContent(entries, index)
		articleTitle = entries[index].Title
		if err != nil {
			if *debug {
				log.Println("Error retrieving article content")
				log.Println(err)
			}
			content = "Error retrieving article content"
		}

	}
	return lipgloss.
		NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Bold(true).
		Render(articleTitle) +
		"\n\n" +
		content
}
