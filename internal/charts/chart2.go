package charts

import (
	"database/sql"
	"sort"
	"strconv"
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

	// Build string labels for Chart.js
	yearLabels := make([]string, len(yearRange))
	for i, y := range yearRange {
		yearLabels[i] = strconv.Itoa(y)
	}

	datasets := make([]any, 0, len(setters))
	for _, name := range setters {
		values := make([]int, len(yearRange))
		for i, y := range yearRange {
			values[i] = data[name][y]
		}
		datasets = append(datasets, map[string]any{
			"label":       name,
			"data":        values,
			"fill":        true,
			"pointRadius": 0,
		})
	}

	chartDef := map[string]any{
		"type": "line",
		"data": map[string]any{
			"labels":   yearLabels,
			"datasets": datasets,
		},
		"options": map[string]any{
			"responsive":          true,
			"maintainAspectRatio": false,
			"animation":           false,
			"plugins": map[string]any{
				"tooltip": map[string]any{"enabled": false},
				"legend":  map[string]any{"display": true, "position": "bottom"},
			},
			"scales": map[string]any{
				"x": map[string]any{
					"ticks": map[string]any{
						"maxRotation": 0,
						"minRotation": 0,
					},
				},
				"y": map[string]any{
					"title": map[string]any{
						"display": true,
						"text":    "Crosswords per year",
					},
				},
			},
		},
	}

	tmplData := map[string]any{
		"Title":        "Crosswords per year, per setter",
		"Preamble":     "This chart shows an area span for the number of crosswords set per setter, per year.  Interesting to see when a setter started and stopped.",
		"Order":        2,
		"DivID":        "mychart2",
		"JSVar":        "chart2",
		"DefaultChart": "area",
		"Height":       800,
		"LegendHover":  true,
		"ChartJSON":    toJSON(chartDef),
	}
	return executeTemplate(tmplDir, "chart.tmpl", tmplData)
}
