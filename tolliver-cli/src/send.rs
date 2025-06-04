use tolliver::structs::tolliver_connection::TolliverConnection;

use crate::dynamic_proto::message_from_proto_file;

use super::structs::Function;

pub fn handle_send(function: Function, connection: &mut TolliverConnection) {
	let (proto_path, message_name) = match (function.args.get(0), function.args.get(1)) {
		(Some(path), Some(name)) => (path, name),
		_ => {
			eprintln!("Usage: send <proto path> <message name> <message values>");
			eprintln!("Example:");
			eprintln!("\t send proto_files/items.proto Shirt color: \"Red\" size: LARGE");
			return;
		}
	};
	let (_, values_strings) = function.args.split_at(2);
	println!("{:?}", values_strings);
	let message_values = values_strings.join(" ");

	let message_descriptor = match message_from_proto_file(proto_path, message_name) {
		Ok(res) => res,
		Err(e) => {
			eprintln!("{e}");
			return;
		}
	};
	let mut message = message_descriptor.new_instance();
	match protobuf::text_format::merge_from_str(&mut *message, &message_values) {
		Ok(()) => {}
		Err(e) => {
			eprintln!("Could not put values into message: {:?}", e);
			return;
		}
	};
	let bytes = message.write_to_bytes_dyn().unwrap();
	match connection.send_bytes(bytes) {
		Ok(()) => println!("Message sent"),
		Err(e) => {
			eprintln!("Could not send message: {e}")
		}
	}
}
