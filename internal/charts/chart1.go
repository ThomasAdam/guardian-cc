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

	labels := make([]string, 0, len(sorted))
	crypticData := make([]int, 0, len(sorted))
	prizeData := make([]int, 0, len(sorted))
	for _, s := range sorted {
		labels = append(labels, s.name)
		crypticData = append(crypticData, s.data.cryptic.count)
		prizeData = append(prizeData, s.data.prize.count)
	}

	chartDef := map[string]any{
		"type": "bar",
		"data": map[string]any{
			"labels": labels,
			"datasets": []any{
				map[string]any{
					"label": "Cryptic",
					"data":  crypticData,
					"stack": "total",
				},
				map[string]any{
					"label": "Prize",
					"data":  prizeData,
					"stack": "total",
				},
			},
		},
		"options": map[string]any{
			"responsive":          true,
			"maintainAspectRatio": false,
			"plugins": map[string]any{
				"legend": map[string]any{"display": true},
			},
			"scales": map[string]any{
				"x": map[string]any{
					"stacked": true,
					"ticks": map[string]any{
						"maxRotation": 75,
						"minRotation": 75,
					},
				},
				"y": map[string]any{
					"stacked": true,
					"title": map[string]any{
						"display": true,
						"text":    "No. of crosswords set",
					},
					"max": 800,
				},
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
		"Height":       1100,
		"ChartJSON":    toJSON(chartDef),
	}
	return executeTemplate(tmplDir, "chart.tmpl", data)
}
