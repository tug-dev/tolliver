use std::{
	io::{self},
	net::{self, TcpStream},
};

use crate::structs::tolliver_connection::TolliverConnection;

pub fn connect<A>(addr: A) -> io::Result<TolliverConnection>
where
	A: net::ToSocketAddrs,
{
	let tcp_stream = TcpStream::connect(addr)?;
	Ok(TolliverConnection { stream: tcp_stream })
}
