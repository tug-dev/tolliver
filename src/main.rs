use std::io::{stdin, stdout, Write};

use cli::{parser::parse_function, type_parsing::hex_string_to_bytes};
use tolliver::{client::connect, structs::tolliver_connection::TolliverConnection};

mod cli;

fn main() {
	let mut connections = Vec::new();
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
			"connect" => handle_connection(function, &mut connections),
			other => {
				println!("Unknown command: {other}")
			}
		}
	}
}

fn handle_connection(function: cli::parser::Function, connections: &mut Vec<TolliverConnection>) {
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

fn get_user_input() -> String {
	stdout().flush().unwrap();
	let mut input = String::new();
	stdin().read_line(&mut input).unwrap();
	return input;
}
