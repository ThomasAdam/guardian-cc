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

Since the crosswords are JSON documents, they are imported into
[DuckDB](https://duckdb.org/) via a normalised relational schema.  The import
and chart rendering tool is written in Go, using
[go-duckdb](https://github.com/marcboeker/go-duckdb) for database access and
Go's `html/template` / `text/template` packages for HTML generation.  Charts
are rendered client-side via [c3js](https://c3js.org).

## Building

```
go build -o guardian-cc ./cmd/guardian-cc/main.go
```

## Usage

```
# Import all crossword JSON files
./guardian-cc import

# Import a single file
./guardian-cc import crosswords/cryptic/setter/Rufus/21625.JSON

# Render charts to gcc-analysis.html
./guardian-cc render

# Serve the page with server-side pagination (default :8080)
./guardian-cc serve

# Serve on a custom address
./guardian-cc serve :3000
```

The `serve` command starts an HTTP server that serves `gcc-analysis.html` and
provides a `/api/dt` endpoint for DataTables server-side processing.  Charts
5, 5a, and 6 use this to paginate, sort, and search their large datasets
directly against DuckDB instead of loading everything into the browser at once.

Patches and ideas for graphs welcome!

-- Thomas Adam
