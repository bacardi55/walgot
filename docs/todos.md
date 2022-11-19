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

- [x] Add Search
- [ ] Improve UI
  - [-] Improve article list view
    - [x] Improve table readability
    - [x] Adapt table to screen size
    - [ ] Dynamic Sort table
      - [ ] By date
      - [ ] By title
    - [x] Add status in footer for easier readability
  - [x] Improve article view
    - [x] Add reading % in article view
    - [x] Make title static at the top
    - [ ] Better management for links to avoid breaking
    - [ ] Better management for images url
    - [x] Adapt reading view to screen size
  - [x] Display possible API errors in a dialog box
- [ ] Simplify start
  - [ ] setup create default configuration file
  - [ ] Wizard to create credentials.json ?
- [x] Add Configuration option
  - [x] Filters when starting
  - [x] Sort when starting
- [x] Add entry
- [x] Delete article
- [x] Open public/original article link
- [x] Manage sharing as public link
- [x] Yank/Copy URL
- [ ] Option to mark entry as read as soon as you read it 
- [ ] Add a way to open links in content


To Investigate:

- [ ] Offline? Local cache?
- [ ] Manage tags ?
- [ ] Manage annotations ?
- [ ] STT for reading article?
- [ ] Images?
- [ ] Bulk updates?

