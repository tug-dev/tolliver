package db

import (
	"database/sql"
	_ "embed"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

func Init(path string) uuid.UUID {
	// TODO: don't make multiple connections thats silly change this
	db, err := sql.Open("sqlite", path)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(schema)
	if err != nil {
		panic(err)
	}

	rows, err := db.Query("SELECT uuid FROM instance")
	if err != nil {
		panic(err)
	}

	var id uuid.UUID
	found := false
	defer rows.Close()
	for rows.Next() {
		var idBytes []byte
		if err := rows.Scan(&idBytes); err != nil {
			panic(err)
		}

		if idBytes != nil {
			id, _ = uuid.FromBytes(idBytes)
			found = true
			break
		}
	}

	if !found {
		id, _ = uuid.NewV7()
		db.Exec("INSERT INTO instance (uuid) VALUES ($1)", id[:])
	}

	return id
}
