package api

import (
	"encoding/json"
	"strconv"

	"github.com/Strubbl/wallabago/v7"
)

// InitWallabagoAPI set wallabago config.
func InitWallabagoAPI(credentialsFile string) error {
	return wallabago.ReadConfig(credentialsFile)
}

// GetEntries returns entries from wallabag APIs.
func GetEntries(itemsPerPage int, pageNumber int) (wallabago.Entries, error) {
	return wallabago.GetEntries(
		wallabago.APICall,
		-1,
		-1,
		"updated",
		"desc",
		pageNumber,
		itemsPerPage,
		"",
	)
}

// GetNbTotalEntries returns the total number of entries saved in wallabag.
func GetNbTotalEntries() (int, error) {
	return wallabago.GetNumberOfTotalArticles()
}

// UpdateEntry update an article on wallabag.
func UpdateEntry(entryID, archive, starred int) ([]byte, error) {
	tmp := map[string]string{
		"archive": strconv.Itoa(archive),
		"starred": strconv.Itoa(starred),
	}
	body, _ := json.Marshal(tmp)
	url := wallabago.Config.WallabagURL + "/api/entries/" + strconv.Itoa(entryID) + ".json"
	// Send request and return result:
	return wallabago.APICall(
		url,
		"PATCH",
		body,
	)
}
