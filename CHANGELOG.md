# Walgot Changelog

## v0.3.0 - Work in Progress (current developed version)

### New:

- Features:
  - Copy URL (original or public) to clipboard via Y keybind
  - Save a new entry on wallabag ("N")
  - Delete entry on wallabag ("D")
  - Search - Search for exact term (case insensitive) in article title ("/")
  - Filter for public articles in table view ("p")
  - Toggle for public status ("P")
  - Open article link in default browser ("O")
- UI improvements:
  - Listing view:
    - Adapt list view based on screen width to optimize info display
  - Article reading view:
    - Include all links footnotes instead of mid text
    - Adapt reading view if screen size is small
    - Display status (starred, new, public) in reading view footer
    - Improve detail view with fixed title and % read
    - Adapt footer depending on term height and width

### Bug fixes:

- Add notif after deleting an entry
- Add notif message after adding an entry
- Make scroll smoother when reading an article
- Prevent crash during reloading when trying to select an entry

### others

- Maintenance: Upgrade dependencies
- Add some unit tests (needs a lot more)
- Add automated build on sourcehut

