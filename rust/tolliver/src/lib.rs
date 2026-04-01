pub mod client;
pub mod error;
pub mod server;
pub mod structs;

/// The "version type" is the type of the literal version number. It should
/// correspond with the number of bytes in the VERSION_LENGTH.
type VersionType = u64;
/// The "server status code type" is the type of the literal status code. It
/// should correspond with the number of bytes in SERVER_RESPONSE_CODE_LENGTH.
pub type HandshakeCodeType = u8;
pub type MessageTypeNumber = u8;

/// The version of the protocol
const VERSION: VersionType = 1;
/// The number of bytes the version number is encoded in
const VERSION_LENGTH: usize = 8;
/// The number of bytes the UUID is encoded in
const UUID_LENGTH: usize = 16;
/// The number of bytes the server success/error response is encoded in
const HANDSHAKE_CODE_LENGTH: usize = 1;
/// The number of bytes the message type is encoded in
const MESSAGE_TYPE_LENGTH: usize = 1;
/// The type of message that is being sent
pub enum MessageType {
	HandshakeRequest = 0,
	HandshakeResponse = 1,
	HandshakeFinal = 2,
}

/// The general status code type.
pub type StatusCode = u8;
/// The number of bytes a [StatusCode] is encoded in.
const STATUS_CODE_LENGTH: usize = 1;

#[cfg(test)]
mod tests {
	mod items {
		include!(concat!(env!("OUT_DIR"), "/snazzy.items.rs"));
	}

	use core::panic;
	use std::thread;

	use server::TolliverServer;
	use structs::tolliver_connection::ProtoIdType;
	use uuid::Uuid;

	use super::*;

	const EXAMPLE_PROTO_ID: ProtoIdType = 0;

	#[test]
	fn start_server() {
		let server = TolliverServer::bind(Uuid::max()).unwrap();
		let incoming = server.run();
		let address = server.listener.local_addr().unwrap();
		thread::spawn(move || {
			client::connect(address, Uuid::nil()).unwrap();
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

		let server = TolliverServer::bind(Uuid::max()).unwrap();
		let incoming = server.run();
		let address = server.listener.local_addr().unwrap();
		thread::spawn(move || {
			let mut conn = client::connect(address, Uuid::nil()).unwrap();
			conn.unreliable_send(EXAMPLE_PROTO_ID, &shirt).unwrap();
		});
		for mut connection in incoming {
			assert_eq!(expected_shirt, connection.read().unwrap().read().unwrap());
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

		let server = TolliverServer::bind(Uuid::max()).unwrap();
		let incoming = server.run();
		let address = server.listener.local_addr().unwrap();
		thread::spawn(move || {
			let mut conn = client::connect(address, Uuid::nil()).unwrap();
			conn.unreliable_send(EXAMPLE_PROTO_ID, &red_shirt).unwrap();
			conn.unreliable_send(EXAMPLE_PROTO_ID, &blue_shirt).unwrap();
			conn.unreliable_send(EXAMPLE_PROTO_ID, &red_shirt).unwrap();
			conn.unreliable_send(EXAMPLE_PROTO_ID, &red_shirt).unwrap();
			conn.unreliable_send(EXAMPLE_PROTO_ID, &blue_shirt).unwrap();
		});
		for mut connection in incoming {
			assert_eq!(
				expected_red_shirt,
				connection.read().unwrap().read().unwrap()
			);
			assert_eq!(
				expected_blue_shirt,
				connection.read().unwrap().read().unwrap()
			);
			assert_eq!(
				expected_red_shirt,
				connection.read().unwrap().read().unwrap()
			);
			assert_eq!(
				expected_red_shirt,
				connection.read().unwrap().read().unwrap()
			);
			assert_eq!(
				expected_blue_shirt,
				connection.read().unwrap().read().unwrap()
			);
			return;
		}
		panic!("Incoming somehow ended")
	}
}
