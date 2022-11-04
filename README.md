# Walgot - a WALlabag GO Tui client

**The status of this code is still very early stage and work in progress. This isn't yet at MVP stage so be carreful while using it!**

## installation
TODO

## configuration
TODO

``` json
{
  "WallabagURL": "https://your.wallabag.instance",
  "ClientId": "client ID generate in your profile on wallabag"
  "ClientSecret": "client secrete generate in your profile on wallabag"
  "UserName": "your username",
  "UserPassword": "your password"
}
```

## Usage:

Everywhere:
- ctrl+c: quit
- h: help


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


## Remaining TODOs:

MVP:

- [x] Retrieve articles from wallabag
- [-] Articles list view
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
- [ ] Configurable
  - [x] Load a json configuration file
  - [ ] Make configuration file location configurable
    - [ ] Manage shortpath (eg: "~/")


After MVP:

- [ ] Improve article list view
  - [ ] Sort entries
    - [ ] By date
    - [ ] By title
- [ ] Add Search
- [ ] Improve UI
  - [ ] Improve table readability
  - [ ] Improve article view 


To Investigate:

- [ ] Offline? Local cache?
- [ ] Manage tags ?
- [ ] Manage annotations ?
- [ ] Manage sharing as public link? (Does API even allow this?)


## Thank you

- [Bubbletea](https://github.com/charmbracelet/bubbletea) a TUI framework from building rich TUI applications
- [Wallabago](https://github.com/Strubbl/wallabago) golang library for wallabag API
