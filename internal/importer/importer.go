package importer

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CrosswordJSON matches the JSON structure from the Guardian scraper.
type CrosswordJSON struct {
	ID      string `json:"id"`
	Number  any    `json:"number"` // can be int or string
	Name    string `json:"name"`
	Creator struct {
		Name   string `json:"name"`
		WebURL string `json:"webUrl"`
	} `json:"creator"`
	Date               float64        `json:"date"` // epoch milliseconds
	WebPublicationDate float64        `json:"webPublicationDate"`
	Entries            []EntryJSON    `json:"entries"`
	CrosswordType      string         `json:"crosswordType"`
	PDF                *string        `json:"pdf"`
	Dimensions         *DimensionJSON `json:"dimensions"`
}

type EntryJSON struct {
	ID          string `json:"id"`
	Number      int    `json:"number"`
	HumanNumber string `json:"humanNumber"`
	Clue        string `json:"clue"`
	Direction   string `json:"direction"`
	Length      int    `json:"length"`
	Solution    string `json:"solution"`
	Position    struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"position"`
}

type DimensionJSON struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

// Import imports one or more JSON files into the database.
// If files is empty, it walks the default crossword directories.
func Import(db *sql.DB, files []string) error {
	if len(files) == 0 {
		dirs := []string{
			"./crosswords/cryptic/setter",
			"./crosswords/prize/setter",
		}
		for _, dir := range dirs {
			err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && strings.HasSuffix(strings.ToUpper(d.Name()), ".JSON") {
					files = append(files, path)
				}
				return nil
			})
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("walking %s: %w", dir, err)
			}
		}
	}

	insCW, err := db.Prepare(`INSERT OR IGNORE INTO crosswords
		(id, number, name, creator_name, creator_weburl, date, crossword_type, pdf)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing crossword insert: %w", err)
	}
	defer insCW.Close()

	insEntry, err := db.Prepare(`INSERT INTO entries
		(crossword_id, entry_id, number, human_number, clue, direction, length, solution, pos_x, pos_y)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing entry insert: %w", err)
	}
	defer insEntry.Close()

	existsStmt, err := db.Prepare("SELECT 1 FROM crosswords WHERE id = ?")
	if err != nil {
		return fmt.Errorf("preparing exists check: %w", err)
	}
	defer existsStmt.Close()

	for _, f := range files {
		if err := importFile(db, f, insCW, insEntry, existsStmt); err != nil {
			fmt.Fprintf(os.Stderr, "Error importing %s: %v\n", f, err)
			continue
		}
	}
	return nil
}

func importFile(db *sql.DB, path string, insCW, insEntry, existsStmt *sql.Stmt) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var cw CrosswordJSON
	if err := json.Unmarshal(data, &cw); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}

	// Normalise creator
	if cw.Creator.Name == "" {
		cw.Creator.Name = "Unknown"
	}
	cw.Creator.Name = strings.TrimRight(cw.Creator.Name, " \t")
	if cw.Creator.WebURL == "" {
		cw.Creator.WebURL = "http://www.example.org"
	}

	// Convert number to string
	numStr := fmt.Sprintf("%v", cw.Number)

	// Convert epoch-ms date
	t := time.UnixMilli(int64(cw.Date)).UTC()
	dateStr := t.Format("2006-01-02")

	// Skip if already imported
	var dummy int
	err = existsStmt.QueryRow(cw.ID).Scan(&dummy)
	if err == nil {
		fmt.Printf("Skipped (exists): %s\n", cw.ID)
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	var pdf *string
	if cw.PDF != nil {
		pdf = cw.PDF
	}

	if _, err := tx.Stmt(insCW).Exec(
		cw.ID, numStr, cw.Name,
		cw.Creator.Name, cw.Creator.WebURL,
		dateStr, cw.CrosswordType, pdf,
	); err != nil {
		tx.Rollback()
		return fmt.Errorf("inserting crossword: %w", err)
	}

	entryStmt := tx.Stmt(insEntry)
	for _, e := range cw.Entries {
		if _, err := entryStmt.Exec(
			cw.ID, e.ID, e.Number, e.HumanNumber,
			e.Clue, e.Direction, e.Length, e.Solution,
			e.Position.X, e.Position.Y,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("inserting entry: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	fmt.Printf("Added: %s\n", cw.ID)
	return nil
}
