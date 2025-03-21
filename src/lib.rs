pub mod client;
pub mod server;
pub mod structs;

#[cfg(test)]
mod tests {
	use core::panic;
	use std::thread;

	use server::TolliverServer;

	use super::*;

	#[test]
	fn start_server() {
		let server = TolliverServer::bind().unwrap();
		let incoming = server.run();
		let address = server.listener.local_addr().unwrap();
		thread::spawn(move || {
			client::connect(address).unwrap();
		});
		for _connection in incoming {
			return;
		}
		panic!("Incoming somehow ended")
	}
}
