use std::{
	io::{self, IoSlice, IoSliceMut, Read, Write},
	net::{self, TcpStream},
};

use uuid::Uuid;

use crate::{
	error::TolliverError,
	structs::{
		handshake::{HandshakeError, HandshakeFinalCode},
		tolliver_connection::TolliverConnection,
	},
	MessageType, MessageTypeNumber, StatusCode, VersionType, HANDSHAKE_CODE_LENGTH,
	MESSAGE_TYPE_LENGTH, STATUS_CODE_LENGTH, UUID_LENGTH, VERSION, VERSION_LENGTH,
};

pub fn connect<A>(addr: A, uuid: Uuid) -> Result<TolliverConnection, HandshakeError>
where
	A: net::ToSocketAddrs,
{
	let mut stream = TcpStream::connect(addr)?;

	send_handshake_request(uuid, &mut stream)?;
	// Pass some errors to handshake final
	let (status_code, remote_uuid) = get_handshake_response(&mut stream)?;
	send_handshake_final(&mut stream, status_code)?;
	Ok(TolliverConnection::new(stream, remote_uuid)?)
}

fn send_handshake_request(uuid: Uuid, stream: &mut TcpStream) -> io::Result<()> {
	let message_type = MessageType::HandshakeRequest as MessageTypeNumber;
	let message_type_bytes = message_type.to_be_bytes();
	debug_assert_eq!(message_type_bytes.len(), MESSAGE_TYPE_LENGTH);

	let version_bytes = VERSION.to_be_bytes();
	debug_assert_eq!(version_bytes.len(), VERSION_LENGTH);

	let uuid_bytes = uuid.into_bytes();
	debug_assert_eq!(uuid_bytes.len(), UUID_LENGTH);

	stream.write_vectored(&[
		IoSlice::new(&message_type_bytes),
		IoSlice::new(&version_bytes),
		IoSlice::new(&uuid_bytes),
	])?;
	Ok(())
}

fn get_handshake_response(stream: &mut TcpStream) -> Result<(StatusCode, Uuid), HandshakeError> {
	let mut message_type_buf = [0; MESSAGE_TYPE_LENGTH];
	let message_type_io_slice = IoSliceMut::new(&mut message_type_buf);

	let mut version_buf = [0; VERSION_LENGTH];
	let version_io_slice = IoSliceMut::new(&mut version_buf);

	let mut uuid_buf = [0; UUID_LENGTH];
	let uuid_io_slice = IoSliceMut::new(&mut uuid_buf);

	let mut handshake_code_buf = [0; HANDSHAKE_CODE_LENGTH];
	let handshake_code_io_slice = IoSliceMut::new(&mut handshake_code_buf);

	stream.read_vectored(&mut [
		message_type_io_slice,
		version_io_slice,
		uuid_io_slice,
		handshake_code_io_slice,
	])?;
	let message_type_num = MessageTypeNumber::from_be_bytes(message_type_buf);
	let handshake_response_number = MessageType::HandshakeResponse as MessageTypeNumber;
	if message_type_num != handshake_response_number {
		//TODO Simplify these errors
		return Err(HandshakeError::TolliverError(TolliverError::TolliverError(
			"Remote did not reply to handshake request with handshake response".to_string(),
		)));
	}
	let version = VersionType::from_be_bytes(version_buf);
	let handshake_final_code = if version == VERSION {
		HandshakeFinalCode::Success
	} else {
		HandshakeFinalCode::IncompatibleVersion
	};
	let status_code = handshake_final_code as StatusCode;
	let uuid = Uuid::from_bytes(uuid_buf);

	match handshake_code_buf {
		[0] => Ok((status_code, uuid)),
		[code] => Err(HandshakeError::TolliverError(TolliverError::TolliverError(
			format!("Handshake failed, remote sent status code: {code}"),
		))),
	}
}

fn send_handshake_final(
	stream: &mut TcpStream,
	status_code: StatusCode,
) -> Result<(), HandshakeError> {
	let message_type = MessageType::HandshakeFinal as MessageTypeNumber;
	let message_type_bytes = message_type.to_be_bytes();
	debug_assert_eq!(message_type_bytes.len(), MESSAGE_TYPE_LENGTH);

	let status_code_bytes = status_code.to_be_bytes();
	debug_assert_eq!(status_code_bytes.len(), STATUS_CODE_LENGTH);

	stream.write_vectored(&[
		IoSlice::new(&message_type_bytes),
		IoSlice::new(&status_code_bytes),
	])?;
	Ok(())
}
