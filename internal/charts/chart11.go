package charts

import (
	"database/sql"
	"sort"
	"time"
)

// Chart11 generates "New setters per year" bar chart.
// A setter's debut year is the earliest year in which they appear in the DB.
// The tooltip lists every setter who debuted in the hovered bar's year.
type Chart11 struct{}

func (c *Chart11) Order() string { return "11" }

func (c *Chart11) Render(db *sql.DB, tmplDir string) (string, error) {
	rows, err := db.Query(`
		SELECT creator_name,
		       CAST(EXTRACT(YEAR FROM MIN(date)) AS INTEGER) AS debut_year
		FROM crosswords
		GROUP BY creator_name
		ORDER BY debut_year, creator_name
	`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	// debut year -> sorted list of setter names
	debutNames := make(map[int][]string)
	for rows.Next() {
		var name string
		var year int
		if err := rows.Scan(&name, &year); err != nil {
			return "", err
		}
		debutNames[year] = append(debutNames[year], name)
	}

	// Build a contiguous year range
	currentYear := time.Now().Year()
	minYear := currentYear
	for y := range debutNames {
		if y < minYear {
			minYear = y
		}
	}

	type yearEntry struct {
		Year    int
		Count   int
		Setters []string
	}
	var results []yearEntry
	for y := minYear; y <= currentYear; y++ {
		names := debutNames[y]
		results = append(results, yearEntry{
			Year:    y,
			Count:   len(names),
			Setters: names,
		})
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Year < results[j].Year })

	labels := make([]int, len(results))
	values := []any{"New setters"}
	for i, r := range results {
		labels[i] = r.Year
		values = append(values, r.Count)
	}

	chartDef := map[string]any{
		"bindto": "#mychart11",
		"size":   map[string]any{"height": 400},
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
				"label": "New setters making their debut",
				"min":   0,
			},
		},
	}

	// Build the year->names lookup for the JS tooltip.
	// Keyed by the x-axis index (0-based position in the year range) so the
	// tooltip callback can look up names by d[0].x directly.
	type debutEntry struct {
		Year    int      `json:"year"`
		Setters []string `json:"setters"`
	}
	debutIndex := make([]debutEntry, len(results))
	for i, r := range results {
		setters := r.Setters
		if setters == nil {
			setters = []string{}
		}
		debutIndex[i] = debutEntry{Year: r.Year, Setters: setters}
	}

	data := map[string]any{
		"Title":        "New setters making their debut, per year",
		"Preamble":     "How many distinct setters made their Guardian crossword debut each year.  Hover over a bar to see who debuted that year.  Shows how the pool of contributors has grown (or shrunk) over time.",
		"Order":        11,
		"DivID":        "mychart11",
		"JSVar":        "chart11",
		"DefaultChart": "bar",
		"ChartJSON":    toJSON(chartDef),
		"DebutIndex":   toJSON(debutIndex),
		"BarColorJS":   barColorJS("chart11"),
	}
	return executeTemplate(tmplDir, "chart11.tmpl", data)
}
