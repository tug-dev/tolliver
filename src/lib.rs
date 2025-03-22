pub mod client;
pub mod server;
pub mod structs;

pub mod snazzy {
	pub mod items {
		include!(concat!(env!("OUT_DIR"), "/snazzy.items.rs"));
	}
}

#[cfg(test)]
mod tests {
	use core::panic;
	use std::thread;

	use server::TolliverServer;
	use snazzy::items;

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

	#[test]
	fn one_send() {
		let mut shirt = items::Shirt::default();
		shirt.color = "Red".to_string();
		shirt.set_size(items::shirt::Size::Large);
		let expected_shirt = shirt.clone();

		let server = TolliverServer::bind().unwrap();
		let incoming = server.run();
		let address = server.listener.local_addr().unwrap();
		thread::spawn(move || {
			let mut conn = client::connect(address).unwrap();
			conn.fast_send(shirt).unwrap();
		});
		for mut connection in incoming {
			let received_shirt: items::Shirt = connection.read().unwrap();
			assert_eq!(expected_shirt, received_shirt);
			return;
		}
		panic!("Incoming somehow ended")
	}
}
