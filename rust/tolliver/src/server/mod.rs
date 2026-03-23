use log::info;
use uuid::Uuid;

use crate::structs::incoming::Incoming;
use std::{
	io,
	net::{self, TcpListener},
};

pub struct TolliverServer {
	pub listener: TcpListener,
	pub uuid: Uuid,
}

impl TolliverServer {
	/// Starts the Tolliver server at an avaliable port, to specify the port use `bind_at`
	///
	/// # Errors
	///
	/// This function will return an [`io::Error`] if the server cannot be started.
	pub fn bind(uuid: Uuid) -> io::Result<Self> {
		Self::bind_at("0.0.0.0:0", uuid)
	}

	/// Starts the Tolliver server at a specific address, similar to `TcpListener`
	///
	/// # Errors
	///
	/// This function will return an [`io::Error`] if the server cannot be started.
	pub fn bind_at<A>(addr: A, uuid: Uuid) -> io::Result<Self>
	where
		A: net::ToSocketAddrs,
	{
		let binded_data = Self {
			listener: TcpListener::bind(addr)?,
			uuid,
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
		let addr = match self.listener.local_addr() {
			Ok(res) => res.to_string(),
			Err(_) => "unknown address".to_string(),
		};
		info!("Tolliver server started at {addr}");
		Incoming { listener: self }
	}
}
