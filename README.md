# Walgot - a WALlabag GO Tui client

**This app is still very early stage and work in progress, so be careful while using it!**

[![builds.sr.ht status](https://builds.sr.ht/~bacardi55/walgot.svg)](https://builds.sr.ht/~bacardi55/walgot?)
[![license: AGPL-3.0-only](https://img.shields.io/badge/license-AGPL--3.0--only-informational.svg)](LICENSE)

Official repository and project is on [sourcehut](https://git.sr.ht/~bacardi55/walgot). Github and codeberg are only mirrors. This is where [binaries are uploaded](https://git.sr.ht/~bacardi55/walgot/refs) (Click on the version to see the files in the release detail page).

## What is walgot?

Walgot is a TUI [wallabag](https://wallabag.org) client. Wallabag is an opensource "read it later" application that can be selfhosted. This application aims to be an easy and interactive TUI client for it.

Walgot is built in [golang](golang.org/), leveraging the great [bubbletea](https://github.com/charmbracelet/bubbletea) library as well as [Wallabago](https://github.com/Strubbl/wallabago), a library to simplify wallabag API calls.

## Online only at this stage

**Important note**: The way walgot works is by downloading **all** articles from wallabag API at the start of the session (or when using the refresh keybind). Then walgot will allow filtering, viewing or updates (read, star) of articles and push changes via API. There is no local cache or database, so no offline usage (at least for now).

Once the initial load is done, internet access is not needed anymore to read articles content. It will be needed for changing article status. An offline mode might be added later, but there are more urgent features to build first :).

## installation

### dependencies

Requires `go ≥ 1.17` and `make`.

### binary

Download one of the binary available on the binary file in the release page. 
Make sure it is executable (with `chmod +x binaryFile`) and then run walgot with the command:

### Manual
Use the makefile provided and run `make build`, it will create a binary file in a `bin` directory.

``` bash
git clone https://git.sr.ht/~bacardi55/walgot
cd walgot
make dependencies
make build
```


## configuration

Walgot is configure with 2 different files: `walgot.json` and `credentials.json`. An example of these two files can be found in the `example` directory in the git repository.

To use the default configuration file:

``` bash
mkdir ~/.config/walgot/
cp example/*.json ~/.config/walgot/
# NOTA: Here you need to edit at least ~/.config/walgot/credentials.json
# Then start walgot:
./bin/walgot
```

Don't forget to at least edit the credentials.json file, or it won't work.

### walgot.json

The main configuration file, in json format. The default value are as below:

``` json
{
    "CredentialsFile": "~/.config/walgot/credentials.json",
    "DefaultListViewUnread": true,
    "DefaultListViewStarred": false,
    "DebugMode": true,
    "LogFile": "/tmp/walgot.log",
    "NbEntriesPerAPICall": 255,
    "DefaultSorting": "created",
    "DefaultOrder": "desc"
}
```

You only need to set the value you want to change in your configuration file, not everything.

- DefaultSorting: can only be 'created', 'updated' or 'archived', default 'created'
- DefaultOrder: can only be 'desc' or 'asc', default 'desc'

### credentials.json

In the `walgot.json` file above, we indicate the path to the credentials file for connecting to Wallabag. The format is as follow:

``` json
{
  "WallabagURL": "https://your.wallabag.instance",
  "ClientId": "client ID generate in your profile on wallabag"
  "ClientSecret": "client secrete generate in your profile on wallabag"
  "UserName": "your username",
  "UserPassword": "your password"
}
```

Default place is `~/.config/walgot/credentials.json` but can be changed in the `walgot.json` file.

## Usage:

### Start

``` help
Usage walgot:
  -config string
    	file name of config JSON file (default "~/.config/walgot/walgot.json")
  -d	enable debug output
  -version
    	get walgot version
```

example:

``` bash
/path/to/walgot -d -config "/my/config/file.json"
```


### Keybinds

``` 
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
```


## Remaining TODOs:

MVP:

- [x] Retrieve articles from wallabag
- [x] Articles list view
  - [x] Display all article in a scrollable table
  - [x] Filter entries:
    - [x] Only unread
    - [x] Only starred
    - [x] Only archived
- [x] Article detail view
  - [x] Display article in readable format (html2text + wrap)
- [x] Help view
- [x] Action on article
  - [x] On listing view
    - [x] Toggle Archive / Unread
    - [x] Toggle star 
  - [x] On detail view
    - [x] Toggle Archive / Unread
    - [x] Toggle star 
- [x] Configurable
  - [x] Load a json configuration file
  - [x] Make configuration file location configurable
    - [x] Manage shortpath (eg: "~/")

After MVP:

- [ ] Add Search
- [ ] Improve UI
  - [-] Improve article list view
    - [x] Improve table readability
    - [ ] Dynamic Sort table
      - [ ] By date
      - [ ] By title
  - [ ] Improve article view 
    - [ ] Add reading % in article view
  - [x] Display possible API errors in a dialog box
- [ ] Simplify Auto
  - [ ] setup create default configuration file
  - [ ] Wizard to create credentials.json ?
- [x] Add Configuration option
  - [x] Filters when starting
  - [x] Sort when starting
- [ ] Add entry
- [x] Open original article link
- [ ] Delete article

To Investigate:

- [ ] Add a CLI parameter for "cleaning wallabag" (= remove archived and unstarred articles older than X) → This will make the API faster. Or should it be a different program?
- [ ] Offline? Local cache?
- [ ] Manage tags ?
- [ ] Manage annotations ?
- [ ] Manage sharing as public link? (Does API even allow this?)
- [ ] STT for reading article?
- [ ] Images?
- [ ] Bulk updates?

## Thank you

- [Bubbletea](https://github.com/charmbracelet/bubbletea) a TUI framework from building rich TUI applications
- [Wallabago](https://github.com/Strubbl/wallabago) golang library for wallabag API
