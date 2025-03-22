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
			conn.fast_send(&shirt).unwrap();
		});
		for mut connection in incoming {
			assert_eq!(expected_shirt, connection.read().unwrap());
			return;
		}
		panic!("Incoming somehow ended")
	}

	#[test]
	fn multi_send() {
		let mut red_shirt = items::Shirt::default();
		red_shirt.color = "Red".to_string();
		red_shirt.set_size(items::shirt::Size::Large);
		let expected_red_shirt = red_shirt.clone();

		let mut blue_shirt = items::Shirt::default();
		blue_shirt.color = "Blue".to_string();
		blue_shirt.set_size(items::shirt::Size::Medium);
		let expected_blue_shirt = blue_shirt.clone();

		let server = TolliverServer::bind().unwrap();
		let incoming = server.run();
		let address = server.listener.local_addr().unwrap();
		thread::spawn(move || {
			let mut conn = client::connect(address).unwrap();
			conn.fast_send(&red_shirt).unwrap();
			conn.fast_send(&blue_shirt).unwrap();
			conn.fast_send(&red_shirt).unwrap();
			conn.fast_send(&red_shirt).unwrap();
			conn.fast_send(&blue_shirt).unwrap();
		});
		for mut connection in incoming {
			assert_eq!(expected_red_shirt, connection.read().unwrap());
			assert_eq!(expected_blue_shirt, connection.read().unwrap());
			assert_eq!(expected_red_shirt, connection.read().unwrap());
			assert_eq!(expected_red_shirt, connection.read().unwrap());
			assert_eq!(expected_blue_shirt, connection.read().unwrap());
			return;
		}
		panic!("Incoming somehow ended")
	}
}
