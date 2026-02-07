package db

import (
	"database/sql"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

func Ack(mesId uint32, recipientId uuid.UUID, db *sql.DB) {
	db.Exec("DELETE FROM delivery WHERE message_id=$1 AND recipient_id=$2", int64(mesId), recipientId[:])
}
