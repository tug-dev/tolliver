use std::{
	io::{self, Write},
	net::TcpStream,
};

pub struct TolliverClient {
	tcp_stream: TcpStream,
}

impl TolliverClient {
	pub fn connect() -> io::Result<Self> {
		let tcp_stream = TcpStream::connect("127.0.0.1:8080")?;
		Ok(Self { tcp_stream })
	}

	pub fn fast_send(&mut self, buf: &[u8]) -> io::Result<()> {
		self.tcp_stream.write_all(buf)
	}
}
