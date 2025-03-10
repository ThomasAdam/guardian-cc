package main

import (
	"database/sql"
	"log"

	_ "github.com/gookit/goutil/dump"

	_ "github.com/marcboeker/go-duckdb/v2"
)

// FIXME: should be in its own object.
func db_run_stmt_transactional(db *sql.DB, s string) {

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(s)
	if err != nil {
		log.Fatalf("Couldn't run: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func db_create() *sql.DB {
	log.Println("Creating database...")
	db, err := sql.Open("duckdb", "cc.duck")
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func cc_get_setter_count(db *sql.DB) {
	const ccSetterQuery string = `
	COPY(SELECT
        crosswordType,
        JSON_GROUP_ARRAY(STRUCT_PACK(name, num)) AS puzzles
	FROM (
		SELECT
			IFNULL(creator->>'name', 'unknown') AS name,
			crosswordType,
			COUNT(*) AS num
		FROM cc
		GROUP BY ALL
		ORDER BY num desc
	)
	GROUP BY crosswordType
	ORDER BY crosswordType) to 'foo1.json';`

	log.Print("Running setter count query...")
	db_run_stmt_transactional(db, ccSetterQuery)
	log.Print("    Done")
}

func main() {
	var db *sql.DB

	db = db_create()
	defer db.Close()

	log.Println("Importing JSON files...")

	const importStr string = `
	SET preserve_insertion_order = false;
	CREATE OR REPLACE TABLE cc AS
	(SELECT * FROM read_json_auto('**/*.JSON', ignore_errors=true))`

	db_run_stmt_transactional(db, importStr)

	log.Print("Running queries...")
	cc_get_setter_count(db)
}
