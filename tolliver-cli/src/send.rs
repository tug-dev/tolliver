use tolliver::structs::tolliver_connection::{ProtoIdType, TolliverConnection};

use crate::dynamic_proto::message_from_proto_file;

use super::structs::Function;

pub fn handle_id_send(function: Function, connection: &mut TolliverConnection) {
	let func_args = &function.args;
	let (proto_path, message_name, proto_id) =
		match (func_args.get(0), func_args.get(1), func_args.get(2)) {
			(Some(path), Some(name), Some(proto_id)) => (path, name, proto_id),
			_ => {
				eprintln!(
					"Usage: id_send <proto path> <message name> <proto message id> <message values>"
				);
				eprintln!(
					"This is the same as the send command except you can specify the id of the message"
				);
				eprintln!();
				eprintln!("Example:");
				eprintln!("\t send proto_files/items.proto Shirt 3 color: \"Red\" size: LARGE");
				return;
			}
		};
	let values_strings = function.args.split_at(3).1;
	let message_values = values_strings.join(" ");
	let proto_id_num = match proto_id.parse() {
		Ok(res) => res,
		Err(e) => {
			eprintln!(
				"Proto id must be a number but \"{proto_id}\" could not be parsed as one: {e}"
			);
			return;
		}
	};
	send(
		proto_id_num,
		&proto_path,
		&message_name,
		&message_values,
		connection,
	);
}

pub fn handle_send(function: Function, connection: &mut TolliverConnection) {
	let func_args = &function.args;
	let (proto_path, message_name) = match (func_args.get(0), func_args.get(1)) {
		(Some(path), Some(name)) => (path, name),
		_ => {
			eprintln!("Usage: send <proto path> <message name> <message values>");
			eprintln!(
				"The id of the message will be set to 0, to specify it use the send_id command"
			);
			eprintln!();
			eprintln!("Example:");
			eprintln!("\t send proto_files/items.proto Shirt color: \"Red\" size: LARGE");
			return;
		}
	};
	let values_strings = function.args.split_at(2).1;
	let message_values = values_strings.join(" ");
	send(0, &proto_path, &message_name, &message_values, connection);
}

fn send(
	proto_id: ProtoIdType,
	proto_path: &str,
	message_name: &str,
	message_values: &str,
	connection: &mut TolliverConnection,
) {
	let message_descriptor = match message_from_proto_file(proto_path, message_name) {
		Ok(res) => res,
		Err(e) => {
			eprintln!("{e}");
			return;
		}
	};
	let mut message = message_descriptor.new_instance();
	match protobuf::text_format::merge_from_str(&mut *message, message_values) {
		Ok(()) => {}
		Err(e) => {
			eprintln!("Could not put values into message: {:?}", e);
			return;
		}
	};
	let bytes = message.write_to_bytes_dyn().unwrap();
	match connection.send_bytes(proto_id, bytes) {
		Ok(()) => println!("Message sent"),
		Err(e) => {
			eprintln!("Could not send message: {e}")
		}
	}
}
