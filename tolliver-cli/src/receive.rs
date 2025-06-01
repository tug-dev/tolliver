use tolliver::structs::tolliver_connection::TolliverConnection;

use super::structs::Function;

mod items {
	include!(concat!(env!("OUT_DIR"), "/protos/mod.rs"));
}

pub fn handle_receive(function: Function, connection: &mut TolliverConnection) {
	let _message_name = match function.args.get(0) {
		Some(message_name) => message_name,
		_ => {
			eprintln!("Usage: receive <proto path>");
			return;
		}
	};
	//TODO Parse the protobuf instead of reading bytes
	match connection.read_bytes() {
		Ok(bytes) => println!("Bytes: {:?}", bytes),
		Err(e) => {
			eprintln!("Could not read message: {e}")
		}
	}
}
