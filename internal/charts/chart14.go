package charts

import (
	"database/sql"
	"fmt"
)

// Chart14 shows the longest consecutive month-on-month streaks for each setter.
// A "month" is identified by YEAR * 100 + MONTH (e.g. 202403 = March 2024).
// A setter holds the streak for a given month if they published at least one
// crossword in that month.
// The chart displays each setter's longest-ever streak (in months), with a
// tooltip showing the start month, end month, and puzzle count during the run.
type Chart14 struct{}

func (c *Chart14) Order() string { return "14" }

func (c *Chart14) Render(db *sql.DB, tmplDir string) (string, error) {
	// Gaps-and-islands approach (same pattern as Chart13 but over calendar months):
	//  yr_mo = YEAR*100 + MONTH is a monotonically increasing integer where
	//  consecutive months differ by 1 (except the Dec→Jan boundary: 12→13, not 12→101).
	//  To handle the year boundary correctly we use a true month ordinal:
	//  (YEAR - min_year) * 12 + MONTH.  We compute this inside the CTE using a
	//  window-function-free trick: YEAR*12 + MONTH is already monotone and consecutive
	//  months always differ by exactly 1, so it works directly.
	rows, err := db.Query(`
		WITH monthly AS (
		    SELECT creator_name,
		           CAST(EXTRACT(YEAR  FROM date) AS INTEGER) * 12
		               + CAST(EXTRACT(MONTH FROM date) AS INTEGER) AS mo_ord,
		           STRFTIME(MIN(date), '%Y-%m')   AS month_label,
		           COUNT(*)                        AS puzzles_that_month
		    FROM crosswords
		    GROUP BY creator_name, mo_ord
		),
		numbered AS (
		    SELECT creator_name,
		           mo_ord,
		           month_label,
		           puzzles_that_month,
		           CAST(ROW_NUMBER() OVER (
		               PARTITION BY creator_name ORDER BY mo_ord
		           ) AS INTEGER) AS rn
		    FROM monthly
		),
		grouped AS (
		    SELECT creator_name,
		           mo_ord - rn                     AS grp,
		           MIN(month_label)                AS streak_start,
		           MAX(month_label)                AS streak_end,
		           COUNT(*)                        AS streak_months,
		           SUM(puzzles_that_month)         AS streak_puzzles
		    FROM numbered
		    GROUP BY creator_name, grp
		),
		best AS (
		    SELECT creator_name,
		           MAX(streak_months) AS best_months
		    FROM grouped
		    GROUP BY creator_name
		)
		SELECT g.creator_name,
		       g.streak_months,
		       g.streak_start,
		       g.streak_end,
		       g.streak_puzzles
		FROM grouped g
		JOIN best b
		  ON g.creator_name = b.creator_name
		 AND g.streak_months = b.best_months
		ORDER BY g.streak_months DESC, g.creator_name
	`)
	if err != nil {
		return "", fmt.Errorf("chart14 query: %w", err)
	}
	defer rows.Close()

	type streakEntry struct {
		Setter      string `json:"setter"`
		Months      int    `json:"months"`
		StreakStart string `json:"start"`
		StreakEnd   string `json:"end"`
		Puzzles     int    `json:"puzzles"`
	}

	seen := make(map[string]bool)
	var streaks []streakEntry
	for rows.Next() {
		var e streakEntry
		if err := rows.Scan(&e.Setter, &e.Months, &e.StreakStart, &e.StreakEnd, &e.Puzzles); err != nil {
			return "", fmt.Errorf("chart14 scan: %w", err)
		}
		if seen[e.Setter] {
			continue
		}
		seen[e.Setter] = true
		streaks = append(streaks, e)
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("chart14 rows: %w", err)
	}

	setterLabels := make([]string, len(streaks))
	monthValues := []any{"Consecutive months"}
	for i, s := range streaks {
		setterLabels[i] = s.Setter
		monthValues = append(monthValues, s.Months)
	}

	chartDef := map[string]any{
		"bindto": "#mychart14",
		"size":   map[string]any{"height": 500},
		"data": map[string]any{
			"columns": []any{monthValues},
			"type":    "bar",
		},
		"axis": map[string]any{
			"x": map[string]any{
				"type":       "category",
				"categories": setterLabels,
				"tick": map[string]any{
					"rotate":    75,
					"multiline": false,
				},
				"height": 130,
			},
			"y": map[string]any{
				"label": "Longest consecutive month streak",
				"min":   0,
			},
		},
	}

	tmplData := map[string]any{
		"Title":        "Longest consecutive month-on-month streaks per setter",
		"Preamble":     "For each setter, the longest run of consecutive calendar months in which they published at least one crossword.  Hover a bar to see when the streak started and ended, and how many puzzles were published during it.",
		"Order":        14,
		"DivID":        "mychart14",
		"JSVar":        "chart14",
		"DefaultChart": "bar",
		"ChartJSON":    toJSON(chartDef),
		"StreakIndex":  toJSON(streaks),
		"BarColorJS":   barColorJS("chart14"),
	}
	return executeTemplate(tmplDir, "chart14.tmpl", tmplData)
}
