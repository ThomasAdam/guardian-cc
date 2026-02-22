package charts

import (
	"database/sql"
	"sort"
	"time"
)

// Chart2 generates "Crosswords per year, per setter" area chart.
type Chart2 struct{}

func (c *Chart2) Order() string { return "2" }

func (c *Chart2) Render(db *sql.DB, tmplDir string) (string, error) {
	rows, err := db.Query(`
		SELECT creator_name AS name,
		       CAST(EXTRACT(YEAR FROM date) AS INTEGER) AS year,
		       COUNT(*) AS count
		FROM crosswords
		GROUP BY creator_name, EXTRACT(YEAR FROM date)
		ORDER BY creator_name, year
	`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	// name -> year -> count
	data := make(map[string]map[int]int)
	for rows.Next() {
		var name string
		var year, count int
		if err := rows.Scan(&name, &year, &count); err != nil {
			return "", err
		}
		if data[name] == nil {
			data[name] = make(map[int]int)
		}
		data[name][year] = count
	}

	setters := make([]string, 0, len(data))
	for k := range data {
		setters = append(setters, k)
	}
	sort.Strings(setters)

	currentYear := time.Now().Year()
	yearRange := make([]int, 0)
	for y := 1998; y <= currentYear; y++ {
		yearRange = append(yearRange, y)
	}

	columns := make([]any, 0, len(setters))
	for _, name := range setters {
		row := []any{name}
		for _, y := range yearRange {
			row = append(row, data[name][y]) // 0 if missing
		}
		columns = append(columns, row)
	}

	chartDef := map[string]any{
		"bindto": "#mychart2",
		"size":   map[string]any{"height": 800},
		"data": map[string]any{
			"columns": columns,
			"type":    "area",
		},
		"tooltip": map[string]any{"show": false},
		"axis": map[string]any{
			"x": map[string]any{
				"type":       "category",
				"tick":       map[string]any{"rotate": "75", "multiline": false},
				"height":     0,
				"categories": yearRange,
			},
			"y": map[string]any{
				"label": "Crosswords per year",
			},
		},
	}

	tmplData := map[string]any{
		"Title":        "Crosswords per year, per setter",
		"Preamble":     "This chart shows an area span for the number of crosswords set per setter, per year.  Interesting to see when a setter started and stopped.  Hover over a legend entry to isolate that setter.",
		"Order":        2,
		"DivID":        "mychart2",
		"JSVar":        "chart2",
		"DefaultChart": "area",
		"ChartJSON":    toJSON(chartDef),
	}
	return executeTemplate(tmplDir, "chart2.tmpl", tmplData)
}
