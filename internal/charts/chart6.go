package charts

import (
	"database/sql"
	"fmt"
	"strings"
)

// Chart6 generates "PDF crossword list" DataTable.
type Chart6 struct{}

func (c *Chart6) Order() string { return "6" }

func (c *Chart6) Render(db *sql.DB, tmplDir string) (string, error) {
	rows, err := db.Query(`
		SELECT creator_name AS name,
		       number,
		       pdf,
		       CAST(date AS VARCHAR) AS date
		FROM crosswords
		WHERE pdf IS NOT NULL
		ORDER BY creator_name, number
	`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var ajaxData [][]string
	for rows.Next() {
		var name, number, pdf, date string
		if err := rows.Scan(&name, &number, &pdf, &date); err != nil {
			return "", err
		}
		// Trim to just the date portion
		date = strings.SplitN(date, " ", 2)[0]

		link := fmt.Sprintf(`<a href="%s">%s</a>`, pdf, number)
		ajaxData = append(ajaxData, []string{name, link, date})
	}

	if err := writeAjaxFile("./ds_ajax2.txt", ajaxData); err != nil {
		return "", err
	}

	columns := []map[string]string{
		{"title": "Setter"},
		{"title": "Crossword (PDF)"},
		{"title": "Date published"},
	}

	data := map[string]any{
		"Title":    "List of all crosswords by setter, which has a PDF version",
		"Preamble": "This table shows the crossword number and a link to the PDF crossword, if available.",
		"Order":    6,
		"Columns":  toJSON(columns),
	}
	return executeTemplate(tmplDir, "chart6.tmpl", data)
}
