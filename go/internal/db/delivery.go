package db

import (
	"database/sql"
	"log"

	"github.com/google/uuid"
)

func GetWork(db *sql.DB) map[uuid.UUID][]byte {
	res, err := db.Query("SELECT (message_id, recipient_id) FROM delivery")
	if err != nil {
		panic(err)
	}

	out := make(map[uuid.UUID][]byte)
	for res.Next() {
		var mesId int
		var recipientId []byte
		var data []byte

		res.Scan(&mesId, &recipientId)
		recipientUUID, _ := uuid.FromBytes(recipientId)
		message, err := db.Query("SELECT data FROM message WHERE id = $1", mesId)
		if err != nil || !message.Next() {
			log.Fatal("Message not found")
		}

		message.Scan(&data)
		out[recipientUUID] = data
	}

	return out
}
