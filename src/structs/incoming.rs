use std::{
	io::{self, Read, Write},
	net::TcpStream,
};

use crate::{
	server::TolliverServer, VersionType, API_KEY_LENGTH, SERVER_RESPONSE_CODE_LENGTH, TEMP_API_KEY,
	VERSION, VERSION_LENGTH,
};

use super::tolliver_connection::TolliverConnection;

pub struct Incoming<'a> {
	pub listener: &'a TolliverServer,
}

impl<'a> Iterator for Incoming<'a> {
	type Item = TolliverConnection;
	fn next(&mut self) -> Option<Self::Item> {
		let stream = self.listener.listener.accept().map(|p| p.0);
		tcp_to_tolliver_connection(stream)
	}
}

fn tcp_to_tolliver_connection(stream: io::Result<TcpStream>) -> Option<TolliverConnection> {
	let mut stream = stream.unwrap();

	check_version(&mut stream)?;
	check_api_key(&mut stream)?;
	// TODO determine this code from an enum
	// TODO log error
	// Send success to client
	if write_response(&mut stream, 0).is_err() {
		return None;
	}

	Some(TolliverConnection { stream })
}

fn check_version(stream: &mut TcpStream) -> Option<()> {
	let mut version_buf = [0; VERSION_LENGTH];
	// TODO Log error
	let _res = stream.read_exact(&mut version_buf);
	let version = VersionType::from_be_bytes(version_buf);

	if version != VERSION {
		// TODO determine this code from an enum
		let code = 2;
		// Ignore result here because we're returning that the connection failed
		// anyway
		// TODO Log error
		let _res = write_response(stream, code);
		return None;
	}
	Some(())
}

fn check_api_key(stream: &mut TcpStream) -> Option<()> {
	let mut api_key = [0; API_KEY_LENGTH];
	// TODO Log error
	let _res = stream.read_exact(&mut api_key);

	if api_key != TEMP_API_KEY {
		// TODO determine this code from an enum
		let code = 3;
		// Ignore result here because we're returning that the connection failed
		// anyway
		// TODO Log error
		let _res = write_response(stream, code);
		return None;
	}
	Some(())
}

fn write_response(stream: &mut TcpStream, code: u8) -> io::Result<()> {
	let code_bytes = code.to_be_bytes();
	debug_assert_eq!(code_bytes.len(), SERVER_RESPONSE_CODE_LENGTH);
	stream.write_all(&code_bytes)?;

	let version_bytes = VERSION.to_be_bytes();
	debug_assert_eq!(version_bytes.len(), VERSION_LENGTH);
	stream.write_all(&version_bytes)?;
	Ok(())
}
