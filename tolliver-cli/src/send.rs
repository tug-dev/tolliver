use tolliver::structs::tolliver_connection::TolliverConnection;

use super::structs::Function;

mod items {
	include!(concat!(env!("OUT_DIR"), "/snazzy.items.rs"));
}

pub fn handle_send(function: Function, _connections: &mut Vec<TolliverConnection>) {
	let (_message_name, _message_values) = match (function.args.get(0), function.args.get(1)) {
		(Some(message_name), Some(message_values)) => (message_name, message_values),
		_ => {
			eprintln!("Usage: send <message name> <message values>");
			return;
		}
	};
	//TODO Implement send
	eprintln!("Send not yet implemented.")
}
