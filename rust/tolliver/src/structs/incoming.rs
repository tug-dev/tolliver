use std::{
	io::{self, IoSlice, IoSliceMut, Read, Write},
	net::TcpStream,
};

use log::warn;
use uuid::Uuid;

use crate::{
	server::TolliverServer,
	structs::handshake::{HandshakeFinalCode, HandshakeResponseCode},
	MessageType, MessageTypeNumber, StatusCode, VersionType, HANDSHAKE_CODE_LENGTH,
	MESSAGE_TYPE_LENGTH, STATUS_CODE_LENGTH, UUID_LENGTH, VERSION, VERSION_LENGTH,
};

use super::tolliver_connection::TolliverConnection;

/// This is the struct that actually does the processing of requests that the
/// server recieves. It implements [`Iterator`] over [`TolliverConnection`].
pub struct Incoming<'a> {
	pub listener: &'a TolliverServer,
}

impl<'a> Iterator for Incoming<'a> {
	type Item = TolliverConnection;

	fn next(&mut self) -> Option<Self::Item> {
		let stream = self.listener.listener.accept().map(|p| p.0);
		// TODO Right now, if a remote sends invalid data this function will return
		// `None` and the iterator will stop.
		tcp_to_tolliver_connection(stream, self.listener.uuid)
	}
}

fn tcp_to_tolliver_connection(
	stream: io::Result<TcpStream>,
	uuid: Uuid,
) -> Option<TolliverConnection> {
	let mut stream = stream.unwrap();

	let remote_uuid = match get_handshake_request(&mut stream) {
		Ok(res) => res,
		Err(code) => {
			send_handshake_response(&mut stream, uuid, code);
			return None;
		}
	};
	send_handshake_response(&mut stream, uuid, HandshakeResponseCode::Success);
	check_handshake_final(&mut stream)?;
	match TolliverConnection::new(stream, remote_uuid) {
		Ok(conn) => Some(conn),
		Err(e) => {
			warn!("Error creating TolliverConnection: {e}");
			None
		}
	}
}

fn get_handshake_request(stream: &mut TcpStream) -> Result<Uuid, HandshakeResponseCode> {
	check_message_type(stream)?;
	check_version(stream)?;
	get_remote_uuid(stream)
}

fn send_handshake_response(
	stream: &mut TcpStream,
	uuid: Uuid,
	response_code: HandshakeResponseCode,
) -> Option<()> {
	let response_code_num = response_code as StatusCode;
	match write_response(stream, uuid, response_code_num) {
		Ok(()) => Some(()),
		Err(e) => {
			warn!("Failed to send success to client: {e}");
			// We don't know how many bytes have been sent to the client so just
			// terminate the TCP connection.
			return None;
		}
	}
}

fn check_message_type(stream: &mut TcpStream) -> Result<(), HandshakeResponseCode> {
	let mut message_type_buf = [0; MESSAGE_TYPE_LENGTH];
	match stream.read_exact(&mut message_type_buf) {
		Ok(()) => {}
		Err(e) => {
			warn!("Handshake failed: could not read message type: {e}");
			return Err(HandshakeResponseCode::GeneralError);
		}
	};
	let message_type = MessageTypeNumber::from_be_bytes(message_type_buf);

	if message_type == (MessageType::HandshakeRequest as MessageTypeNumber) {
		Ok(())
	} else {
		return Err(HandshakeResponseCode::GeneralError);
	}
}

fn check_version(stream: &mut TcpStream) -> Result<(), HandshakeResponseCode> {
	let mut version_buf = [0; VERSION_LENGTH];
	match stream.read_exact(&mut version_buf) {
		Ok(()) => {}
		Err(e) => {
			warn!("Handshake failed: could not read version: {e}");
			return Err(HandshakeResponseCode::GeneralError);
		}
	};
	let version = VersionType::from_be_bytes(version_buf);

	if version == VERSION {
		Ok(())
	} else {
		Err(HandshakeResponseCode::NewerVersionWithoutCompatibility)
	}
}

fn get_remote_uuid(stream: &mut TcpStream) -> Result<Uuid, HandshakeResponseCode> {
	let mut uuid_bytes = [0; UUID_LENGTH];
	match stream.read_exact(&mut uuid_bytes) {
		Ok(()) => {}
		Err(e) => {
			warn!("Handshake failed: could not read remote UUID: {e}");
			return Err(HandshakeResponseCode::GeneralError);
		}
	};
	Ok(Uuid::from_bytes(uuid_bytes))
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

fn check_handshake_final(stream: &mut TcpStream) -> Option<()> {
	let mut message_type_buf = [0; MESSAGE_TYPE_LENGTH];
	let message_type_io_slice = IoSliceMut::new(&mut message_type_buf);

	let mut status_code_buf = [0; STATUS_CODE_LENGTH];
	let status_code_io_slice = IoSliceMut::new(&mut status_code_buf);

	stream
		.read_vectored(&mut [message_type_io_slice, status_code_io_slice])
		.ok()?;

	let message_type = MessageTypeNumber::from_be_bytes(message_type_buf);
	let status_code = StatusCode::from_be_bytes(status_code_buf);

	let handshake_final = MessageType::HandshakeFinal as MessageTypeNumber;
	let success_status = HandshakeFinalCode::Success as StatusCode;
	if message_type == handshake_final && status_code == success_status {
		Some(())
	} else {
		None
	}
}
