use std::{
	fmt::Debug,
	io::{self, Read, Write},
	net::TcpStream,
};

use prost::Message;

use crate::error::TolliverError;

type BodyLengthType = u16;

/// The number of bytes the body length is encoded in
const BODY_LENGTH_LENGTH: usize = 2;
const DB_PATH: &str = "tolliver.db";

pub struct TolliverConnection {
	pub stream: TcpStream,
	db: rusqlite::Connection,
}

impl TolliverConnection {
	pub fn new(stream: TcpStream) -> Result<Self, TolliverError> {
		let db = rusqlite::Connection::open(DB_PATH)?;
		Self::init_db(&db)?;
		let mut conn = Self { stream, db };
		for message in conn.read_from_disk()? {
			conn.complete_send(message)?;
		}
		Ok(conn)
	}

	fn init_db(db: &rusqlite::Connection) -> rusqlite::Result<()> {
		let rows_updated = db.execute(
			"
CREATE TABLE IF NOT EXISTS message (
	id     INTEGER PRIMARY KEY,
	target TEXT,
	data   BLOB
);
PRAGMA journal_mode=WAL;",
			(),
		)?;
		debug_assert!(rows_updated <= 1);
		Ok(())
	}

	/// Receive one message from the connection
	pub fn read<T>(&mut self) -> io::Result<T>
	where
		T: Message,
		T: Default + Debug + Send + Sync,
	{
		let mut body_length_buf = [0; BODY_LENGTH_LENGTH];
		self.stream.read_exact(&mut body_length_buf)?;
		let body_length = BodyLengthType::from_be_bytes(body_length_buf);
		let mut body_buf = vec![0; body_length.into()];
		self.stream.read_exact(&mut body_buf)?;
		let message = Message::decode(&mut &body_buf[..]).unwrap();
		Ok(message)
	}

	/// Sends a fast message with no deliverability guarantees. This attempts to
	/// return as early as possible when it fails so something else can be tried.
	pub fn fast_send(&mut self, object: &impl Message) -> Result<(), TolliverError> {
		// See protocol documentation for details
		let body_length: u16 = match object.encoded_len().try_into() {
			Ok(r) => r,
			Err(_) => {
				return Err(TolliverError::TolliverError(
					"Could not encode length into u16, most likely object is too large".to_string(),
				))
			}
		};

		let total_length = BODY_LENGTH_LENGTH + body_length as usize;
		let mut buf = Vec::with_capacity(total_length);

		let body_length_bytes = body_length.to_be_bytes();
		debug_assert_eq!(body_length_bytes.len(), BODY_LENGTH_LENGTH);
		buf.extend(body_length_bytes);

		let mut body_buf = Vec::with_capacity(body_length.into());
		// Unwrap is safe, since we have reserved sufficient capacity in the vector.
		object.encode(&mut body_buf).unwrap();
		buf.extend(body_buf);
		self.stream.write_all(&buf)?;
		Ok(())
	}

	/// Sends a durable message to the peer. This returns when the message has
	/// been written to disk and is guaranteed to be delivered at some point.
	pub fn send(&mut self, object: &impl Message) -> Result<(), TolliverError> {
		// See protocol documentation for details
		let body_length: u16 = match object.encoded_len().try_into() {
			Ok(r) => r,
			Err(_) => {
				return Err(TolliverError::TolliverError(
					"Could not encode length into u16, most likely object is too large".to_string(),
				))
			}
		};

		let mut body_buf = Vec::with_capacity(body_length.into());
		// Unwrap is safe, since we have reserved sufficient capacity in the vector.
		object.encode(&mut body_buf).unwrap();

		let peer = self.stream.peer_addr()?;

		let total_length = BODY_LENGTH_LENGTH + body_length as usize;
		let mut buf = Vec::with_capacity(total_length);

		let body_length_bytes = body_length.to_be_bytes();
		debug_assert_eq!(body_length_bytes.len(), BODY_LENGTH_LENGTH);
		buf.extend(body_length_bytes);

		buf.extend(body_buf);
		let id = self.save_to_disk(peer.to_string(), &buf)?;
		let message = UnsentMessage {
			id,
			peer: peer.to_string(),
			message: buf,
		};
		self.complete_send(message)?;
		Ok(())
	}

	fn complete_send(&mut self, message: UnsentMessage) -> Result<(), TolliverError> {
		self.stream.write_all(&message.message)?;
		self.delete_from_disk(message.id)?;
		Ok(())
	}

