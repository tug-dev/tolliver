use std::{io, net::TcpStream};

use crate::server::TolliverServer;

use super::tolliver_connection::TolliverConnection;

pub struct Incoming<'a> {
	pub listener: &'a TolliverServer,
}

impl<'a> Iterator for Incoming<'a> {
	type Item = TolliverConnection;
	fn next(&mut self) -> Option<Self::Item> {
		let stream = self.listener.listener.accept().map(|p| p.0);
		Some(tcp_to_tolliver_connection(stream))
	}
}

fn tcp_to_tolliver_connection(stream: io::Result<TcpStream>) -> TolliverConnection {
	TolliverConnection {
		stream: stream.unwrap(),
	}
}
