package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

func main() {
	dbName := "users.sqlite3"
	os.Remove(dbName)
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	createTable := `
	create table users (username text, password text);
	`
	_, err = db.Exec(createTable)
	if err != nil {
		panic(err)
	}
}
