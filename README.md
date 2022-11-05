# Walgot - a WALlabag GO Tui client

**The status of this code is still very early stage and work in progress. This isn't yet at MVP stage so be carreful while using it!**

Official repository and project is on [codeberg](https://codeberg.org/bacardi55/walgot). Github is only a mirror.

## installation

### binary

Download one of the binary available on the binary file in the release page. 
Make sure it is executable (with `chmod +x binaryFile`) and then run walgot with the command:

### compile
Use the makefile provided and run `make build`, it will create a binary file in a `bin` directory.

``` bash
git clone https://… # TODO
cd walgot
make build
mkdir ~/.config/walgot/
cp example/*.json ~/.config/walgot/
# NOTA: Here you need to edit at least ~/.config/walgot/credentials.json
# Then start walgot:
./bin/walgot
```

## configuration
Walgot is configure with 2 different files:

### walgot.json

The main configuration file, in json format. The default value are as below:

``` json
{
    "CredentialsFile": "~/.config/walgot/credentials.json",
    "DefaultListViewUnread": true,
    "DefaultListViewStarred": false,
    "DebugMode": true,
    "LogFile": "/tmp/walgot.log",
    "NbEntriesPerAPICall": 255
}
```

You only need to set the value you want to change in your configuration file, not everything.

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
	- ctrl+c: quit
	- h: help (this page)


	On listing page:
	- r: reload article from wallabag via APIs, takes time depending on the number of articles saved
	- u: toggle display only unread articles (disable archived filter)
	- s: toggle display only starred articles
	- a: toggle archived only articles (disable unread filter)
	- h: display help
	- ↑ or k / ↓ or j: move up / down one item in the list
	- page down / page up: move up / down 10 items in the list
	- home: go to the top of the list
	- end: go to bottom of the list
	- enter: select entry to read content
	- q: quit

	On detail page:
	- q: return to list
	- ↑ or k / ↓ or j: go up / down

    On dialog (modal) view:
    - "enter" or "esc": close the dialog

	On help page:
	- q: return to list
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
- [ ] Action on article
  - [ ] Archive (mark as read)
  - [ ] Mark as unread
  - [ ] Toggle star 
- [x] Configurable
  - [x] Load a json configuration file
  - [x] Make configuration file location configurable
    - [x] Manage shortpath (eg: "~/")


After MVP:

- [ ] Improve article list view
  - [ ] Sort entries
    - [ ] By date
    - [ ] By title
- [ ] Add Search
- [ ] Improve UI
  - [ ] Improve table readability
  - [ ] Improve article view 
- [ ] Auto create default configuration file
- [ ] Wizard to create credentials.json ?
- [ ] Add Configuration option
- [x] Display possible API errors in a dialog box

To Investigate:

- [ ] Offline? Local cache?
- [ ] Manage tags ?
- [ ] Manage annotations ?
- [ ] Manage sharing as public link? (Does API even allow this?)


## Thank you

- [Bubbletea](https://github.com/charmbracelet/bubbletea) a TUI framework from building rich TUI applications
- [Wallabago](https://github.com/Strubbl/wallabago) golang library for wallabag API
