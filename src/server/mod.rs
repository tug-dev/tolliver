use std::{
	io,
	net::{self, TcpListener, TcpStream},
	thread,
};

enum ServerStatus {
	Binded(TcpListener),
	Running(thread::JoinHandle<()>),
}

pub struct TolliverServer {
	status: ServerStatus,
}

impl TolliverServer {
	pub fn bind() -> io::Result<Self> {
		Self::bind_at("0.0.0.0:8080")
	}

	/// Starts the Tolliver server at a specific address, similar to `TcpListener`
	///
	/// # Errors
	///
	/// This function will return an error if the server cannot be started.
	pub fn bind_at<A>(addr: A) -> io::Result<Self>
	where
		A: net::ToSocketAddrs,
	{
		let server = Self {
			status: ServerStatus::Binded(TcpListener::bind(addr)?),
		};

		Ok(server)
	}

	/// Starts a thread with the server on it.
	///
	/// # Errors
	///
	/// Returns `None` if already running.
	pub fn run(mut self) -> Option<Self> {
		// The listener is set in the constructor, so it would only be None if
		// `run` has already been called. In that case panic.
		let listener = match self.status {
			ServerStatus::Binded(listener) => listener,
			ServerStatus::Running(_) => return None,
		};
		let join_handle = thread::spawn(move || {
			// accept connections and process them serially
			for stream in listener.incoming() {
				handle_client(stream.unwrap());
			}
		});
		self.status = ServerStatus::Running(join_handle);
		Some(self)
	}

	/// Waits until the server shuts down.
	///
	/// # Errors
	///
	/// This function will return None if the server has not been started, and
	/// an error if there was an issue stopping the thread.
	pub fn wait(self) -> Option<thread::Result<()>> {
		match self.status {
			ServerStatus::Binded(_) => None,
			ServerStatus::Running(handle) => Some(handle.join()),
		}
	}
}
fn handle_client(stream: TcpStream) {
	println!("Got connection!")
}
