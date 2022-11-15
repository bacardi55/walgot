package tui

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Strubbl/wallabago/v7"
	"github.com/k3a/html2text"
	"github.com/muesli/reflow/wordwrap"
)

// Open link in default browser.
// TODO: test on macOS or windows…
func openLinkInBrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = errors.New("unsupported platform")
	}

	if err != nil {
		return err
	}
	return nil
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

// Calculate the number of API call needed to retrieve all articles.
func getRequiredNbAPICalls(nbArticles, limitArticleByAPICall int) int {
	if nbArticles <= 0 {
		return 0
	}
	if limitArticleByAPICall <= 0 {
		limitArticleByAPICall = nbArticles
	}

	nbCalls := 1
	if nbArticles > limitArticleByAPICall {
		nbCalls = nbArticles / limitArticleByAPICall
		if float64(nbCalls) < float64(nbArticles)/float64(limitArticleByAPICall) {
			nbCalls++
		}
	}
	return nbCalls
}

// Case insensitive strings.Contains:
func containsI(s, t string) bool {
	return strings.Contains(
		strings.ToLower(s),
		strings.ToLower(t),
	)
}
