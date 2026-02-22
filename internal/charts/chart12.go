package charts

import (
	"database/sql"
)

// Chart12 generates "Crosswords published by month of year" bar chart.
// Shows the seasonal publication cadence across all years and setters.
type Chart12 struct{}

func (c *Chart12) Order() string { return "12" }

// monthNames maps month numbers 1–12 to abbreviated names.
var monthNames = []string{
	"Jan", "Feb", "Mar", "Apr", "May", "Jun",
	"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
}

func (c *Chart12) Render(db *sql.DB, tmplDir string) (string, error) {
	rows, err := db.Query(`
		SELECT CAST(EXTRACT(MONTH FROM date) AS INTEGER) AS month,
		       COUNT(*) AS cnt
		FROM crosswords
		GROUP BY month
		ORDER BY month
	`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	// Index 0 = January (month 1)
	counts := make([]int, 12)
	for rows.Next() {
		var month, cnt int
		if err := rows.Scan(&month, &cnt); err != nil {
			return "", err
		}
		if month >= 1 && month <= 12 {
			counts[month-1] = cnt
		}
	}

	labels := make([]string, 12)
	copy(labels, monthNames)

	values := []any{"Crosswords"}
	for _, v := range counts {
		values = append(values, v)
	}

	chartDef := map[string]any{
		"bindto": "#mychart12",
		"size":   map[string]any{"height": 400},
		"data": map[string]any{
			"columns": []any{values},
			"type":    "bar",
		},
		"axis": map[string]any{
			"x": map[string]any{
				"type":       "category",
				"categories": labels,
			},
			"y": map[string]any{
				"label": "Number of crosswords published",
			},
		},
	}

	data := map[string]any{
		"Title":        "Crosswords published by month of year",
		"Preamble":     "Total number of crosswords published in each calendar month, summed across all years and setters.  Dips in August and December can reflect holiday periods when fewer puzzles are commissioned.",
		"Order":        12,
		"DivID":        "mychart12",
		"JSVar":        "chart12",
		"DefaultChart": "bar",
		"ChartJSON":    toJSON(chartDef),
		"BarColorJS":   barColorJS("chart12"),
	}
	return executeTemplate(tmplDir, "chart.tmpl", data)
}
