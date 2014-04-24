Born out of a frustration of working with bugzilla, bugzini is a light-weight front-end
with a primary focus on fast and easy searching and reading of bug reports.

# Installation
bugzini is a self-contained webservice which serves a simple web application in the
local browser. The only build dependency of bugzini is `go`, so make sure you have it
installed first. Then, simply type `make` to make the self-contained `bugzini` application,
containing the webservice and all required resources bundled inside.

# Running
bugzini uses the XML-RPC API of bugzilla, and is currently only tested with the GNOME
bugzilla (running version 3.4). It should work with other bugzilla's as well, if it does
not, please report issues. To run bugzini, simply type `./bugzini` which starts the service
and prints the uri at which bugzini is served. Use `-l` to automatically launch the default
browser with the bugzini uri. To change the remote bugzilla, use `--bz-host` to specify
an alternative host name.

# Getting started
When running bugzini initially, it might take some time to load the bugzilla projects. After
the initial load, all projects are shown in a (searchable) sidebar on the left. To view bugs
of a certain product, simply select it from the list. When a product is selected for the first
time, all open bugs for that product are requested from the remote bugzilla. Again, this might
take some time depending on the number of bugs and remote bugzilla.

After the initial load, everything is stored locally in the browser using IndexedDB and subsequent
visits to the page should load locally and quickly without needing to contact bugzilla.

## Search
bugzini has a special focus on making it easy to quickly search for bug reports. The search
entry at the top of the page provides search as you type. A simple query language allows
for searching in bug fields other than its summary (such as the severity or component) and
features simple and/or type of queries.

  * `term`: search for `term`
  * `term other`: search for `term` `OR` `other`
  * `"one term"`: search for `one term`
  * `field:term`: search for `term` in the specified field
  * `!term other`: search for `term` `AND` `other`
  * `!(term other) next`: search for (`term` OR `other`) `AND` `next`

The `!` indicates that what follows is a necessary criteria for the search and allows for easy
construction of `AND` type queries. For example, to look for all bugs containing `crash` with
a severity of `critical` or `major`, use: `!(severity:critical severity:major) crash`

Note that search will only look at locally stored bug reports and only in the products selected
in the sidebar.

## Starring
Projects can be starred by clicking on the star on the left of the project name. Starred projects
are sorted in the sidebar before other projects, and provide an easy way to quickly navigate
projects that you are personally interested in. The special `All Starred` item in the sidebar
will show bugs from all the starred products at once.

## Creating bookmarks
Apart from products, custom search queries can be bookmarked and are shown in the sidebar.
To bookmark a current search, click on the magnifying glass icon of the main bug search entry.
A popup allows to type in a name for the new bookmark, which is then added in the sidebar. Selecting
such bookmarks will show all bugs corresponding to the search criteria stored in the bookmark.
Naturally, you can continue searching within this selection by using the search bar.

## Synchronizing
After the initial load of bug reports for a product, subsequent visits will only look at locally
stored bugs. Synchronization of reports for a product is currently triggered manually by the user
by clicking on the refresh icon on the right of a product name (in the sidebar). This might change
in the future to a more automatic, periodic synchronization.
