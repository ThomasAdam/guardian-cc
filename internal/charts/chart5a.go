package charts

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

// Chart5a generates "Duplicate clues" DataTable.
type Chart5a struct{}

func (c *Chart5a) Order() string { return "5a" }

// skipClueRE matches residual placeholder clues that should be filtered out.
// Most cross-references are resolved by the resolved_entries view; this
// catches remaining edge cases (length-only entries, instruction links).
var skipClueRE = regexp.MustCompile(
	`^\s+\(\d+\)$` +
		`|^Follow\s+the\s+link\s+below\s+to\s+see\s+today's\s+clues.*$`,
)

func (c *Chart5a) Render(db *sql.DB, tmplDir string) (string, error) {
	rows, err := db.Query(`
		WITH deduped AS (
			SELECT DISTINCT e.crossword_id, e.clue, c.creator_name,
			       c.crossword_type, c.id AS cw_path, c.number AS cw_number
			FROM resolved_entries e
			JOIN crosswords c ON e.crossword_id = c.id
			WHERE e.clue != ''
			  AND e.clue NOT SIMILAR TO '\s+\(\d+\)'
			  AND e.clue NOT SIMILAR TO 'See\s+\d+.*'
			  AND e.clue NOT SIMILAR TO 'See\s+(clues|special)\s+.*'
			  AND e.clue NOT SIMILAR TO 'Follow\s+the\s+link\s+below\s+to\s+see\s+today''s\s+clues.*'
		)
		SELECT creator_name AS name,
		       clue,
		       COUNT(*) AS count,
		       LIST(clue ORDER BY cw_number) AS clues,
		       LIST(crossword_type ORDER BY cw_number) AS types,
		       LIST('<a href="https://www.theguardian.com/' || cw_path || '">' || cw_number || '</a>' ORDER BY cw_number) AS urls
		FROM deduped
		GROUP BY creator_name, clue
		HAVING COUNT(DISTINCT crossword_id) > 1
		ORDER BY creator_name, clue
	`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var ajaxData [][]string
	for rows.Next() {
		var name, clue string
		var count int
		var rawClues, rawTypes, rawURLs any
		if err := rows.Scan(&name, &clue, &count, &rawClues, &rawTypes, &rawURLs); err != nil {
			return "", fmt.Errorf("scanning chart5a row: %w", err)
		}

		clues := toStringSlice(rawClues)
		types := toStringSlice(rawTypes)
		urls := toStringSlice(rawURLs)

		// Filter out residual placeholder clues
		var filtered []string
		for _, cl := range clues {
			if !skipClueRE.MatchString(cl) {
				filtered = append(filtered, cl)
			}
		}
		if len(filtered) == 0 || filtered[0] == "" {
			continue
		}

		clueStr := strings.Join(filtered, "<br />")
		typeStr := strings.Join(types, "<br />")
		urlStr := strings.Join(urls, "<br />")

		ajaxData = append(ajaxData, []string{name, clueStr, typeStr, urlStr})
	}

	if err := writeAjaxFile("./ds_ajax5a.txt", ajaxData); err != nil {
		return "", err
	}

	columns := []map[string]string{
		{"title": "Setter"},
		{"title": "Clues"},
		{"title": "Type"},
		{"title": "Crossword"},
	}

	data := map[string]any{
		"Title":    "Number of duplicate clues per setter",
		"Preamble": "This table shows the number of times a given clue has been used per setter.",
		"Order":    "5a",
		"Columns":  toJSON(columns),
	}
	return executeTemplate(tmplDir, "chart5a.tmpl", data)
}
