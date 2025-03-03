use core::panic;
use std::{
	io,
	net::{self, TcpListener, TcpStream},
	sync::mpsc::{self, Receiver, Sender, TryRecvError},
	thread,
};

pub struct TolliverServer {
	status: ServerStatus,
	thread_stop_tx: Sender<()>,
}

enum ServerStatus {
	Binded(BindedServerData),
	Running(RunningServerData),
}

struct BindedServerData {
	listener: TcpListener,
	thread_stop_rx: Receiver<()>,
}

struct RunningServerData {
	join_handle: thread::JoinHandle<()>,
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
		let (thread_stop_tx, thread_stop_rx) = mpsc::channel();
		let binded_data = BindedServerData {
			listener: TcpListener::bind(addr)?,
			thread_stop_rx,
		};
		let server = Self {
			status: ServerStatus::Binded(binded_data),
			thread_stop_tx,
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
		let binded_data = match self.status {
			ServerStatus::Binded(data) => data,
			ServerStatus::Running(_) => return None,
		};
		let join_handle = thread::spawn(move || {
			// accept connections and process them serially
			for stream in binded_data.listener.incoming() {
				handle_client(stream.unwrap());
				match binded_data.thread_stop_rx.try_recv() {
					Ok(()) => return,
					Err(TryRecvError::Empty) => {}
					// Should never happen
					Err(TryRecvError::Disconnected) => {
						panic!("Thread shutdown sender disconnected")
					}
				}
			}
		});
		let server_data = RunningServerData { join_handle };
		self.status = ServerStatus::Running(server_data);
		Some(self)
	}

	pub fn send_stop(&self) -> Result<(), mpsc::SendError<()>> {
		self.thread_stop_tx.send(())
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
			ServerStatus::Running(data) => Some(data.join_handle.join()),
		}
	}
}
fn handle_client(stream: TcpStream) {
	println!("Got connection!")
}
