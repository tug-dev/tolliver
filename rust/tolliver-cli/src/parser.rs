use std::{fmt::Display, str::Chars};

use super::structs::Function;

pub struct CLIError {
	pub message: String,
}

impl Display for CLIError {
	fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
		write!(f, "{}", self.message)
	}
}

pub fn parse_function(mut chars: Chars<'_>) -> Result<Function, CLIError> {
	let mut words = Vec::new();
	let mut current_string = String::new();
	loop {
		match chars.next() {
			Some(' ') => {
				words.push(current_string);
				current_string = String::new();
				continue;
			}
			Some(char) => current_string.push(char),
			None => {
				words.push(current_string);
				break;
			}
		};
	}
	if words.is_empty() {
		return Err(CLIError {
			message: "Empty input".to_string(),
		});
	}
	let name = words.remove(0);
	Ok(Function { name, args: words })
}
