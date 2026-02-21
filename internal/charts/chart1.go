package charts

import (
	"database/sql"
	"sort"
)

// Chart1 generates "Total number of crosswords, set by author" bar chart.
type Chart1 struct{}

func (c *Chart1) Order() string { return "1" }

type setterTypeCount struct {
	pos   int
	count int
}

func (c *Chart1) Render(db *sql.DB, tmplDir string) (string, error) {
	rows, err := db.Query(`
		SELECT creator_name AS name,
		       crossword_type AS type,
		       COUNT(*) AS count
		FROM crosswords
		GROUP BY creator_name, crossword_type
		ORDER BY count DESC, crossword_type ASC
	`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type nameMap struct {
		cryptic setterTypeCount
		prize   setterTypeCount
	}
	nm := make(map[string]*nameMap)

	position := 0
	for rows.Next() {
		var name, ctype string
		var count int
		if err := rows.Scan(&name, &ctype, &count); err != nil {
			return "", err
		}
		if nm[name] == nil {
			nm[name] = &nameMap{}
		}
		position++
		switch ctype {
		case "cryptic":
			nm[name].cryptic = setterTypeCount{pos: position, count: count}
		case "prize":
			nm[name].prize = setterTypeCount{pos: position, count: count}
		}
	}

	// Fill in missing types
	for _, v := range nm {
		if v.cryptic.pos == 0 && v.prize.pos != 0 {
			v.cryptic.pos = v.prize.pos
		}
		if v.prize.pos == 0 && v.cryptic.pos != 0 {
			v.prize.pos = v.cryptic.pos
		}
	}

	// Sort by cryptic position
	type kv struct {
		name string
		data *nameMap
	}
	sorted := make([]kv, 0, len(nm))
	for k, v := range nm {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].data.cryptic.pos < sorted[j].data.cryptic.pos
	})

	labels := []any{"x"}
	cryptic := []any{"Cryptic"}
	prize := []any{"Prize"}
	for _, s := range sorted {
		labels = append(labels, s.name)
		cryptic = append(cryptic, s.data.cryptic.count)
		prize = append(prize, s.data.prize.count)
	}
	columns := []any{labels, cryptic, prize}

	chartDef := map[string]any{
		"bindto": "#mychart1",
		"size":   map[string]any{"height": 800},
		"data": map[string]any{
			"x":       "x",
			"columns": columns,
			"type":    "bar",
			"empty":   map[string]any{"label": map[string]any{"text": "Unknown"}},
			"groups":  [][]string{{"Prize", "Cryptic"}},
		},
		"axis": map[string]any{
			"x": map[string]any{
				"type":   "category",
				"tick":   map[string]any{"rotate": "75", "multiline": false},
				"height": 0,
			},
			"y": map[string]any{
				"label": "No. of crosswords set",
				"max":   800,
				"tick":  map[string]any{"steps": 20},
			},
		},
	}

	data := map[string]any{
		"Title":        "Total number of crosswords, set by author",
		"Preamble":     "This chart shows the number of crosswords set per setter.  No real surprises here as to the most prolific setters.",
		"Order":        1,
		"DivID":        "mychart1",
		"JSVar":        "chart1",
		"DefaultChart": "bar",
		"ChartJSON":    toJSON(chartDef),
	}
	return executeTemplate(tmplDir, "chart.tmpl", data)
}
