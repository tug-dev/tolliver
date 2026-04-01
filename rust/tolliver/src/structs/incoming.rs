use std::{
	io::{self, IoSlice, Read, Write},
	net::TcpStream,
};

use log::warn;
use uuid::Uuid;

use crate::{
	server::TolliverServer, MessageType, MessageTypeNumber, VersionType, HANDSHAKE_CODE_LENGTH,
	MESSAGE_TYPE_LENGTH, UUID_LENGTH, VERSION, VERSION_LENGTH,
};

use super::{handshake::HandshakeCode, tolliver_connection::TolliverConnection};

/// This is the struct that actually does the processing of requests that the
/// server recieves. It implements [`Iterator`] over [`TolliverConnection`].
pub struct Incoming<'a> {
	pub listener: &'a TolliverServer,
}

impl<'a> Iterator for Incoming<'a> {
	type Item = TolliverConnection;

	fn next(&mut self) -> Option<Self::Item> {
		let stream = self.listener.listener.accept().map(|p| p.0);
		tcp_to_tolliver_connection(stream, self.listener.uuid)
	}
}

fn tcp_to_tolliver_connection(
	stream: io::Result<TcpStream>,
	uuid: Uuid,
) -> Option<TolliverConnection> {
	let mut stream = stream.unwrap();

	let remote_uuid = get_handshake_request(&mut stream, uuid)?;
	// Send success to client
	send_handshake_response(stream, uuid, remote_uuid)
}

fn get_handshake_request(stream: &mut TcpStream, uuid: Uuid) -> Option<Uuid> {
	check_message_type(stream, uuid)?;
	check_version(stream, uuid)?;
	get_remote_uuid(stream)
}

fn send_handshake_response(
	mut stream: TcpStream,
	uuid: Uuid,
	remote_uuid: Uuid,
) -> Option<TolliverConnection> {
	let success_code = HandshakeCode::Success.status_code();
	match write_response(&mut stream, uuid, success_code) {
		Ok(()) => match TolliverConnection::new(stream, remote_uuid) {
			Ok(conn) => Some(conn),
			Err(e) => {
				warn!("Error creating TolliverConnection: {e}");
				None
			}
		},
		Err(e) => {
			warn!("Failed to send success to client: {e}");
			// We don't know how many bytes have been sent to the client so just
			// terminate the TCP connection.
			return None;
		}
	}
}

fn check_message_type(stream: &mut TcpStream, uuid: Uuid) -> Option<()> {
	let mut message_type_buf = [0; MESSAGE_TYPE_LENGTH];
	match stream.read_exact(&mut message_type_buf) {
		Ok(()) => {}
		Err(e) => {
			warn!("Handshake failed: could not read message type: {e}");
			return None;
		}
	};
	let message_type = MessageTypeNumber::from_be_bytes(message_type_buf);

	if message_type != (MessageType::HandshakeRequest as MessageTypeNumber) {
		let handshake_code = HandshakeCode::IncompatibleVersion(0).status_code();
		write_response(stream, uuid, handshake_code).unwrap_or_else(|e| {
			warn!("Handshake failed: could not send incompatible version error: {e}")
		});
		return None;
	}
	Some(())
}

fn check_version(stream: &mut TcpStream, uuid: Uuid) -> Option<()> {
	let mut version_buf = [0; VERSION_LENGTH];
	match stream.read_exact(&mut version_buf) {
		Ok(()) => {}
		Err(e) => {
			warn!("Handshake failed: could not read version: {e}");
			return None;
		}
	};
	let version = VersionType::from_be_bytes(version_buf);

	if version != VERSION {
		let handshake_code = HandshakeCode::IncompatibleVersion(0).status_code();
		write_response(stream, uuid, handshake_code).unwrap_or_else(|e| {
			warn!("Handshake failed: could not send incompatible version error: {e}")
		});
		return None;
	}
	Some(())
}

fn get_remote_uuid(stream: &mut TcpStream) -> Option<Uuid> {
	let mut uuid_bytes = [0; UUID_LENGTH];
	match stream.read_exact(&mut uuid_bytes) {
		Ok(()) => {}
		Err(e) => {
			warn!("Handshake failed: could not read remote UUID: {e}");
			return None;
		}
	};
	Some(Uuid::from_bytes(uuid_bytes))
}

fn write_response(stream: &mut TcpStream, uuid: Uuid, code: u8) -> io::Result<()> {
	let message_type = MessageType::HandshakeResponse as MessageTypeNumber;
	let message_type_bytes = message_type.to_be_bytes();
	debug_assert_eq!(message_type_bytes.len(), MESSAGE_TYPE_LENGTH);

	let version_bytes = VERSION.to_be_bytes();
	debug_assert_eq!(version_bytes.len(), VERSION_LENGTH);

	let uuid_bytes = uuid.into_bytes();
	debug_assert_eq!(uuid_bytes.len(), UUID_LENGTH);

	let handshake_code_bytes = code.to_be_bytes();
	debug_assert_eq!(handshake_code_bytes.len(), HANDSHAKE_CODE_LENGTH);

	stream.write_vectored(&[
		IoSlice::new(&message_type_bytes),
		IoSlice::new(&version_bytes),
		IoSlice::new(&uuid_bytes),
		IoSlice::new(&handshake_code_bytes),
	])?;
	Ok(())
}
