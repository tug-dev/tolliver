use std::{
	io::{stdin, stdout, Write},
	sync::{Arc, Mutex},
	thread,
};

use parser::parse_function;
use receive::handle_receive;
use send::handle_send;
use structs::Function;
use tolliver::{
	client::connect, server::TolliverServer, structs::tolliver_connection::TolliverConnection,
};
use type_parsing::hex_string_to_bytes;

pub mod parser;
pub mod receive;
pub mod send;
pub mod structs;
pub mod type_parsing;

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
		println!("{:?}", function);
		match function.name.as_str() {
			"q" => return,
			"connect" => handle_connection(function, &mut connections.lock().unwrap()),
			"send" => handle_send(function, connections.lock().unwrap().get_mut(0).unwrap()),
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
	let (addr, api_key_string) = match (function.args.get(0), function.args.get(1)) {
		(Some(addr), Some(api_key)) => (addr, api_key),
		_ => {
			eprintln!("Usage: connect <address> <api_key>");
			return;
		}
	};
	let api_key_vec = match hex_string_to_bytes(api_key_string) {
		Ok(key) => key,
		Err(e) => {
			eprintln!("Error parsing API key: {e}");
			return;
		}
	};
	let api_key: [u8; 32] = match api_key_vec.try_into() {
		Ok(key) => key,
		Err(_) => {
			eprintln!("API key cannot be converted to bytes (possibly wrong size)");
			return;
		}
	};
	let connection = match connect(addr, api_key) {
		Ok(res) => res,
		Err(e) => {
			eprintln!("Error connecting: {e}");
			return;
		}
	};
	connections.push(connection);
	println!("Connection established!")
}

/// Starts a Tolliver server and returns a join handle to the thread where it
/// is running.
fn handle_server_start(
	function: Function,
	connections: Arc<Mutex<Vec<TolliverConnection>>>,
) -> Option<thread::JoinHandle<()>> {
	let server_result = match function.args.first() {
		Some(addr) => TolliverServer::bind_at(addr),
		None => TolliverServer::bind(),
	};
	let server = match server_result {
		Ok(res) => res,
		Err(e) => {
			eprintln!("Error when binding to address: {e}");
			return None;
		}
	};
	println!(
		"Server started at {}",
		server.listener.local_addr().unwrap()
	);
	let handle = thread::spawn(move || {
		for conn in server.run() {
			connections.lock().unwrap().push(conn);
		}
	});
	Some(handle)
}

fn get_user_input() -> String {
	stdout().flush().unwrap();
	let mut input = String::new();
	stdin().read_line(&mut input).unwrap();
	return input;
}
