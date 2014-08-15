package main

import (
	"fmt"

	"database/sql"

	_ "github.com/carlsverre/go-libmysql/libmysql"
)

var (
	db *sql.DB
)

func mustExec(query string, args ...interface{}) sql.Result {
	res, err := db.Exec(query, args...)
	if err != nil {
		panic(err)
	}
	return res
}

func main() {
	var (
		err  error
		rows *sql.Rows
	)

	db, err = sql.Open("libmysql", "root@127.0.0.1:3306")
	if err != nil {
		panic(err)
	}

	mustExec("CREATE DATABASE IF NOT EXISTS foo")
	mustExec("USE foo")
	mustExec("DROP TABLE IF EXISTS bar")
	mustExec("CREATE TABLE IF NOT EXISTS bar (id int auto_increment primary key, name varchar(255))")

	for i := 0; i < 10; i++ {
		res := mustExec("INSERT INTO bar (name) VALUES (%s)", fmt.Sprintf("bob%d", i))
		insertId, _ := res.LastInsertId()
		fmt.Printf("Inserted row with id %d\n", insertId)
	}

	rows, err = db.Query("SELECT * FROM bar")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id   int
			name string
		)

		err := rows.Scan(&id, &name)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Person %s has id %d\n", name, id)
	}

	res := mustExec("DELETE FROM bar")
	deleted, _ := res.RowsAffected()
	fmt.Printf("Deleted %d rows\n", deleted)
}
