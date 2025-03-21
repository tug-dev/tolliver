use std::{
	io::{self, Read, Write},
	net::TcpStream,
};

pub struct TolliverConnection {
	pub stream: TcpStream,
}

impl TolliverConnection {
	pub fn on_message(f: fn(&[u8])) {}

	pub fn read(&mut self) -> io::Result<[u8; 6]> {
		let mut buf = [0; 6];
		self.stream.read(&mut buf)?;
		Ok(buf)
	}

	pub fn fast_send(&mut self, buf: &[u8]) -> io::Result<()> {
		self.stream.write_all(buf)
	}
}
