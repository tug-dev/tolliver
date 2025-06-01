use protobuf::Message;
use tolliver::structs::tolliver_connection::TolliverConnection;

use super::structs::Function;

mod items {
	include!(concat!(env!("OUT_DIR"), "/protos/mod.rs"));
}

pub fn handle_send(function: Function, connection: &mut TolliverConnection) {
	let (_message_name, _message_values) = match (function.args.get(0), function.args.get(1)) {
		(Some(message_name), Some(message_values)) => (message_name, message_values),
		_ => {
			eprintln!("Usage: send <message name> <message values>");
			return;
		}
	};
	//TODO Actually send the message the user entered
	let mut shirt = items::items::Shirt::new();
	shirt.color = "Red".to_string();
	let bytes = shirt.write_to_bytes().unwrap();
	match connection.send_bytes(bytes) {
		Ok(()) => println!("Message sent"),
		Err(e) => {
			eprintln!("Could not send message: {e}")
		}
	}
}
