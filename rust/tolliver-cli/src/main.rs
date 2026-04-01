use std::{
	io::{stdin, stdout, Write},
	sync::{Arc, Mutex},
	thread,
};

use parser::parse_function;
use receive::handle_receive;
use send::{handle_id_send, handle_send};
use structs::Function;
use tolliver::{
	client::connect, server::TolliverServer, structs::tolliver_connection::TolliverConnection,
};
use uuid::Uuid;

pub mod dynamic_proto;
pub mod parser;
pub mod receive;
pub mod send;
pub mod structs;

fn main() {
	let connections = Arc::new(Mutex::new(Vec::new()));
	println!("Tolliver repl:");
	loop {
		print!(">>> ");
		let raw_input = get_user_input();
		let input = raw_input.trim();
		if input.is_empty() {
			continue;
		};
		let function = match parse_function(input.chars()) {
			Ok(function) => function,
			Err(err) => {
				eprintln!("Error: {err}");
				continue;
			}
		};
		match function.name.as_str() {
			"q" => return,
			"connect" => handle_connection(function, &mut connections.lock().unwrap()),
			"send" => handle_send(function, connections.lock().unwrap().get_mut(0).unwrap()),
			"id_send" => handle_id_send(function, connections.lock().unwrap().get_mut(0).unwrap()),
			"start" => {
				handle_server_start(function, connections.clone());
			}
			"receive" => handle_receive(function, connections.lock().unwrap().get_mut(0).unwrap()),
			other => {
				println!("Unknown command: {other}")
			}
		}
	}
}

fn handle_connection(function: Function, connections: &mut Vec<TolliverConnection>) {
	let uuid = match get_uuid(function.args.get(1)) {
		Some(res) => res,
		None => return,
	};
	let addr = match function.args.get(0) {
		Some(addr) => addr,
		_ => {
			eprintln!("Usage: connect <address> [uuid]");
			return;
		}
	};
	let connection = match connect(addr, uuid) {
		Ok(res) => res,
		Err(e) => {
			eprintln!("Error connecting: {e}");
			return;
		}
	};
	println!("Connection established to {} with UUID {}", addr, uuid);
	connections.push(connection);
}

/// Starts a Tolliver server and returns a join handle to the thread where it
/// is running.
fn handle_server_start(
	function: Function,
	connections: Arc<Mutex<Vec<TolliverConnection>>>,
) -> Option<thread::JoinHandle<()>> {
	let uuid = get_uuid(function.args.get(1))?;
	let server_result = match function.args.get(0) {
		Some(addr) => TolliverServer::bind_at(addr, uuid),
		None => TolliverServer::bind(uuid),
	};
	let server = match server_result {
		Ok(res) => res,
		Err(e) => {
			eprintln!("Error when binding to address: {e}");
			return None;
		}
	};
	println!(
		"Server started at {} with UUID {}",
		server.listener.local_addr().unwrap(),
		server.uuid
	);
	let handle = thread::spawn(move || {
		for conn in server.run() {
			connections.lock().unwrap().push(conn);
		}
	});
	Some(handle)
}

fn get_uuid(string: Option<&String>) -> Option<Uuid> {
	let uuid_result = match string {
		Some(uuid_str) => Uuid::parse_str(uuid_str),
		None => Ok(Uuid::now_v7()),
	};
	match uuid_result {
		Ok(res) => Some(res),
		Err(e) => {
			eprintln!("Invalid UUID: {e}");
			None
		}
	}
}

fn get_user_input() -> String {
	stdout().flush().unwrap();
	let mut input = String::new();
	stdin().read_line(&mut input).unwrap();
	return input;
}
