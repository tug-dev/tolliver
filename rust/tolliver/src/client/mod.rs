use std::{
	io::{Read, Write},
	net::{self, TcpStream},
};

use crate::{
	structs::{
		handshake::{HandshakeCode, HandshakeError},
		tolliver_connection::TolliverConnection,
	},
	MessageType, MessageTypeNumber, VersionType, API_KEY_LENGTH, HANDSHAKE_CODE_LENGTH,
	MESSAGE_TYPE_LENGTH, VERSION, VERSION_LENGTH,
};

pub fn connect<A>(
	addr: A,
	api_key: [u8; API_KEY_LENGTH],
) -> Result<TolliverConnection, HandshakeError>
where
	A: net::ToSocketAddrs,
{
	let mut stream = TcpStream::connect(addr)?;

	send_handshake_request(api_key, &mut stream)?;

	let mut handshake_code_buf = [0; HANDSHAKE_CODE_LENGTH];
	stream.read_exact(&mut handshake_code_buf)?;
	let mut version_buf = [0; VERSION_LENGTH];
	stream.read_exact(&mut version_buf)?;
	let version = VersionType::from_be_bytes(version_buf);

	match handshake_code_buf {
		[0] => Ok(TolliverConnection::new(stream)?),
		[code] => Err(HandshakeError::Result(HandshakeCode::from_status_code(
			code, version,
		))),
	}
}

fn send_handshake_request(api_key: [u8; 32], stream: &mut TcpStream) -> Result<(), HandshakeError> {
	let total_length = MESSAGE_TYPE_LENGTH + VERSION_LENGTH + API_KEY_LENGTH;
	let mut buf = Vec::with_capacity(total_length);

	let message_type = MessageType::HandshakeRequest as MessageTypeNumber;
	let message_type_bytes = message_type.to_be_bytes();
	debug_assert_eq!(message_type_bytes.len(), MESSAGE_TYPE_LENGTH);

	let version_bytes = VERSION.to_be_bytes();
	debug_assert_eq!(version_bytes.len(), VERSION_LENGTH);

	buf.extend(message_type_bytes);
	buf.extend(version_bytes);
	buf.extend(api_key);
	stream.write_all(&buf)?;
	Ok(())
}
