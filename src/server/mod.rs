use crate::structs::incoming::Incoming;
use std::{
	io,
	net::{self, TcpListener},
};

pub struct TolliverServer {
	pub listener: TcpListener,
}

impl TolliverServer {
	/// Starts the Tolliver server at an avaliable port, to specify the port use `bind_at`
	///
	/// # Errors
	///
	/// This function will return an [`io::Error`] if the server cannot be started.
	pub fn bind() -> io::Result<Self> {
		Self::bind_at("0.0.0.0:0")
	}

	/// Starts the Tolliver server at a specific address, similar to `TcpListener`
	///
	/// # Errors
	///
	/// This function will return an [`io::Error`] if the server cannot be started.
	pub fn bind_at<A>(addr: A) -> io::Result<Self>
	where
		A: net::ToSocketAddrs,
	{
		let binded_data = Self {
			listener: TcpListener::bind(addr)?,
		};

		Ok(binded_data)
	}

	/// Returns an iterator over the connections being received on this
	/// server.
	///
	/// # Errors
	///
	/// Iterator only stops (returns [`None`]) if an error occurs.
	pub fn run(&self) -> Incoming<'_> {
		Incoming { listener: self }
	}
}
