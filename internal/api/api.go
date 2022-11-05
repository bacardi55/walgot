package api

import "github.com/Strubbl/wallabago/v7"

func InitWallabagoAPI(credentialsFile string) error {
	return wallabago.ReadConfig(credentialsFile)
}

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

func GetNbTotalEntries() (int, error) {
	return wallabago.GetNumberOfTotalArticles()
}
