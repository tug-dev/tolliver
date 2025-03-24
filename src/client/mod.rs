use std::{
	io::{Read, Write},
	net::{self, TcpStream},
};

use handshake_error::HandshakeError;

use crate::{
	structs::tolliver_connection::TolliverConnection, VersionType, API_KEY_LENGTH,
	SERVER_RESPONSE_CODE_LENGTH, VERSION, VERSION_LENGTH,
};

mod handshake_error;

pub fn connect<A>(
	addr: A,
	api_key: [u8; API_KEY_LENGTH],
) -> Result<TolliverConnection, HandshakeError>
where
	A: net::ToSocketAddrs,
{
	let mut stream = TcpStream::connect(addr)?;

	let total_length = VERSION_LENGTH + API_KEY_LENGTH;
	let mut buf = Vec::with_capacity(total_length);

	let version_bytes = VERSION.to_be_bytes();
	debug_assert_eq!(version_bytes.len(), VERSION_LENGTH);
	buf.extend(version_bytes);
	buf.extend(api_key);
	stream.write_all(&buf)?;

	let mut server_respose_code_buf = [0; SERVER_RESPONSE_CODE_LENGTH];
	stream.read_exact(&mut server_respose_code_buf)?;
	let mut version_buf = [0; VERSION_LENGTH];
	stream.read_exact(&mut version_buf)?;
	let version = VersionType::from_be_bytes(version_buf);

	return match server_respose_code_buf[0] {
		0 => Ok(TolliverConnection { stream }),
		1 => Err(HandshakeError::GeneralError),
		2 => Err(HandshakeError::IncompatibleVersion(version)),
		3 => Err(HandshakeError::Unauthorized),
		code => Err(HandshakeError::UnknowErrorCode(code)),
	};
	// if version != VERSION {
	// 	todo!("Version mismatch")
	// }
}
