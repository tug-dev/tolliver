package db

import (
	"database/sql"

	"github.com/google/uuid"
)

// TODO: check about sqlite enforcing uniqueness constraints and maybe use transaction

func GetSubscriberUUIDs(channel, key string, db *sql.DB) []uuid.UUID {
	res, err := db.Query("SELECT id FROM subscription WHERE (channel = $1 OR channel = \"\") AND (key = $2 OR key = \"\")")
	if err != nil {
		panic(err)
	}
	out := make([]uuid.UUID, 0, 10)
	for res.Next() {
		b := make([]byte, 16)
		res.Scan(&b)
		id, err := uuid.FromBytes(b)
		if err != nil {
			panic(err)
		}
		out = append(out, id)
	}

	return out
}

func Subscribe(channel, key string, id uuid.UUID, db *sql.DB) {
	_, err := db.Exec("INSERT INTO subscription (channel, key, id) VALUES ($1, $2, $3)", channel, key, id)
	if err != nil {
		panic(err)
	}
}

func Unsubscribe(channel, key string, id uuid.UUID, db *sql.DB) {
	_, err := db.Exec("DELETE FROM subscription WHERE channel = $1 AND key = $2 AND id = $3)", channel, key, id)
	if err != nil {
		panic(err)
	}
}
