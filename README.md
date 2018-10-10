Guardian Cryptic Crossword Analysis
===================================

[The Guardian](https://www.theguardian.com) has been publishing cryptic
crosswords for well over fifty years.  During that time, there have been a
plethora of different setters, all with their unique style.

Unfortunately, the Guardian only has crosswords going back to the year 1999 on
their website.  So the analysis can only go that far back.

This repository contains JSON files for all the crosswords the Guardian has
hosted.

You can see [some graphs of this data here](https://xteddy.org/gcc-analysis.html)

These charts are updated daily.

Since the crosswords are JSON documents, the backend to storing them is in
[mongodb](https://www.mongodb.com/).  Then, using the the mongodb Perl API,
and Template::Toolkit, the static page is generated, and then the charts are
rendered client-side via [c3js](https://c3js.org).

The following perl modules are used:

```
DateTime
DateTime::Format::Duration
List::Util
Module::Pluggable
MongoDB
Sort::Key::DateTime

The chart-rendering namespace is:

```
Guardian::Cryptic::Crosswords
```

There's also a `Plugins` directory, containing the template for each graph.

Patches and ideas for graphs welcome!

-- Thomas Adam