	fn delete_from_disk(&mut self, message_id: i32) -> rusqlite::Result<()> {
		let rows_affected = self
			.db
			.execute("DELETE FROM message WHERE id = ?1", (message_id,))?;
		debug_assert_eq!(rows_affected, 1);
		Ok(())
	}

	/// Saves the message to disk and returns the saved ID.
	///
	/// # Errors
	///
	/// This function will return an error if the SQL execution fails.
	fn save_to_disk(&mut self, peer: String, body_buf: &Vec<u8>) -> rusqlite::Result<i32> {
		let id: i32 = self.db.query_row(
			"INSERT INTO message (target, data) VALUES (?1, ?2) RETURNING id",
			(peer, body_buf),
			|r| r.get(0),
		)?;
		Ok(id)
	}

	/// Returns a vector of unsent messages.
	///
	/// # Errors
	///
	/// This function will return an error if the SQL execution fails.
	fn read_from_disk(&mut self) -> rusqlite::Result<Vec<UnsentMessage>> {
		let mut stmt = self.db.prepare("SELECT id, target, data FROM message")?;
		let body_bufs = stmt.query_map([], |r| {
			Ok(UnsentMessage {
				id: r.get(0)?,
				peer: r.get(1)?,
				message: r.get(2)?,
			})
		})?;
		body_bufs.collect()
	}
}

/// This represents a message that has been saved to disk and not yet sent.
#[derive(Debug, PartialEq)]
struct UnsentMessage {
	id: i32,
	peer: String,
	message: Vec<u8>,
}

#[cfg(test)]
mod tests {

	use std::net::{TcpListener, TcpStream};

	use crate::structs::tolliver_connection::UnsentMessage;

	use super::TolliverConnection;

	fn setup_conn() -> TolliverConnection {
		let db = rusqlite::Connection::open_in_memory().unwrap();
		TolliverConnection::init_db(&db).unwrap();
		let listener = TcpListener::bind("127.0.0.1:0").unwrap();
		let listener_addr = listener.local_addr().unwrap();
		let stream = TcpStream::connect(listener_addr).unwrap();

		TolliverConnection { stream, db }
	}

	#[test]
	fn single_read_write() {
		let mut conn = setup_conn();

		// Just some random bytes
		let body_buf: Vec<u8> = vec![0, 8, 255, 42];
		// Documentation address as per https://www.rfc-editor.org/rfc/rfc5737#section-3
		let peer = "192.0.2.0:443";
		conn.save_to_disk(peer.to_string(), &body_buf).unwrap();

		let expected_body_bufs = vec![UnsentMessage {
			id: 1,
			peer: peer.to_string(),
			message: body_buf,
		}];
		let actual_body_bufs = conn.read_from_disk().unwrap();
		assert_eq!(expected_body_bufs, actual_body_bufs);
	}

	#[test]
	fn multi_read_write() {
		let mut conn = setup_conn();

		// Just some random bytes
		let body_buf: Vec<u8> = vec![0, 8, 255, 42];
		let body_buf2: Vec<u8> = vec![99, 98, 97, 2, 1, 0];
		// Documentation address as per https://www.rfc-editor.org/rfc/rfc5737#section-3
		let peer = "192.0.2.0:443";
		conn.save_to_disk(peer.to_string(), &body_buf).unwrap();
		conn.save_to_disk(peer.to_string(), &body_buf2).unwrap();

		let expected_body_bufs = vec![
			UnsentMessage {
				id: 1,
				peer: peer.to_string(),
				message: body_buf,
			},
			UnsentMessage {
				id: 2,
				peer: peer.to_string(),
				message: body_buf2,
			},
		];
		let actual_body_bufs = conn.read_from_disk().unwrap();
		assert_eq!(expected_body_bufs, actual_body_bufs);
	}

	#[test]
	fn empty_read_write() {
		let mut conn = setup_conn();

		// Just some random bytes
		let body_buf: Vec<u8> = vec![];
		// Documentation address as per https://www.rfc-editor.org/rfc/rfc5737#section-3
		let peer = "192.0.2.0:443";
		conn.save_to_disk(peer.to_string(), &body_buf).unwrap();

		let expected_body_bufs = vec![UnsentMessage {
			id: 1,
			peer: peer.to_string(),
			message: body_buf,
		}];
		let actual_body_bufs = conn.read_from_disk().unwrap();
		assert_eq!(expected_body_bufs, actual_body_bufs);
	}
}
