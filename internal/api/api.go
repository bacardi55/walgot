package api

import "github.com/Strubbl/wallabago/v7"

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
