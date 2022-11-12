package tui

import (
	"encoding/json"
	"testing"

	"github.com/Strubbl/wallabago/v7"
)

func TestGetRequiredNbAPICalls(t *testing.T) {
	var tests = []struct {
		inputNbArticles            int
		inputLimitArticleByAPICall int
		expectedNbCalls            int
	}{
		{0, 0, 0},
		{100, 10, 10},
		{500, 255, 2},
		{100, 100, 1},
		{-1, 100, 0},
		{100, -1, 1},
	}

	for _, test := range tests {
		result := getRequiredNbAPICalls(test.inputNbArticles, test.inputLimitArticleByAPICall)
		if test.expectedNbCalls != result {
			t.Errorf("getRequiredNbAPICalls(%v, %v): expectedNbCalls %v, got %v", test.inputNbArticles, test.inputLimitArticleByAPICall, test.expectedNbCalls, result)
		}
	}
}

func TestGetSelectedEntryIndex(t *testing.T) {
	entryOne := `{"is_archived":1,"is_starred":0,"user_name":"strubbl","user_email":"wallabag.test@wallabag.test","user_id":1,"tags":[{"id":4,"label":"3min","slug":"3min"}],"is_public":false,"id":12327,"uid":null,"title":"Formel 1, Racing Point zittert vor Finale: Viel zu verlieren","url":"https:\/\/www.motorsport-magazin.com\/formel1\/news-268468-formel-1-2020-abu-dhabi-qualifying-rennen-racing-point-zittert-vor-finale-viel-zu-verlieren-sergio-perez-gesamtwertung-lance-stroll-mclaren-mercedes\/","hashed_url":"537b9884cb6d90ec538eb79ef41beb5bcb506b27","origin_url":null,"given_url":"https:\/\/www.motorsport-magazin.com\/formel1\/news-268468-formel-1-2020-abu-dhabi-qualifying-rennen-racing-point-zittert-vor-finale-viel-zu-verlieren-sergio-perez-gesamtwertung-lance-stroll-mclaren-mercedes\/","hashed_given_url":"537b9884cb6d90ec538eb79ef41beb5bcb506b27","archived_at":"2020-12-12T21:33:51+0100","content":"Er baut auf eine weitere starke Performance in der Anfangsphase des Rennens.","created_at":"2020-12-12T21:31:04+0100","updated_at":"2020-12-12T21:33:51+0100","published_at":null,"published_by":null,"starred_at":null,"annotations":[],"mimetype":"text\/html; charset=utf-8","language":"de","reading_time":3,"domain_name":"www.motorsport-magazin.com","preview_picture":"https:\/\/wallabag.test\/assets\/images\/7\/0\/70cc31ae\/a3e0a994.jpeg","http_status":"200","headers":{"server":"nginx\/1.10.3 (Ubuntu)","date":"Sat, 12 Dec 2020 20:31:03 GMT","content-type":"text\/html; charset=utf-8","content-length":"55474","connection":"keep-alive","set-cookie":"PHPSESSID=ed369ef2e34da3dcb696bdf0b94b60e4; path=\/; secure; HttpOnly","expires":"Thu, 19 Nov 1981 08:52:00 GMT","cache-control":"no-store, no-cache, must-revalidate","pragma":"no-cache","vary":"Accept-Encoding","x-storage":"mem","age":"0","x-cache":"MISS","x-cache-hits":"0","x-cache-grace":"none","x-cache-debug":"Main Site","access-control-allow-origin":"https:\/\/ads.motorsport-magazin.com","accept-ranges":"bytes"},"_links":{"self":{"href":"\/api\/entries\/12327"}}}`
	entryTwo := `{"is_archived":1,"is_starred":0,"user_name":"strubbl","user_email":"wallabag.test@wallabag.test","user_id":1,"tags":[{"id":4,"label":"3min","slug":"3min"}],"is_public":false,"id":12328,"uid":null,"title":"Formel 1, Racing Point zittert vor Finale: Viel zu verlieren","url":"https:\/\/www.motorsport-magazin.com\/formel1\/news-268468-formel-1-2020-abu-dhabi-qualifying-rennen-racing-point-zittert-vor-finale-viel-zu-verlieren-sergio-perez-gesamtwertung-lance-stroll-mclaren-mercedes\/","hashed_url":"537b9884cb6d90ec538eb79ef41beb5bcb506b27","origin_url":null,"given_url":"https:\/\/www.motorsport-magazin.com\/formel1\/news-268468-formel-1-2020-abu-dhabi-qualifying-rennen-racing-point-zittert-vor-finale-viel-zu-verlieren-sergio-perez-gesamtwertung-lance-stroll-mclaren-mercedes\/","hashed_given_url":"537b9884cb6d90ec538eb79ef41beb5bcb506b27","archived_at":"2020-12-12T21:33:51+0100","content":"Er baut auf eine weitere starke Performance in der Anfangsphase des Rennens.","created_at":"2020-12-12T21:31:04+0100","updated_at":"2020-12-12T21:33:51+0100","published_at":null,"published_by":null,"starred_at":null,"annotations":[],"mimetype":"text\/html; charset=utf-8","language":"de","reading_time":3,"domain_name":"www.motorsport-magazin.com","preview_picture":"https:\/\/wallabag.test\/assets\/images\/7\/0\/70cc31ae\/a3e0a994.jpeg","http_status":"200","headers":{"server":"nginx\/1.10.3 (Ubuntu)","date":"Sat, 12 Dec 2020 20:31:03 GMT","content-type":"text\/html; charset=utf-8","content-length":"55474","connection":"keep-alive","set-cookie":"PHPSESSID=ed369ef2e34da3dcb696bdf0b94b60e4; path=\/; secure; HttpOnly","expires":"Thu, 19 Nov 1981 08:52:00 GMT","cache-control":"no-store, no-cache, must-revalidate","pragma":"no-cache","vary":"Accept-Encoding","x-storage":"mem","age":"0","x-cache":"MISS","x-cache-hits":"0","x-cache-grace":"none","x-cache-debug":"Main Site","access-control-allow-origin":"https:\/\/ads.motorsport-magazin.com","accept-ranges":"bytes"},"_links":{"self":{"href":"\/api\/entries\/12327"}}}`
	entryThree := `{"is_archived":1,"is_starred":0,"user_name":"strubbl","user_email":"wallabag.test@wallabag.test","user_id":1,"tags":[{"id":4,"label":"3min","slug":"3min"}],"is_public":false,"id":12329,"uid":null,"title":"Formel 1, Racing Point zittert vor Finale: Viel zu verlieren","url":"https:\/\/www.motorsport-magazin.com\/formel1\/news-268468-formel-1-2020-abu-dhabi-qualifying-rennen-racing-point-zittert-vor-finale-viel-zu-verlieren-sergio-perez-gesamtwertung-lance-stroll-mclaren-mercedes\/","hashed_url":"537b9884cb6d90ec538eb79ef41beb5bcb506b27","origin_url":null,"given_url":"https:\/\/www.motorsport-magazin.com\/formel1\/news-268468-formel-1-2020-abu-dhabi-qualifying-rennen-racing-point-zittert-vor-finale-viel-zu-verlieren-sergio-perez-gesamtwertung-lance-stroll-mclaren-mercedes\/","hashed_given_url":"537b9884cb6d90ec538eb79ef41beb5bcb506b27","archived_at":"2020-12-12T21:33:51+0100","content":"Er baut auf eine weitere starke Performance in der Anfangsphase des Rennens.","created_at":"2020-12-12T21:31:04+0100","updated_at":"2020-12-12T21:33:51+0100","published_at":null,"published_by":null,"starred_at":null,"annotations":[],"mimetype":"text\/html; charset=utf-8","language":"de","reading_time":3,"domain_name":"www.motorsport-magazin.com","preview_picture":"https:\/\/wallabag.test\/assets\/images\/7\/0\/70cc31ae\/a3e0a994.jpeg","http_status":"200","headers":{"server":"nginx\/1.10.3 (Ubuntu)","date":"Sat, 12 Dec 2020 20:31:03 GMT","content-type":"text\/html; charset=utf-8","content-length":"55474","connection":"keep-alive","set-cookie":"PHPSESSID=ed369ef2e34da3dcb696bdf0b94b60e4; path=\/; secure; HttpOnly","expires":"Thu, 19 Nov 1981 08:52:00 GMT","cache-control":"no-store, no-cache, must-revalidate","pragma":"no-cache","vary":"Accept-Encoding","x-storage":"mem","age":"0","x-cache":"MISS","x-cache-hits":"0","x-cache-grace":"none","x-cache-debug":"Main Site","access-control-allow-origin":"https:\/\/ads.motorsport-magazin.com","accept-ranges":"bytes"},"_links":{"self":{"href":"\/api\/entries\/12327"}}}`
	var item1 wallabago.Item
	json.Unmarshal([]byte(entryOne), &item1)
	var item2 wallabago.Item
	json.Unmarshal([]byte(entryTwo), &item2)
	var item3 wallabago.Item
	json.Unmarshal([]byte(entryThree), &item3)

	var items = []wallabago.Item{item1, item2, item3}
	var tests = []struct {
		inputItems    []wallabago.Item
		inputID       int
		expectedIndex int
	}{
		{items, 12327, 0},
		{items, 12328, 1},
		{items, 12329, 2},
		{items, 12325, -1},
		{items, 123, -1},
	}

	for _, test := range tests {
		result := getSelectedEntryIndex(test.inputItems, test.inputID)
		if test.expectedIndex != result {
			t.Errorf("getSelectedEntryIndex(%v, %v): expectedIndex %v, got %v", test.inputItems, test.inputID, test.expectedIndex, result)
		}
	}
}
