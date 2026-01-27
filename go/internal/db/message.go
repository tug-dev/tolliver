package db

import (
	"database/sql"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

func SaveMessage(mes []byte, recipients []uuid.UUID, channel, key string, db *sql.DB) uint64 {
	res, err := db.Exec("INSERT INTO message (channel, key, data) VALUES ($1, $2, $3)", channel, key, mes)
	if err != nil {
		panic(err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}

	for _, v := range recipients {
		_, err := db.Exec("INSERT INTO delivery (message_id, recipient_id) VALUES ($1, $2)", id, v[:])
		if err != nil {
			panic(err)
		}
	}

	return uint64(id)
}
