package charts

import (
	"database/sql"
	"sort"
)

// Chart3 generates "Frequency of word duplications" bar chart.
type Chart3 struct{}

func (c *Chart3) Order() string { return "3" }

func (c *Chart3) Render(db *sql.DB, tmplDir string) (string, error) {
	rows, err := db.Query(`
		SELECT name, MAX(cnt) AS max_count
		FROM (
			SELECT c.creator_name AS name,
			       e.solution,
			       COUNT(*) AS cnt
			FROM entries e
			JOIN crosswords c ON e.crossword_id = c.id
			GROUP BY c.creator_name, e.solution
			HAVING COUNT(*) > 1
		) sub
		GROUP BY name
		ORDER BY max_count DESC
	`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type entry struct {
		name  string
		count int
	}
	var results []entry
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.name, &e.count); err != nil {
			return "", err
		}
		results = append(results, e)
	}

	// Sort by count descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].count > results[j].count
	})

	labels := make([]string, len(results))
	values := []any{"Setters"}
	for i, r := range results {
		labels[i] = r.name
		values = append(values, r.count)
	}

	chartDef := map[string]any{
		"bindto": "#myChart3",
		"size":   map[string]any{"height": 800},
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
				"label": "Frequency of duplicated answers",
				"tick":  map[string]any{"steps": 20},
			},
		},
	}

	data := map[string]any{
		"Title":        "Frequency of word duplications across all crosswords, per setter",
		"Preamble":     "This chart shows the number of words a given setter has used more than once, across all crosswords for that setter.",
		"Order":        3,
		"DivID":        "mychart3",
		"JSVar":        "chart3",
		"DefaultChart": "bar",
		"ChartJSON":    toJSON(chartDef),
	}
	return executeTemplate(tmplDir, "chart.tmpl", data)
}
