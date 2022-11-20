package tui

import (
	"errors"
	"net/url"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/Strubbl/wallabago/v7"
	"github.com/atotto/clipboard"
	"github.com/k3a/html2text"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
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

// Copy link.
// TODO: test on macOS or windows…
func copyLinkToClipboard(url string) error {
	return clipboard.WriteAll(url)
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
func getSelectedEntryContent(entries []wallabago.Item, index, maxWidth int) string {
	content := getContentForViewport(entries[index].Content)

	w := 72
	if maxWidth < w {
		w = maxWidth - 2
	}

	return wrap.String(wordwrap.String(content, w), w)
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

// Validate is the given string is a URL format.
func isValidURL(u string) bool {
	if _, err := url.ParseRequestURI(u); err != nil {
		return false
	}
	return true
}

func getContentForViewport(contentHTML string) string {
	content, links := getCleanedContentAndLinks(contentHTML)
	content += "\r\n\r\n\r\n" + generateFootnoteLinks(links)

	return content
}

func getCleanedContentAndLinks(contentHTML string) (string, []string) {
	content := html2text.HTML2TextWithOptions(contentHTML, html2text.WithLinksInnerText())
	return parseLinksInContent(content)
}

// parse links in the recieved string.
func parseLinksInContent(content string) (string, []string) {
	var links []string

	re := regexp.MustCompile("(?i)<((https?|gopher|gemini)://[^>]*)>")
	// Find and loop over all matching strings.
	results := re.FindAllStringSubmatch(content, -1)
	for i := range results {
		content = strings.Replace(content, results[i][0], "["+strconv.Itoa(i+1)+"]", 1)
		links = append(links, results[i][1])
	}

	return content, links
}

// Generate footnote text for links in article:
func generateFootnoteLinks(links []string) string {
	footnotes := "Links:\r\n\r\n"
	for i, l := range links {
		footnotes += "[" + strconv.Itoa(i+1) + "]: " + l + "\r\n"
	}

	return footnotes
}
