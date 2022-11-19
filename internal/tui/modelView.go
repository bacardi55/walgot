package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Strubbl/wallabago/v7"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

// Return the header part of the view.
func (m model) headerView() string {
	titleStyle := lipgloss.
		NewStyle().
		Width(m.TermSize.Width).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Align(lipgloss.Center)

	subtitle := ""
	if !m.Ready {
		subtitle += " - Loading…"
	} else if m.Reloading {
		subtitle += " - Reloading"
	} else if m.SelectedID > 0 {
		subtitle += " - Reading"
	} else {
		if m.Options.Filters.Search != "" {
			subtitle += " - Searching for " + m.Options.Filters.Search
		}
		if m.Options.Filters.Unread {
			subtitle += " - Unread"
		}
		if m.Options.Filters.Starred {
			subtitle += " - Starred"
		}
		if m.Options.Filters.Archived {
			subtitle += " - Archived"
		}
		if m.Options.Filters.Public {
			subtitle += " - Public"
		}
		if len(subtitle) == 0 && !m.Reloading {
			subtitle = " - All"
		}
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
	if m.Dialog.Message != "" {
		return dialogView(&m)
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
		m.Table.SetRows(getTableRows(m.Entries, m.Options.Filters, m.TermSize.Width))
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
  - p: Toggle public only articles (articles with a public link)
  - A: Toggle Archive / Unread for the current article (and update wallabag backend)
  - S: Toggle Starred / Unstarred for the current article (and update wallabag backend)
  - P: Toggle Public status - Public means article can be shared with a public link
  - O: Open article public link url in default browser. If article isn't public, it will open the original article link.
  - Y: Yank (copy) URL to clipboard. If article isn't public, it will open the original article link.
  - /: Open search box
  - N: Add a new url to wallabag.
  - D: Delete the selected entry.
  - esc: Clean search filter, if any
  - h: Display help
  - ↑ or k / ↓ or j: Move up / down one item in the list
  - page down / page up: Move up / down 10 items in the list
  - home: Go to the top of the list
  - end: Go to bottom of the list
  - enter: Select entry to read content
  - q: Remove search filter if any, otherwise quit

  On detail page:
  - A: Toggle Archive / Unread for the current article (and update wallabag backend)
  - S: Toggle Starred / Unstarred for the current article (and update wallabag backend)
  - P: Toggle Public status - Public means article can be shared with a public link
  - O: Open article public link url in default browser. If article isn't public, it will open the original article link.
  - Y: Yank (copy) URL to clipboard. If article isn't public, it will open the original article link.
  - D: Delete the selected entry.
  - q: Return to list
  - ↑ or k / ↓ or j: Go up / down

  On any dialog (modal) view:
  - "esc": Close the dialog

  On search modal view:
  - "enter": start search


  On help page:
  - q, esc: Return to list

  Status explanation:
  - ⭐ Starred article
  - 🆕 Unread article
  - 🔗 Article with a public shareable link
`)

	return lipgloss.
		NewStyle().
		Width(m.TermSize.Width).
		Align(lipgloss.Left).
		Render(string(text))
}

// Get article detail view.
func entryDetailView(m model) string {
	i := getSelectedEntryIndex(m.Entries, m.SelectedID)
	header := entryDetailViewTitle(&m.Entries[i])
	footer := entryDetailViewFooter(m.Viewport, &m.Entries[i])

	return lipgloss.
		NewStyle().
		Width(m.TermSize.Width).
		Align(lipgloss.Center).
		Render(header + "\n" + m.Viewport.View() + "\n" + footer)
}

// Retrieve title for detail view.
func entryDetailViewTitle(entry *wallabago.Item) string {
	title := lipgloss.
		NewStyle().
		Bold(true).
		Width(80).
		Align(lipgloss.Center).
		Render(wordwrap.String(entry.Title, 72))

	return lipgloss.
		NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Render(title)
}

// Retrieve footer for detail view.
func entryDetailViewFooter(viewport viewport.Model, entry *wallabago.Item) string {
	status := ""
	if entry.IsArchived == 0 {
		status += "🆕"
	}
	if entry.IsStarred == 1 {
		status += "⭐"
	}
	if entry.IsPublic {
		status += "🔗"
	}

	statusInfo := lipgloss.
		NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderRight(true).
		Render(fmt.Sprintf("%s", status))

	readInfo := lipgloss.
		NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderRight(true).
		Render(fmt.Sprintf("%3.f%%", viewport.ScrollPercent()*100))

	width := viewport.Width - lipgloss.Width(readInfo) - lipgloss.Width(statusInfo)
	if width < 0 {
		width = 0
	}
	line := strings.Repeat("─", width)

	return lipgloss.JoinHorizontal(lipgloss.Center, statusInfo, line, readInfo)
}

// Get list view.
func listView(m model) string {
	return m.Table.View()
}

// Get dialog view.
func dialogView(m *model) string {
	dialogBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 0).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)

	actionButton := ""
	if m.Dialog.Action == "search" || m.Dialog.Action == "add" {
		text := strings.Title(m.Dialog.Action) + " (Enter)"
		actionButton = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#888B7E")).
			Padding(0, 3).
			MarginTop(1).
			Underline(true).
			Render(text)
	}

	closeButton := lipgloss.NewStyle().
		Background(lipgloss.Color("#FFF7DB")).
		Foreground(lipgloss.Color("#888B7E")).
		Padding(0, 3).
		MarginTop(1).
		Underline(true).
		Render("Close (esc)")

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		actionButton,
		closeButton,
	)

	content := lipgloss.
		NewStyle().
		Width(50).
		Align(lipgloss.Left).
		Render(m.Dialog.Message)

	if m.Dialog.ShowInput {
		m.Dialog.TextInput.PromptStyle = lipgloss.
			NewStyle().
			Foreground(lipgloss.Color("205")).
			Align(lipgloss.Left)
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			m.Dialog.TextInput.View(),
		)
	}

	ui := lipgloss.JoinVertical(lipgloss.Center, content, buttons)

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
	var columns []table.Column

	if maxWidth > 130 {
		columns = []table.Column{
			{Title: "ID", Width: baseWidth},
			{Title: "Status", Width: baseWidth},
			{Title: "Title", Width: baseWidth * 12},
			{Title: "Domain", Width: baseWidth * 4},
			{Title: "Created", Width: baseWidth * 2},
		}
	} else if maxWidth > 80 {
		columns = []table.Column{
			{Title: "ID", Width: baseWidth},
			{Title: "Status", Width: baseWidth},
			{Title: "Title", Width: baseWidth * 18},
		}
	} else {
		columns = []table.Column{
			{Title: "ID", Width: 0},
			{Title: "Title", Width: baseWidth * 20},
		}

	}

	return columns
}

// Create rows
// TODO: create test for this function.
func getTableRows(items []wallabago.Item, filters walgotTableFilters, maxWidth int) []table.Row {
	r := []table.Row{}

	for i := 0; i < len(items); i++ {
		title := items[i].Title
		id := strconv.Itoa(items[i].ID)
		domainName := items[i].DomainName
		status := "  "
		createdAt := items[i].CreatedAt.Time.Format("2006-02-01")

		// Public filter:
		if filters.Public && !items[i].IsPublic {
			continue
		}
		// Unread filter:
		if filters.Unread && items[i].IsArchived != 0 {
			continue
		}
		// Archived filter:
		if filters.Archived && items[i].IsArchived != 1 {
			continue
		}
		// Starred filter:
		if filters.Starred && items[i].IsStarred != 1 {
			continue
		}
		// Search filter:
		if filters.Search != "" && !containsI(items[i].Title, filters.Search) {
			continue
		}

		archivedEntry := true
		if items[i].IsArchived == 0 {
			status = "🆕"
			archivedEntry = false
		}
		if items[i].IsStarred == 1 {
			status += "⭐"
		} else {
			status += "  "
		}
		if items[i].IsPublic {
			status += "🔗"
		}

		if archivedEntry {
			// This create a bug in the selected row,
			// where it stops the selected style (blue background).
			// TODO: Create an issue on bubble bugtracker
			title = lipgloss.NewStyle().Faint(true).Render(title)
		}

		var new table.Row
		if maxWidth > 130 {
			new = table.Row{
				id,
				status,
				title,
				domainName,
				createdAt,
			}
		} else if maxWidth > 80 {
			new = table.Row{
				id,
				status,
				title,
			}
		} else {
			new = table.Row{
				id,
				title,
			}
		}

		r = append(r, new)
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
