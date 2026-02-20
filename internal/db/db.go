package db

import (
	"database/sql"
	"fmt"

	_ "github.com/marcboeker/go-duckdb"
)

const DefaultDBFile = "./guardian.duckdb"

// Open opens (or creates) a DuckDB database at the given path.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("duckdb", path)
	if err != nil {
		return nil, fmt.Errorf("opening duckdb %s: %w", path, err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging duckdb: %w", err)
	}
	return db, nil
}

// CreateSchema creates the tables if they don't already exist.
func CreateSchema(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS crosswords (
			id             VARCHAR PRIMARY KEY,
			number         VARCHAR,
			name           VARCHAR,
			creator_name   VARCHAR,
			creator_weburl VARCHAR,
			date           DATE,
			crossword_type VARCHAR,
			pdf            VARCHAR
		)`,
		`CREATE TABLE IF NOT EXISTS entries (
			crossword_id VARCHAR,
			entry_id     VARCHAR,
			number       INTEGER,
			human_number VARCHAR,
			clue         VARCHAR,
			direction    VARCHAR,
			length       INTEGER,
			solution     VARCHAR,
			pos_x        INTEGER,
			pos_y        INTEGER,
			FOREIGN KEY (crossword_id) REFERENCES crosswords(id)
		)`,
		// View that resolves "See N" / "See N across" / "See N (M)" style
		// cross-reference clues by looking up the target entry in the same
		// crossword.  When a bare "See N" matches both across and down, we
		// prefer the target whose clue is NOT itself a cross-reference.
		// Entries that are not cross-references pass through unchanged.
		`CREATE OR REPLACE VIEW resolved_entries AS
		WITH ref_parsed AS (
			SELECT e.*,
			       CASE WHEN e.clue ~ '^See\s+\d+'
			            THEN regexp_extract(e.clue, '^See\s+(\d+)', 1)
			            ELSE NULL
			       END AS target_num,
			       CASE WHEN e.clue ~ '^See\s+\d+\s+across'  THEN 'across'
			            WHEN e.clue ~ '^See\s+\d+\s+down'    THEN 'down'
			            ELSE NULL
			       END AS target_dir
			FROM entries e
		),
		resolved AS (
			SELECT r.crossword_id,
			       r.entry_id,
			       r.number,
			       r.human_number,
			       r.direction,
			       r.length,
			       r.solution,
			       r.pos_x,
			       r.pos_y,
			       r.clue AS original_clue,
			       r.target_num,
			       t.clue AS target_clue,
			       -- Prefer target whose clue is real (not itself a "See N")
			       ROW_NUMBER() OVER (
			           PARTITION BY r.crossword_id, r.entry_id
			           ORDER BY CASE WHEN t.clue NOT SIMILAR TO 'See\s+\d+.*' THEN 0 ELSE 1 END,
			                    t.direction
			       ) AS rn
			FROM ref_parsed r
			JOIN entries t
			  ON t.crossword_id = r.crossword_id
			  AND t.human_number = r.target_num
			  AND (r.target_dir IS NULL OR t.direction = r.target_dir)
			WHERE r.target_num IS NOT NULL
		)
		-- Resolved cross-references (pick best match)
		SELECT r.crossword_id,
		       r.entry_id,
		       r.number,
		       r.human_number,
		       COALESCE(r.target_clue, r.original_clue) AS clue,
		       r.direction,
		       r.length,
		       r.solution,
		       r.pos_x,
		       r.pos_y
		FROM resolved r
		WHERE r.rn = 1
		UNION ALL
		-- Non-cross-reference entries pass through unchanged
		SELECT e.crossword_id,
		       e.entry_id,
		       e.number,
		       e.human_number,
		       e.clue,
		       e.direction,
		       e.length,
		       e.solution,
		       e.pos_x,
		       e.pos_y
		FROM ref_parsed e
		WHERE e.target_num IS NULL`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("creating schema: %w", err)
		}
	}
	return nil
}
