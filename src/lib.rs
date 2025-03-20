pub mod client;
pub mod server;
pub mod structs;

#[cfg(test)]
mod tests {
	use core::panic;
	use std::thread;

	use super::*;

	#[test]
	fn start_server() {
		let server = server::TolliverServer::bind().unwrap();
		let incoming = server.run();
		thread::spawn(move || {
			client::TolliverClient::connect().unwrap();
		});
		for _connection in incoming {
			return;
		}
		panic!("Incoming somehow ended")
	}
}
