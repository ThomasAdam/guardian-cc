package charts

import (
	"database/sql"
	"sort"
)

// Chart9 generates "Average clue length per setter" bar chart.
// Clue length is measured in characters (excluding the trailing length hint
// such as "(6)" or "(3,4)"), giving a sense of how elaborate each setter's
// clue-writing style is.
type Chart9 struct{}

func (c *Chart9) Order() string { return "9" }

func (c *Chart9) Render(db *sql.DB, tmplDir string) (string, error) {
	// Strip the trailing length hint "(N)" / "(N,M)" / "(N-M)" before measuring.
	// regexp_replace removes the last parenthesised group and any leading/trailing
	// whitespace so we measure only the clue text proper.
	rows, err := db.Query(`
		SELECT c.creator_name AS name,
		       ROUND(
		           AVG(LENGTH(TRIM(regexp_replace(e.clue, '\s*\([\d,\-]+\)\s*$', '')))),
		           1
		       ) AS avg_len
		FROM entries e
		JOIN crosswords c ON e.crossword_id = c.id
		WHERE e.clue IS NOT NULL AND e.clue != ''
		GROUP BY c.creator_name
		ORDER BY avg_len DESC
	`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type row struct {
		name   string
		avgLen float64
	}
	var results []row
	for rows.Next() {
		var name string
		var avgLen float64
		if err := rows.Scan(&name, &avgLen); err != nil {
			return "", err
		}
		results = append(results, row{name, avgLen})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].avgLen > results[j].avgLen
	})

	labels := make([]string, len(results))
	values := []any{"Avg clue length (chars)"}
	for i, r := range results {
		labels[i] = r.name
		values = append(values, r.avgLen)
	}

	chartDef := map[string]any{
		"bindto": "#mychart9",
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
				"label": "Average clue length (characters)",
			},
		},
	}

	data := map[string]any{
		"Title":        "Average clue length per setter",
		"Preamble":     "Mean character-count of clue text per setter (the trailing length hint such as \"(6)\" is excluded). Longer clues tend to indicate more elaborate cryptic constructions or surface readings.",
		"Order":        9,
		"DivID":        "mychart9",
		"JSVar":        "chart9",
		"DefaultChart": "bar",
		"ChartJSON":    toJSON(chartDef),
		"BarColorJS":   barColorJS("chart9"),
	}
	return executeTemplate(tmplDir, "chart.tmpl", data)
}
