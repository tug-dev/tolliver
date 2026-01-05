package db

import (
	"database/sql"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

func Ack(mesId uint32, recipientId uuid.UUID, dbPath string) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.Exec("DELETE FROM delivery WHERE message_id=$1 AND recipient_id=$2", int64(mesId), recipientId[:])
}
