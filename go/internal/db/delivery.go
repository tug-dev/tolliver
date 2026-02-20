package db

import (
	"database/sql"
	"log"

	"github.com/google/uuid"
)

type Delivery struct {
	Receiver uuid.UUID
	Payload  []byte
	MesId    uint64
	Channel  string
	Key      string
}

func GetWork(db *sql.DB) []Delivery {
	res, err := db.Query("SELECT (message_id, recipient_id) FROM delivery")
	if err != nil {
		panic(err)
	}

	out := make([]Delivery, 0, 10)
	for res.Next() {
		var mesId int
		var recipientId []byte
		var data []byte
		var channel, key string

		res.Scan(&mesId, &recipientId)
		recipientUUID, _ := uuid.FromBytes(recipientId)
		message, err := db.Query("SELECT channel, key, data FROM message WHERE id = $1", mesId)
		if err != nil || !message.Next() {
			log.Fatal("Message not found")
		}

		message.Scan(&channel, &key, &data)
		out = append(out, Delivery{Receiver: recipientUUID, Payload: data, MesId: uint64(mesId), Channel: channel, Key: key})
	}

	return out
}

func GetUndeliveredByUUID(db *sql.DB, id uuid.UUID) []Delivery {
	res, err := db.Query("SELECT (message_id, recipient_id) FROM delivery WHERE recipient_id = $1", id[:])
	if err != nil {
		panic(err)
	}

	out := make([]Delivery, 0, 10)
	for res.Next() {
		var mesId int
		var recipientId []byte
		var data []byte
		var channel, key string

		res.Scan(&mesId, &recipientId)
		recipientUUID, _ := uuid.FromBytes(recipientId)
		message, err := db.Query("SELECT channel, key, data FROM message WHERE id = $1", mesId)
		if err != nil || !message.Next() {
			log.Fatal("Message not found")
		}

		message.Scan(&channel, &key, &data)
		out = append(out, Delivery{Receiver: recipientUUID, Payload: data, MesId: uint64(mesId), Channel: channel, Key: key})
	}

	return out
}
