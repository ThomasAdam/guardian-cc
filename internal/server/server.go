package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Serve starts an HTTP server on the given address.
// It serves static files from the current directory and handles
// DataTables server-side processing requests at /api/dt.
func Serve(db *sql.DB, addr string) error {
	mux := http.NewServeMux()

	// DataTables server-side processing endpoint
	mux.HandleFunc("/api/dt", func(w http.ResponseWriter, r *http.Request) {
		handleDataTable(db, w, r)
	})

	// Static files (gcc-analysis.html, ds_ajax*.txt, ui/, etc.)
	mux.Handle("/", http.FileServer(http.Dir(".")))

	log.Printf("Listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

// tableConfig defines the SQL and column mapping for each DataTable.
type tableConfig struct {
	// baseQuery is the FROM + JOIN + WHERE portion (no SELECT, no ORDER/LIMIT).
	baseQuery string
	// columns maps DataTables column index to SQL expression.
	columns []columnDef
}

type columnDef struct {
	// sql is the SQL expression for SELECT and ORDER BY.
	sql string
	// searchable indicates whether this column supports text search.
	searchable bool
}

var tables = map[string]tableConfig{
	"chart5": {
		baseQuery: `FROM (
			SELECT DISTINCT e.crossword_id, e.solution, e.clue,
			       c.creator_name, c.crossword_type,
			       c.id AS cw_path, c.number AS cw_number
			FROM resolved_entries e
			JOIN crosswords c ON e.crossword_id = c.id
			WHERE e.clue != ''
			  AND e.clue NOT SIMILAR TO '\s+\(\d+\)'
			  AND e.clue NOT SIMILAR TO 'See\s+\d+.*'
			  AND e.clue NOT SIMILAR TO 'See\s+(clues|special)\s+.*'
			  AND e.clue NOT SIMILAR TO 'Follow\s+the\s+link\s+below\s+to\s+see\s+today''s\s+clues.*'
			) d
			GROUP BY d.creator_name, d.solution
			HAVING COUNT(DISTINCT d.crossword_id) > 1`,
		columns: []columnDef{
			{sql: "d.creator_name", searchable: true},
			{sql: "d.solution", searchable: true},
			{sql: "STRING_AGG(d.clue, '<br />')", searchable: true},
			{sql: "STRING_AGG(d.crossword_type, '<br />')", searchable: false},
			{sql: "STRING_AGG('<a href=\"https://www.theguardian.com/' || d.cw_path || '\">' || d.cw_number || '</a>', '<br />')", searchable: false},
		},
	},
	"chart5a": {
		baseQuery: `FROM (
			SELECT DISTINCT e.crossword_id, e.clue, c.creator_name,
			       c.crossword_type, c.id AS cw_path, c.number AS cw_number
			FROM resolved_entries e
			JOIN crosswords c ON e.crossword_id = c.id
			WHERE e.clue != ''
			  AND e.clue NOT SIMILAR TO '\s+\(\d+\)'
			  AND e.clue NOT SIMILAR TO 'See\s+\d+.*'
			  AND e.clue NOT SIMILAR TO 'See\s+(clues|special)\s+.*'
			  AND e.clue NOT SIMILAR TO 'Follow\s+the\s+link\s+below\s+to\s+see\s+today''s\s+clues.*'
			) d
			GROUP BY d.creator_name, d.clue
			HAVING COUNT(DISTINCT d.crossword_id) > 1`,
		columns: []columnDef{
			{sql: "d.creator_name", searchable: true},
			{sql: "STRING_AGG(d.clue, '<br />')", searchable: true},
			{sql: "STRING_AGG(d.crossword_type, '<br />')", searchable: false},
			{sql: "STRING_AGG('<a href=\"https://www.theguardian.com/' || d.cw_path || '\">' || d.cw_number || '</a>', '<br />')", searchable: false},
		},
	},
	"chart6": {
		baseQuery: `FROM crosswords
			WHERE pdf IS NOT NULL`,
		columns: []columnDef{
			{sql: "creator_name", searchable: true},
			{sql: "'<a href=\"' || pdf || '\">' || number || '</a>'", searchable: true},
			{sql: "CAST(date AS VARCHAR)", searchable: true},
		},
	},
}

// dtResponse is the DataTables server-side processing response format.
type dtResponse struct {
	Draw            int        `json:"draw"`
	RecordsTotal    int        `json:"recordsTotal"`
	RecordsFiltered int        `json:"recordsFiltered"`
	Data            [][]string `json:"data"`
}

func handleDataTable(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	tableName := q.Get("table")
	tc, ok := tables[tableName]
	if !ok {
		http.Error(w, "unknown table", http.StatusBadRequest)
		return
	}

	draw, _ := strconv.Atoi(q.Get("draw"))
	start, _ := strconv.Atoi(q.Get("start"))
	length, _ := strconv.Atoi(q.Get("length"))
	if length <= 0 {
		length = 10
	}
	searchValue := q.Get("search[value]")

	// Build SELECT clause
	selectCols := make([]string, len(tc.columns))
	for i, col := range tc.columns {
		selectCols[i] = col.sql
	}
	selectClause := strings.Join(selectCols, ", ")

	// Build search WHERE clause (applied as a HAVING or wrapping subquery)
	var searchFilter string
	if searchValue != "" {
		escaped := strings.ReplaceAll(searchValue, "'", "''")
		var conditions []string
		for _, col := range tc.columns {
			if col.searchable {
				conditions = append(conditions,
					fmt.Sprintf("CAST(%s AS VARCHAR) ILIKE '%%%s%%'", col.sql, escaped))
			}
		}
		if len(conditions) > 0 {
			searchFilter = "(" + strings.Join(conditions, " OR ") + ")"
		}
	}

	// Build ORDER BY
	orderClause := ""
	orderCol := q.Get("order[0][column]")
	orderDir := q.Get("order[0][dir]")
	if orderCol != "" {
		colIdx, err := strconv.Atoi(orderCol)
		if err == nil && colIdx >= 0 && colIdx < len(tc.columns) {
			dir := "ASC"
			if strings.EqualFold(orderDir, "desc") {
				dir = "DESC"
			}
			orderClause = fmt.Sprintf("ORDER BY %d %s", colIdx+1, dir)
		}
	}

	// We wrap the grouped query as a subquery so we can apply search
	// filtering and pagination on top of it.
	innerQuery := fmt.Sprintf("SELECT %s %s", selectClause, tc.baseQuery)

	// Total count (unfiltered)
	totalSQL := fmt.Sprintf("SELECT COUNT(*) FROM (%s) _t", innerQuery)
	var totalCount int
	if err := db.QueryRow(totalSQL).Scan(&totalCount); err != nil {
		log.Printf("error counting total: %v", err)
		http.Error(w, "query error", http.StatusInternalServerError)
		return
	}

	// Filtered count + data query
	filteredCount := totalCount
	outerWhere := ""
	if searchFilter != "" {
		// Build column aliases for the subquery
		aliases := make([]string, len(tc.columns))
		for i := range tc.columns {
			aliases[i] = fmt.Sprintf("col%d", i)
		}
		aliasedCols := make([]string, len(tc.columns))
		for i, col := range tc.columns {
			aliasedCols[i] = fmt.Sprintf("%s AS %s", col.sql, aliases[i])
		}
		aliasedSelect := strings.Join(aliasedCols, ", ")
		aliasedInner := fmt.Sprintf("SELECT %s %s", aliasedSelect, tc.baseQuery)

		// Build search on aliases
		var searchConds []string
		for i, col := range tc.columns {
			if col.searchable {
				escaped := strings.ReplaceAll(searchValue, "'", "''")
				searchConds = append(searchConds,
					fmt.Sprintf("CAST(%s AS VARCHAR) ILIKE '%%%s%%'", aliases[i], escaped))
				_ = col
			}
		}
		outerWhere = "WHERE " + strings.Join(searchConds, " OR ")

		// Filtered count
		filteredSQL := fmt.Sprintf("SELECT COUNT(*) FROM (%s) _t %s", aliasedInner, outerWhere)
		if err := db.QueryRow(filteredSQL).Scan(&filteredCount); err != nil {
			log.Printf("error counting filtered: %v", err)
			http.Error(w, "query error", http.StatusInternalServerError)
			return
		}

		// Data query with aliases
		selectStar := strings.Join(aliases, ", ")
		dataSQL := fmt.Sprintf("SELECT %s FROM (%s) _t %s %s LIMIT %d OFFSET %d",
			selectStar, aliasedInner, outerWhere, orderClause, length, start)
		rows, err := db.Query(dataSQL)
		if err != nil {
			log.Printf("error querying data: %v (sql: %s)", err, dataSQL)
			http.Error(w, "query error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		resp := dtResponse{
			Draw:            draw,
			RecordsTotal:    totalCount,
			RecordsFiltered: filteredCount,
			Data:            make([][]string, 0),
		}
		numCols := len(tc.columns)
		for rows.Next() {
			row := make([]string, numCols)
			ptrs := make([]any, numCols)
			for i := range row {
				ptrs[i] = &row[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				log.Printf("error scanning: %v", err)
				continue
			}
			resp.Data = append(resp.Data, row)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// No search filter -- simpler query
	dataSQL := fmt.Sprintf("SELECT * FROM (%s) _t %s LIMIT %d OFFSET %d",
		innerQuery, orderClause, length, start)
	rows, err := db.Query(dataSQL)
	if err != nil {
		log.Printf("error querying data: %v (sql: %s)", err, dataSQL)
		http.Error(w, "query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	resp := dtResponse{
		Draw:            draw,
		RecordsTotal:    totalCount,
		RecordsFiltered: filteredCount,
		Data:            make([][]string, 0),
	}
	numCols := len(tc.columns)
	for rows.Next() {
		row := make([]string, numCols)
		ptrs := make([]any, numCols)
		for i := range row {
			ptrs[i] = &row[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			log.Printf("error scanning: %v", err)
			continue
		}
		resp.Data = append(resp.Data, row)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
