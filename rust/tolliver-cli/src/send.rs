use tolliver::structs::tolliver_connection::TolliverConnection;

use crate::dynamic_proto::message_from_proto_file;

use super::structs::Function;

pub fn handle_send(function: Function, connection: &mut TolliverConnection) {
	let func_args = &function.args;
	let (channel, key, proto_path, message_name) = match (
		func_args.get(0),
		func_args.get(1),
		func_args.get(2),
		func_args.get(3),
	) {
		(Some(channel), Some(key), Some(path), Some(name)) => (channel, key, path, name),
		_ => {
			eprintln!("Usage: send <channel> <key> <proto path> <message name> <message values>");
			eprintln!(
				"The id of the message will be set to 0, to specify it use the send_id command"
			);
			eprintln!();
			eprintln!("Example:");
			eprintln!("\t send test_channel d366a0c proto_files/items.proto Shirt color: \"Red\" size: LARGE");
			return;
		}
	};
	let values_strings = function.args.split_at(4).1;
	let message_values = values_strings.join(" ");
	send(
		channel,
		key,
		&proto_path,
		&message_name,
		&message_values,
		connection,
	);
}

fn send(
	channel: &str,
	key: &str,
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
	println!("{message_values}");
	match protobuf::text_format::merge_from_str(&mut *message, message_values) {
		Ok(()) => {}
		Err(e) => {
			eprintln!("Could not put values into message: {:?}", e);
			return;
		}
	};
	let bytes = message.write_to_bytes_dyn().unwrap();
	match connection.send_bytes(channel, key, bytes) {
		Ok(()) => println!("Message sent"),
		Err(e) => {
			eprintln!("Could not send message: {e}")
		}
	}
}
