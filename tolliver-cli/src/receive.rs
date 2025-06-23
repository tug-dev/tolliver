use tolliver::structs::tolliver_connection::TolliverConnection;

use crate::dynamic_proto::message_from_proto_file;

use super::structs::Function;

pub fn handle_receive(function: Function, connection: &mut TolliverConnection) {
	let read_message = match connection.read() {
		Ok(res) => res,
		Err(e) => {
			eprintln!("Could not read message: {e}");
			return;
		}
	};
	let (proto_path, message_name) = match (function.args.get(0), function.args.get(1)) {
		(Some(path), Some(name)) => (path, name),
		(None, None) => {
			println!("Bytes: {:?}", read_message.body);
			return;
		}
		_ => {
			eprintln!("Usage: receive <proto path> <message name>");
			return;
		}
	};
	let message_descriptor = match message_from_proto_file(proto_path, message_name) {
		Ok(res) => res,
		Err(e) => {
			eprintln!("{e}");
			return;
		}
	};
	let message = match message_descriptor.parse_from_bytes(&read_message.body) {
		Ok(res) => res,
		Err(e) => {
			eprintln!("Could not parse bytes into proto: {e}");
			eprintln!("Bytes: {:?}", read_message.body);
			return;
		}
	};
	let output = protobuf::text_format::print_to_string(&*message);
	println!("Proto ID: {}", read_message.proto_id);
	println!("{output}");
}
