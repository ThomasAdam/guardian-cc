package charts

import (
	"database/sql"
	"fmt"
)

// Chart13 shows the longest consecutive week-on-week streaks for each setter.
// A "week" is an ISO year-week (Monday-based). A setter holds the streak for
// a given week if they published at least one crossword in that week.
// The chart displays each setter's longest-ever streak (in weeks), with a
// tooltip showing the exact start date, end date, and how many crosswords
// appeared during the run.
type Chart13 struct{}

func (c *Chart13) Order() string { return "13" }

func (c *Chart13) Render(db *sql.DB, tmplDir string) (string, error) {
	// Gaps-and-islands approach:
	//  1. Collapse each (setter, week) to one row.
	//  2. Number each setter's weeks with ROW_NUMBER() in chronological order.
	//  3. week_ordinal - row_number is constant within a consecutive run, so
	//     GROUP BY (setter, grp) isolates each individual streak.
	//  4. Keep only the longest streak per setter.
	rows, err := db.Query(`
		WITH weekly AS (
		    SELECT creator_name,
		           CAST(EXTRACT(ISOYEAR FROM date) AS INTEGER) * 100
		               + CAST(EXTRACT(WEEK  FROM date) AS INTEGER) AS yr_wk,
		           MIN(date) AS week_start,
		           COUNT(*)  AS puzzles_that_week
		    FROM crosswords
		    GROUP BY creator_name, yr_wk
		),
		numbered AS (
		    SELECT creator_name,
		           yr_wk,
		           week_start,
		           puzzles_that_week,
		           CAST(ROW_NUMBER() OVER (
		               PARTITION BY creator_name ORDER BY yr_wk
		           ) AS INTEGER) AS rn
		    FROM weekly
		),
		grouped AS (
		    SELECT creator_name,
		           yr_wk - rn                     AS grp,
		           MIN(week_start)::VARCHAR        AS streak_start,
		           MAX(week_start)::VARCHAR        AS streak_end,
		           COUNT(*)                        AS streak_weeks,
		           SUM(puzzles_that_week)          AS streak_puzzles
		    FROM numbered
		    GROUP BY creator_name, grp
		),
		best AS (
		    SELECT creator_name,
		           MAX(streak_weeks) AS best_weeks
		    FROM grouped
		    GROUP BY creator_name
		)
		SELECT g.creator_name,
		       g.streak_weeks,
		       g.streak_start,
		       g.streak_end,
		       g.streak_puzzles
		FROM grouped g
		JOIN best b
		  ON g.creator_name = b.creator_name
		 AND g.streak_weeks  = b.best_weeks
		ORDER BY g.streak_weeks DESC, g.creator_name
	`)
	if err != nil {
		return "", fmt.Errorf("chart13 query: %w", err)
	}
	defer rows.Close()

	type streakEntry struct {
		Setter      string `json:"setter"`
		Weeks       int    `json:"weeks"`
		StreakStart string `json:"start"`
		StreakEnd   string `json:"end"`
		Puzzles     int    `json:"puzzles"`
	}

	// One entry per setter – if a setter has two equal-length best streaks we
	// take the first one returned (earliest chronologically via ORDER BY).
	seen := make(map[string]bool)
	var streaks []streakEntry
	for rows.Next() {
		var e streakEntry
		if err := rows.Scan(&e.Setter, &e.Weeks, &e.StreakStart, &e.StreakEnd, &e.Puzzles); err != nil {
			return "", fmt.Errorf("chart13 scan: %w", err)
		}
		if seen[e.Setter] {
			continue
		}
		seen[e.Setter] = true
		streaks = append(streaks, e)
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("chart13 rows: %w", err)
	}

	// Build C3 columns: x-axis = setter name, y = streak weeks.
	setterLabels := make([]string, len(streaks))
	weekValues := []any{"Consecutive weeks"}
	for i, s := range streaks {
		setterLabels[i] = s.Setter
		weekValues = append(weekValues, s.Weeks)
	}

	chartDef := map[string]any{
		"bindto": "#mychart13",
		"size":   map[string]any{"height": 500},
		"data": map[string]any{
			"columns": []any{weekValues},
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
				"label": "Longest consecutive week streak",
				"min":   0,
			},
		},
	}

	tmplData := map[string]any{
		"Title":        "Longest consecutive week-on-week streaks per setter",
		"Preamble":     "For each setter, the longest run of consecutive ISO weeks in which they published at least one crossword.  Hover a bar to see when the streak started and ended, and how many puzzles were published during it.",
		"Order":        13,
		"DivID":        "mychart13",
		"JSVar":        "chart13",
		"DefaultChart": "bar",
		"ChartJSON":    toJSON(chartDef),
		"StreakIndex":  toJSON(streaks),
		"BarColorJS":   barColorJS("chart13"),
	}
	return executeTemplate(tmplDir, "chart13.tmpl", tmplData)
}
