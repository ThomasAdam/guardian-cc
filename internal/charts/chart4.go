package charts

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// Chart4 generates "Setter Biographies" with per-setter area charts.
type Chart4 struct{}

func (c *Chart4) Order() string { return "4" }

type setterInfo struct {
	FirstDate    string
	LastDate     string
	Duration     string
	TotalAll     int
	TotalCryptic int
	TotalPrize   int
	SelfRefCount int
	// year -> count (aggregated across types)
	YearCounts map[int]int
	// For per-year graph data: slice of {year, count, type, pmonth}
	GraphData []graphEntry
}

type graphEntry struct {
	Year   int
	Count  int
	Type   string
	PMonth int
}

func (c *Chart4) Render(db *sql.DB, tmplDir string) (string, error) {
	setters := make(map[string]*setterInfo)

	// 1. Date ranges per setter
	rangeRows, err := db.Query(`
		SELECT creator_name AS name,
		       CAST(MIN(date) AS VARCHAR) AS first_date,
		       CAST(MAX(date) AS VARCHAR) AS last_date
		FROM crosswords
		GROUP BY creator_name
		ORDER BY creator_name
	`)
	if err != nil {
		return "", err
	}
	defer rangeRows.Close()

	for rangeRows.Next() {
		var name, firstDate, lastDate string
		if err := rangeRows.Scan(&name, &firstDate, &lastDate); err != nil {
			return "", err
		}
		// Trim to just date portion (YYYY-MM-DD)
		firstDate = strings.SplitN(firstDate, " ", 2)[0]
		lastDate = strings.SplitN(lastDate, " ", 2)[0]

		ft, _ := time.Parse("2006-01-02", firstDate)
		lt, _ := time.Parse("2006-01-02", lastDate)

		setters[name] = &setterInfo{
			FirstDate:  firstDate,
			LastDate:   lastDate,
			Duration:   formatDuration(ft, lt),
			YearCounts: make(map[int]int),
		}
	}

	// 2. Self-reference counts
	selfRows, err := db.Query(`
		SELECT c.creator_name AS name,
		       COUNT(*) AS count
		FROM entries e
		JOIN crosswords c ON e.crossword_id = c.id
		WHERE POSITION(c.creator_name IN e.clue) > 0
		GROUP BY c.creator_name
	`)
	if err != nil {
		return "", err
	}
	defer selfRows.Close()

	for selfRows.Next() {
		var name string
		var count int
		if err := selfRows.Scan(&name, &count); err != nil {
			return "", err
		}
		if s, ok := setters[name]; ok {
			s.SelfRefCount = count
		}
	}

	// 3. Crosswords per year per type per setter
	graphRows, err := db.Query(`
		SELECT creator_name AS name,
		       CAST(EXTRACT(YEAR FROM date) AS INTEGER) AS year,
		       crossword_type AS type,
		       COUNT(*) AS count,
		       CAST(CEIL(CAST(COUNT(*) AS DOUBLE) / 12) AS INTEGER) AS pmonth
		FROM crosswords
		GROUP BY creator_name, EXTRACT(YEAR FROM date), crossword_type
		ORDER BY creator_name, year, crossword_type
	`)
	if err != nil {
		return "", err
	}
	defer graphRows.Close()

	for graphRows.Next() {
		var name, ctype string
		var year, count, pmonth int
		if err := graphRows.Scan(&name, &year, &ctype, &count, &pmonth); err != nil {
			return "", err
		}
		s, ok := setters[name]
		if !ok {
			continue
		}
		s.GraphData = append(s.GraphData, graphEntry{Year: year, Count: count, Type: ctype, PMonth: pmonth})
		s.YearCounts[year] += count
		s.TotalAll += count
		if ctype == "cryptic" {
			s.TotalCryptic += count
		} else if ctype == "prize" {
			s.TotalPrize += count
		}
	}

	// Build per-setter chart definitions
	type setterChart struct {
		DivID        string
		JSVar        string
		Person       string
		FirstDate    string
		LastDate     string
		Duration     string
		TotalAll     int
		TotalCryptic int
		TotalPrize   int
		SelfRef      int
		ChartDef     string
	}

	names := make([]string, 0, len(setters))
	for k := range setters {
		names = append(names, k)
	}
	sort.Strings(names)

	var chartsData []setterChart
	for i, name := range names {
		s := setters[name]

		// Get sorted year labels
		years := make([]int, 0, len(s.YearCounts))
		for y := range s.YearCounts {
			years = append(years, y)
		}
		sort.Ints(years)

		yearLabels := make([]any, len(years))
		valuesRow := []any{name}
		avgRow := []any{"Average per month"}
		for j, y := range years {
			yearLabels[j] = y
			cnt := s.YearCounts[y]
			valuesRow = append(valuesRow, cnt)
			avgRow = append(avgRow, int(math.Ceil(float64(cnt)/12)))
		}

		chartDef := map[string]any{
			"bindto": fmt.Sprintf("#mychart4%d", i),
			"size":   map[string]any{"height": 200, "width": 600},
			"data": map[string]any{
				"columns": []any{valuesRow, avgRow},
				"type":    "area",
			},
			"axis": map[string]any{
				"x": map[string]any{
					"type":       "category",
					"tick":       map[string]any{"rotate": "75", "multiline": false},
					"height":     0,
					"categories": yearLabels,
				},
				"y": map[string]any{
					"label": "Number of crosswords",
					"tick":  map[string]any{"steps": 1},
					"min":   1,
				},
			},
			"legend": map[string]any{"show": false},
		}

		chartsData = append(chartsData, setterChart{
			DivID:        fmt.Sprintf("mychart4%d", i),
			JSVar:        fmt.Sprintf("chart4%d", i),
			Person:       name,
			FirstDate:    s.FirstDate,
			LastDate:     s.LastDate,
			Duration:     s.Duration,
			TotalAll:     s.TotalAll,
			TotalCryptic: s.TotalCryptic,
			TotalPrize:   s.TotalPrize,
			SelfRef:      s.SelfRefCount,
			ChartDef:     toJSON(chartDef),
		})
	}

	data := map[string]any{
		"Title":        "Setter Biographies",
		"Preamble":     "This shows information about each setter",
		"Order":        4,
		"DefaultChart": "area",
		"Charts":       chartsData,
	}
	return executeTemplate(tmplDir, "chart4.tmpl", data)
}

// formatDuration computes a human-readable duration like "5 years, 3 months, 12 days".
func formatDuration(from, to time.Time) string {
	years := to.Year() - from.Year()
	months := int(to.Month()) - int(from.Month())
	days := to.Day() - from.Day()

	if days < 0 {
		months--
		// Get days in previous month
		prev := to.AddDate(0, 0, -to.Day())
		days += prev.Day()
	}
	if months < 0 {
		years--
		months += 12
	}

	// Add 1 so the duration is inclusive of both the first and last day.
	// A setter whose first and last crossword share the same date was active
	// for 1 day, not 0.
	days++

	return fmt.Sprintf("%d years, %d months, %d days", years, months, days)
}
