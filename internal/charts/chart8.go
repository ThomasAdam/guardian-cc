package charts

import (
	"database/sql"
	"sort"
)

// Chart8 generates "Unique-answer ratio per setter" bar chart.
// A high ratio means the setter rarely repeats answers; a low ratio means
// many of their answers appear in multiple crosswords.
type Chart8 struct{}

func (c *Chart8) Order() string { return "8" }

func (c *Chart8) Render(db *sql.DB, tmplDir string) (string, error) {
	rows, err := db.Query(`
		SELECT c.creator_name AS name,
		       COUNT(DISTINCT e.solution)                        AS unique_solutions,
		       COUNT(e.solution)                                 AS total_solutions,
		       ROUND(
		           100.0 * COUNT(DISTINCT e.solution) / COUNT(e.solution),
		           1
		       ) AS ratio
		FROM entries e
		JOIN crosswords c ON e.crossword_id = c.id
		WHERE e.solution IS NOT NULL AND e.solution != ''
		GROUP BY c.creator_name
		HAVING COUNT(e.solution) > 0
		ORDER BY ratio DESC
	`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type row struct {
		name  string
		ratio float64
	}
	var results []row
	for rows.Next() {
		var name string
		var unique, total int
		var ratio float64
		if err := rows.Scan(&name, &unique, &total, &ratio); err != nil {
			return "", err
		}
		results = append(results, row{name, ratio})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ratio > results[j].ratio
	})

	labels := make([]string, len(results))
	values := []any{"Unique %"}
	for i, r := range results {
		labels[i] = r.name
		values = append(values, r.ratio)
	}

	chartDef := map[string]any{
		"bindto": "#mychart8",
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
				"label": "Unique answers (%)",
				"max":   100,
				"min":   0,
			},
		},
	}

	data := map[string]any{
		"Title":        "Unique-answer ratio per setter",
		"Preamble":     "The percentage of a setter's answers that are unique — i.e. used only once in their entire back-catalogue. A high percentage means a wider vocabulary; a low percentage means many repeated answers.",
		"Order":        8,
		"DivID":        "mychart8",
		"JSVar":        "chart8",
		"DefaultChart": "bar",
		"ChartJSON":    toJSON(chartDef),
		"BarColorJS":   barColorJS("chart8"),
	}
	return executeTemplate(tmplDir, "chart.tmpl", data)
}
