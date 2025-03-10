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
	copy(select ifnull(creator->>'name', 'unknown') as name,
	crosswordType,
	count (*) as num
	from cc
	group by all
	order by crosswordType, num desc, name) to 'run1.json' (ARRAY)`

	log.Print("Running setter count query...")
	db_run_stmt_transactional(db, ccSetterQuery)
	log.Print("    Done")

	/*
	   	type cc_count struct {
	   		Name          string `json:"name"`
	   		CrosswordType string `json:"crosswordType"`
	   		Count         int64  `json:"num"`
	   	}

	   rows, err := db.Query(ccSetterQuery)

	   	if err != nil {
	   		log.Fatal("Couldn't run ccSetterQuery")
	   	}

	   defer rows.Close()

	   var ccRes []cc_count

	   // Loop through rows, using Scan to assign column data to struct fields.

	   	for rows.Next() {
	   		var ccc cc_count
	   		if err := rows.Scan(&ccc.Name, &ccc.CrosswordType, &ccc.Count); err != nil {
	   			log.Fatal("Couldn't scan rows: ", err.Error())
	   		}
	   		ccRes = append(ccRes, ccc)
	   	}

	   	if err = rows.Err(); err != nil {
	   		log.Fatal("Rows error: ", err.Error())
	   	}

	   // Convert to JSON
	   jsonData, err := json.Marshal(ccRes)

	   	if err != nil {
	   		log.Fatal("Error converting to JSON:", err.Error())
	   	}

	   // Print the JSON as a string
	   fmt.Println(string(jsonData))
	*/
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
