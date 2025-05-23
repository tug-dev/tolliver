use tolliver::structs::tolliver_connection::TolliverConnection;

use super::structs::Function;

mod items {
	include!(concat!(env!("OUT_DIR"), "/protos/mod.rs"));
}

pub fn handle_receive(function: Function, _connections: &mut Vec<TolliverConnection>) {
	let _message_name = match function.args.get(0) {
		Some(message_name) => message_name,
		_ => {
			eprintln!("Usage: receive <proto path>");
			return;
		}
	};
	//TODO Implement receive
	eprintln!("Receive not yet implemented.")
}
