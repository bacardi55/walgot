# TODOs:

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
  - [x] Improve article view 
    - [x] Add reading % in article view
    - [x] Make title static at the top
  - [x] Display possible API errors in a dialog box
- [ ] Simplify start
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

