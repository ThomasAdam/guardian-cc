package charts

import (
	"database/sql"
)

// Chart7 generates "Most-used answers across all crosswords" bar chart.
type Chart7 struct{}

func (c *Chart7) Order() string { return "7" }

func (c *Chart7) Render(db *sql.DB, tmplDir string) (string, error) {
	rows, err := db.Query(`
		SELECT e.solution,
		       COUNT(*) AS cnt
		FROM entries e
		WHERE e.solution IS NOT NULL AND e.solution != ''
		GROUP BY e.solution
		ORDER BY cnt DESC
		LIMIT 50
	`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	labels := make([]string, 0, 50)
	values := []any{"Count"}
	for rows.Next() {
		var solution string
		var cnt int
		if err := rows.Scan(&solution, &cnt); err != nil {
			return "", err
		}
		labels = append(labels, solution)
		values = append(values, cnt)
	}

	chartDef := map[string]any{
		"bindto": "#mychart7",
		"size":   map[string]any{"height": 600},
		"data": map[string]any{
			"columns": []any{values},
			"type":    "bar",
		},
		"axis": map[string]any{
			"x": map[string]any{
				"type":       "category",
				"tick":       map[string]any{"rotate": "75", "multiline": false},
				"height":     0,
				"categories": labels,
			},
			"y": map[string]any{
				"label": "Times used across all crosswords",
			},
		},
	}

	data := map[string]any{
		"Title":        "Most-used answers across all crosswords (top 50)",
		"Preamble":     "The top 50 solutions that appear most frequently across every crossword in the archive, regardless of setter.  These are the classic crossword chestnuts.",
		"Order":        7,
		"DivID":        "mychart7",
		"JSVar":        "chart7",
		"DefaultChart": "bar",
		"ChartJSON":    toJSON(chartDef),
		"BarColorJS":   barColorJS("chart7"),
	}
	return executeTemplate(tmplDir, "chart.tmpl", data)
}
