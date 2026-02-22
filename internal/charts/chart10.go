package charts

import (
	"database/sql"
	"sort"
)

// Chart10 generates "Across vs Down balance per setter" stacked bar chart.
type Chart10 struct{}

func (c *Chart10) Order() string { return "10" }

func (c *Chart10) Render(db *sql.DB, tmplDir string) (string, error) {
	rows, err := db.Query(`
		SELECT c.creator_name AS name,
		       e.direction,
		       COUNT(*) AS cnt
		FROM entries e
		JOIN crosswords c ON e.crossword_id = c.id
		WHERE e.direction IN ('across', 'down')
		GROUP BY c.creator_name, e.direction
		ORDER BY c.creator_name, e.direction
	`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type counts struct {
		across int
		down   int
	}
	data := make(map[string]*counts)
	for rows.Next() {
		var name, direction string
		var cnt int
		if err := rows.Scan(&name, &direction, &cnt); err != nil {
			return "", err
		}
		if data[name] == nil {
			data[name] = &counts{}
		}
		switch direction {
		case "across":
			data[name].across = cnt
		case "down":
			data[name].down = cnt
		}
	}

	// Sort setters by total clues (across+down) descending, most prolific first.
	type kv struct {
		name   string
		across int
		down   int
	}
	sorted := make([]kv, 0, len(data))
	for k, v := range data {
		sorted = append(sorted, kv{k, v.across, v.down})
	}
	sort.Slice(sorted, func(i, j int) bool {
		ti := sorted[i].across + sorted[i].down
		tj := sorted[j].across + sorted[j].down
		return ti > tj
	})

	labels := []any{"x"}
	across := []any{"Across"}
	down := []any{"Down"}
	for _, s := range sorted {
		labels = append(labels, s.name)
		across = append(across, s.across)
		down = append(down, s.down)
	}

	chartDef := map[string]any{
		"bindto": "#mychart10",
		"size":   map[string]any{"height": 600},
		"data": map[string]any{
			"colors":  map[string]string{"Across": "#4e79a7", "Down": "#f28e2b"},
			"x":       "x",
			"columns": []any{labels, across, down},
			"type":    "bar",
			"groups":  [][]string{{"Across", "Down"}},
		},
		"axis": map[string]any{
			"x": map[string]any{
				"type":   "category",
				"tick":   map[string]any{"rotate": "75", "multiline": false},
				"height": 0,
			},
			"y": map[string]any{
				"label": "Number of clues",
			},
		},
	}

	tmplData := map[string]any{
		"Title":        "Across vs Down clue balance per setter",
		"Preamble":     "A stacked bar showing the total number of Across and Down clues each setter has written. For a standard 15×15 grid you would expect roughly equal numbers, so large imbalances can indicate a preference for one direction or a different grid style.",
		"Order":        10,
		"DivID":        "mychart10",
		"JSVar":        "chart10",
		"DefaultChart": "bar",
		"ChartJSON":    toJSON(chartDef),
	}
	return executeTemplate(tmplDir, "chart.tmpl", tmplData)
}
