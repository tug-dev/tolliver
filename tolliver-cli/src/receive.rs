use tolliver::structs::tolliver_connection::TolliverConnection;

use crate::dynamic_proto::message_from_proto_file;

use super::structs::Function;

pub fn handle_receive(function: Function, connection: &mut TolliverConnection) {
	let bytes = match connection.read_bytes() {
		Ok(bytes) => bytes,
		Err(e) => {
			eprintln!("Could not read message: {e}");
			return;
		}
	};
	let (proto_path, message_name) = match (function.args.get(0), function.args.get(1)) {
		(Some(path), Some(name)) => (path, name),
		(None, None) => {
			println!("Bytes: {:?}", bytes);
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
	let message = match message_descriptor.parse_from_bytes(&bytes) {
		Ok(res) => res,
		Err(e) => {
			eprintln!("Could not parse bytes into proto: {e}");
			eprintln!("Bytes: {:?}", bytes);
			return;
		}
	};
	let output = protobuf::text_format::print_to_string(&*message);
	println!("{output}");
}
