package api

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/Strubbl/wallabago/v7"
)

// InitWallabagoAPI set wallabago config.
func InitWallabagoAPI(credentialsFile string) error {
	return wallabago.ReadConfig(credentialsFile)
}

// GetEntries returns entries from wallabag APIs.
func GetEntries(itemsPerPage, pageNumber int, sortField, sortOrder string) (wallabago.Entries, error) {
	return wallabago.GetEntries(
		wallabago.APICall,
		-1,
		-1,
		sortField,
		sortOrder,
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
func UpdateEntry(entryID, archive, starred, public int) ([]byte, error) {
	tmp := map[string]string{
		"archive": strconv.Itoa(archive),
		"starred": strconv.Itoa(starred),
		"public":  strconv.Itoa(public),
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

// AddEntry add an entry on wallabag.
func AddEntry(url string) (wallabago.Item, error) {
	postData := map[string]string{
		"url": url,
	}
	postDataJSON, err := json.Marshal(postData)
	if err != nil {
		return wallabago.Item{}, err
	}
	entriesURL := wallabago.Config.WallabagURL + "/api/entries.json"
	body, err := wallabago.APICall(entriesURL, "POST", postDataJSON)
	if err != nil {
		return wallabago.Item{}, err
	}

	var item wallabago.Item
	err = json.Unmarshal(body, &item)
	if err != nil {
		return wallabago.Item{}, err
	}

	return item, nil
}

// DeleteEntry removes an entry from wallabag.
func DeleteEntry(id int) error {
	url := wallabago.Config.WallabagURL +
		"/api/entries/" +
		strconv.Itoa(id)

	response, err := wallabago.APICall(
		url,
		"DELETE",
		[]byte{},
	)
	if err != nil {
		return errors.New("Couldn't delete entry:" + strconv.Itoa(id))
	}

	return nil
}
