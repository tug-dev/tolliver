CREATE TABLE IF NOT EXISTS message (
	id     INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	channel TEXT NOT NULL,
    `key` TEXT NOT NULL,
	data   BLOB NOT NULL
);

-- Should be deleted after ack of delivery
CREATE TABLE IF NOT EXISTS delivery (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    message_id INTEGER NOT NULL,
    recipient_id BLOB NOT NULL,
    FOREIGN KEY(message_id) REFERENCES message(id)
);

-- Should only have a single row for this nodes UUID
CREATE TABLE IF NOT EXISTS instance (
    uuid BLOB PRIMARY KEY NOT NULL
);

CREATE TABLE IF NOT EXISTS subscription (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    channel TEXT,
    `key` TEXT,
    instance_id TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS subscription_delivery (
    subscription_id INTEGER NOT NULL,
    recipient_id BLOB NOT NULL,
    FOREIGN KEY(subscription_id) REFERENCES subscription(id)
);

CREATE TABLE IF NOT EXISTS unsubscription_delivery (
    subscription_id INTEGER NOT NULL,
    recipient_id BLOB NOT NULL,
    FOREIGN KEY(subscription_id) REFERENCES subscription(id)
);

CREATE INDEX IF NOT EXISTS subscription_instance_id_idx ON subscription (
    instance_id
);

CREATE INDEX IF NOT EXISTS subscription_key_idx ON subscription (
   `key`
);

CREATE INDEX IF NOT EXISTS subscription_channel_idx ON subscription (
    channel
);
