use std::{
	io::{self, Write},
	net::{self, TcpStream},
};

pub struct TolliverClient {
	tcp_stream: TcpStream,
}

impl TolliverClient {
	pub fn connect<A>(addr: A) -> io::Result<Self>
	where
		A: net::ToSocketAddrs,
	{
		let tcp_stream = TcpStream::connect(addr)?;
		Ok(Self { tcp_stream })
	}

	pub fn fast_send(&mut self, buf: &[u8]) -> io::Result<()> {
		self.tcp_stream.write_all(buf)
	}
}
