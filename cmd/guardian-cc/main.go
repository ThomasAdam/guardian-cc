package main

import (
	"fmt"
	"os"

	"github.com/ThomasAdam/guardian-cc/internal/charts"
	"github.com/ThomasAdam/guardian-cc/internal/db"
	"github.com/ThomasAdam/guardian-cc/internal/importer"
	"github.com/ThomasAdam/guardian-cc/internal/server"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: guardian-cc <command> [args...]\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  import [file.JSON ...]  Import crossword JSON files into DuckDB\n")
	fmt.Fprintf(os.Stderr, "                          With no args, imports all files under crosswords/\n")
	fmt.Fprintf(os.Stderr, "  render                  Render charts to gcc-analysis.html\n")
	fmt.Fprintf(os.Stderr, "  serve [addr]            Serve the analysis page with server-side pagination\n")
	fmt.Fprintf(os.Stderr, "                          Default addr is :8080\n")
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "import":
		runImport(os.Args[2:])
	case "render":
		runRender()
	case "serve":
		addr := ":8080"
		if len(os.Args) > 2 {
			addr = os.Args[2]
		}
		runServe(addr)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		usage()
	}
}

func runImport(files []string) {
	database, err := db.Open(db.DefaultDBFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	if err := db.CreateSchema(database); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating schema: %v\n", err)
		os.Exit(1)
	}

	if err := importer.Import(database, files); err != nil {
		fmt.Fprintf(os.Stderr, "Error importing: %v\n", err)
		os.Exit(1)
	}
}

func runRender() {
	database, err := db.Open(db.DefaultDBFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	if err := db.CreateSchema(database); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating schema: %v\n", err)
		os.Exit(1)
	}

	tmplDir := "ui/chart_defs"
	outputFile := "./gcc-analysis.html"

	if err := charts.RenderAll(database, tmplDir, outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering charts: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Written: %s\n", outputFile)
}

func runServe(addr string) {
	database, err := db.Open(db.DefaultDBFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	if err := db.CreateSchema(database); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating schema: %v\n", err)
		os.Exit(1)
	}

	if err := server.Serve(database, addr); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
