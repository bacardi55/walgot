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