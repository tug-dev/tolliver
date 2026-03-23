use std::{
	io::{Read, Write},
	net::{self, TcpStream},
};

use uuid::Uuid;

use crate::{
	structs::{
		handshake::{HandshakeCode, HandshakeError},
		tolliver_connection::TolliverConnection,
	},
	MessageType, MessageTypeNumber, VersionType, HANDSHAKE_CODE_LENGTH, MESSAGE_TYPE_LENGTH,
	UUID_LENGTH, VERSION, VERSION_LENGTH,
};

pub fn connect<A>(addr: A, uuid: Uuid) -> Result<TolliverConnection, HandshakeError>
where
	A: net::ToSocketAddrs,
{
	let mut stream = TcpStream::connect(addr)?;

	send_handshake_request(uuid, &mut stream)?;
	get_handshake_response(stream)
}

fn send_handshake_request(uuid: Uuid, stream: &mut TcpStream) -> Result<(), HandshakeError> {
	let total_length = MESSAGE_TYPE_LENGTH + VERSION_LENGTH + UUID_LENGTH;
	let mut buf = Vec::with_capacity(total_length);

	let message_type = MessageType::HandshakeRequest as MessageTypeNumber;
	let message_type_bytes = message_type.to_be_bytes();
	debug_assert_eq!(message_type_bytes.len(), MESSAGE_TYPE_LENGTH);
	buf.extend(message_type_bytes);

	let version_bytes = VERSION.to_be_bytes();
	debug_assert_eq!(version_bytes.len(), VERSION_LENGTH);
	buf.extend(version_bytes);

	let uuid_bytes = uuid.into_bytes();
	debug_assert_eq!(uuid_bytes.len(), UUID_LENGTH);
	buf.extend(uuid_bytes);

	stream.write_all(&buf)?;
	Ok(())
}

fn get_handshake_response(mut stream: TcpStream) -> Result<TolliverConnection, HandshakeError> {
	let mut handshake_code_buf = [0; HANDSHAKE_CODE_LENGTH];
	stream.read_exact(&mut handshake_code_buf)?;
	let mut version_buf = [0; VERSION_LENGTH];
	stream.read_exact(&mut version_buf)?;
	let version = VersionType::from_be_bytes(version_buf);

	match handshake_code_buf {
		//TODO Make non-nil
		[0] => Ok(TolliverConnection::new(stream, Uuid::nil())?),
		[code] => Err(HandshakeError::Result(HandshakeCode::from_status_code(
			code, version,
		))),
	}
}
