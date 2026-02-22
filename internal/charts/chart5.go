package charts

import (
	"database/sql"
	"fmt"
	"strings"
)

// Chart5 generates "Duplicate answers" DataTable.
type Chart5 struct{}

func (c *Chart5) Order() string { return "5" }

func (c *Chart5) Render(db *sql.DB, tmplDir string) (string, error) {
	rows, err := db.Query(`
		WITH deduped AS (
			SELECT DISTINCT e.crossword_id, e.solution, e.clue,
			       c.creator_name, c.crossword_type,
			       c.id AS cw_path, c.number AS cw_number
			FROM resolved_entries e
			JOIN crosswords c ON e.crossword_id = c.id
			WHERE e.clue != ''
			  AND e.clue NOT SIMILAR TO '\s+\(\d+\)'
			  AND e.clue NOT SIMILAR TO 'See\s+\d+.*'
			  AND e.clue NOT SIMILAR TO 'See\s+(clues|special)\s+.*'
			  AND e.clue NOT SIMILAR TO 'Follow\s+the\s+link\s+below\s+to\s+see\s+today''s\s+clues.*'
		)
		SELECT creator_name AS name,
		       solution,
		       COUNT(*) AS count,
		       LIST(clue ORDER BY cw_number) AS clues,
		       LIST(crossword_type ORDER BY cw_number) AS types,
		       LIST('<a href="https://www.theguardian.com/' || cw_path || '">' || cw_number || '</a>' ORDER BY cw_number) AS urls
		FROM deduped
		GROUP BY creator_name, solution
		HAVING COUNT(DISTINCT crossword_id) > 1
		ORDER BY creator_name, solution
	`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var ajaxData [][]string
	for rows.Next() {
		var name, solution string
		var count int
		var rawClues, rawTypes, rawURLs any
		if err := rows.Scan(&name, &solution, &count, &rawClues, &rawTypes, &rawURLs); err != nil {
			return "", fmt.Errorf("scanning chart5 row: %w", err)
		}

		clues := toStringSlice(rawClues)
		types := toStringSlice(rawTypes)
		urls := toStringSlice(rawURLs)

		clueStr := strings.Join(clues, "<br />")
		typeStr := strings.Join(types, "<br />")
		urlStr := strings.Join(urls, "<br />")

		ajaxData = append(ajaxData, []string{name, solution, clueStr, typeStr, urlStr})
	}

	if err := writeAjaxFile("./ds_ajax.txt", ajaxData); err != nil {
		return "", err
	}

	columns := []map[string]string{
		{"title": "Setter"},
		{"title": "Answer"},
		{"title": "Clues"},
		{"title": "Type"},
		{"title": "Crossword"},
	}

	data := map[string]any{
		"Title":    "Number of duplicate answers and their questions",
		"Preamble": "This table shows the number of times a given clue has been used and the different questions which have been used to make up that clue.",
		"Order":    5,
		"Columns":  toJSON(columns),
	}
	return executeTemplate(tmplDir, "chart5.tmpl", data)
}
