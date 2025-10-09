use std::{
	fmt::Debug,
	io::{self, Read, Write},
	net::TcpStream,
};

use prost::Message;

use crate::error::TolliverError;

use super::read_message::ReadMessage;

pub type BodyLengthType = u16;
pub type ProtoIdType = u64;

/// The number of bytes the body length is encoded in
const BODY_LENGTH_LENGTH: usize = 2;
/// The number of bytes the body length is encoded in
const PROTO_ID_LENGTH: usize = 8;
const DB_PATH: &str = "tolliver.db";

/// Compile time assertions
const _: () = {
	assert!(BodyLengthType::BITS == BODY_LENGTH_LENGTH as u32 * 8);
	assert!(ProtoIdType::BITS == PROTO_ID_LENGTH as u32 * 8);
};

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
		db.pragma_update(None, "journal_mode", &"WAL")?;
		let rows_updated = db.execute(
			"
CREATE TABLE IF NOT EXISTS message (
	id     INTEGER PRIMARY KEY,
	target TEXT,
	data   BLOB
)",
			(),
		)?;
		debug_assert!(rows_updated <= 1);
		Ok(())
	}

	/// Receive one message from the connection, returns a tuple containing the
	/// message and the numerical id to identify what proto message the body of
	/// the message was encoded with.
	///
	/// # Errors
	///
	/// This function will return an error only if there is a problem reading
	/// bytes from the stream.
	pub fn read(&mut self) -> io::Result<ReadMessage> {
		let mut proto_id_buf = [0; PROTO_ID_LENGTH];
		self.stream.read_exact(&mut proto_id_buf)?;
		let proto_id = ProtoIdType::from_be_bytes(proto_id_buf);

		let mut body_length_buf = [0; BODY_LENGTH_LENGTH];
		self.stream.read_exact(&mut body_length_buf)?;
		let body_length = BodyLengthType::from_be_bytes(body_length_buf);

		let mut body_buf = vec![0; body_length.into()];
		self.stream.read_exact(&mut body_buf)?;

		let message = ReadMessage {
			proto_id,
			body: body_buf,
		};
		Ok(message)
	}

	/// Sends a fast message with no deliverability guarantees. This attempts to
	/// return as early as possible when it fails so something else can be tried.
	pub fn unreliable_send(
		&mut self,
		proto_id: ProtoIdType,
		object: &impl Message,
	) -> Result<(), TolliverError> {
		let body_bytes = Self::message_to_bytes(object);
		self.unreliable_send_bytes(proto_id, body_bytes)
	}

	pub fn unreliable_send_bytes(
		&mut self,
		proto_id: ProtoIdType,
		bytes: Vec<u8>,
	) -> Result<(), TolliverError> {
		let message = Self::body_to_tolliver_message(proto_id, bytes)?;
		self.stream.write_all(&message)?;
		Ok(())
	}

	/// Sends a durable message to the peer. This returns when the message has
	/// been written to disk and is guaranteed to be delivered at some point.
	pub fn send(
		&mut self,
		proto_id: ProtoIdType,
		object: &impl Message,
	) -> Result<(), TolliverError> {
		let body_bytes = Self::message_to_bytes(object);
		self.send_bytes(proto_id, body_bytes)
	}

	pub fn send_bytes(
		&mut self,
		proto_id: ProtoIdType,
		bytes: Vec<u8>,
	) -> Result<(), TolliverError> {
		let peer = self.stream.peer_addr()?;
		let message_bytes = Self::body_to_tolliver_message(proto_id, bytes)?;
		let id = self.save_to_disk(peer.to_string(), &message_bytes)?;
		let unsent_message = UnsentMessage {
			id,
			peer: peer.to_string(),
			message_bytes,
		};
		self.complete_send(unsent_message)?;
		Ok(())
	}

	/// Given the bytes of the encoded ProtoBuf message, this functon adds the
	/// required data to turn it into a Tolliver message ready to be sent over TCP
	/// or written to disk.
	///
	/// For details, see the Tolliver protocol documentation.
	fn body_to_tolliver_message(
		proto_id: ProtoIdType,
		body: Vec<u8>,
	) -> Result<Vec<u8>, TolliverError> {
		// This length check can be done before encoding the ProtoBuf but we also
		// need to check the length if the *_bytes functions are used. The only
		// benefit of doing it before would be for it to return an error slightly
		// faster in the failing case but that would only happen if the ProtoBuf
		// was ~4.254 x 10^22 petabytes so very rarely, if ever, triggered.
		let body_length: BodyLengthType =
			match body.len().try_into() {
				Ok(r) => r,
				Err(_) => return Err(TolliverError::TolliverError(
					"Could not encode length into BodyLengthType, most likely object is too large"
						.to_string(),
				)),
			};

		let total_length = PROTO_ID_LENGTH + BODY_LENGTH_LENGTH + body_length as usize;
		let mut buf = Vec::with_capacity(total_length);

		buf.extend(proto_id.to_be_bytes());
		buf.extend(body_length.to_be_bytes());

		buf.extend(body);
		Ok(buf)
	}

	fn message_to_bytes(object: &impl Message) -> Vec<u8> {
		let mut body_bytes = Vec::with_capacity(object.encoded_len());
		// Unwrap is safe, since we have reserved sufficient capacity in the vector.
		object.encode(&mut body_bytes).unwrap();
		body_bytes
	}

	fn complete_send(&mut self, message: UnsentMessage) -> Result<(), TolliverError> {
		self.stream.write_all(&message.message_bytes)?;
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
				message_bytes: r.get(2)?,
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
	message_bytes: Vec<u8>,
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
			message_bytes: body_buf,
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
				message_bytes: body_buf,
			},
			UnsentMessage {
				id: 2,
				peer: peer.to_string(),
				message_bytes: body_buf2,
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
			message_bytes: body_buf,
		}];
		let actual_body_bufs = conn.read_from_disk().unwrap();
		assert_eq!(expected_body_bufs, actual_body_bufs);
	}
}
