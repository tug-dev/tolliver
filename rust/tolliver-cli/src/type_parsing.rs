pub fn hex_string_to_bytes(input: &str) -> Result<Vec<u8>, String> {
	let mut bytes = Vec::new();
	let mut chars = input.chars();
	loop {
		let first_char = match chars.next() {
			Some(res) => res,
			None => break,
		};
		let second_char = match chars.next() {
			Some(res) => res,
			None => return Err("There must be an even number of characters. Every two characters corresponds to a byte.".to_string()),
		};
		let n1 = match hex_to_number(first_char) {
			Some(res) => res,
			None => return Err(format!("Invalid character: {first_char}")),
		};
		let n2 = match hex_to_number(second_char) {
			Some(res) => res,
			None => return Err(format!("Invalid character: {second_char}")),
		};
		let byte = n1 * 16 + n2;
		bytes.push(byte);
	}
	Ok(bytes)
}

fn hex_to_number(character: char) -> Option<u8> {
	match character {
		'0' => Some(0),
		'1' => Some(1),
		'2' => Some(2),
		'3' => Some(3),
		'4' => Some(4),
		'5' => Some(5),
		'6' => Some(6),
		'7' => Some(7),
		'8' => Some(8),
		'9' => Some(9),
		'a' => Some(10),
		'b' => Some(11),
		'c' => Some(12),
		'd' => Some(13),
		'e' => Some(14),
		'f' => Some(15),
		_ => None,
	}
}

#[cfg(test)]
mod hex_string_tests {
	use crate::type_parsing::hex_string_to_bytes;

	#[test]
	fn single_byte() {
		let actual = hex_string_to_bytes("42");
		assert_eq!(actual, Ok(vec![66]));
	}

	#[test]
	fn single_byte_min() {
		let actual = hex_string_to_bytes("00");
		assert_eq!(actual, Ok(vec![0]));
	}

	#[test]
	fn single_byte_max() {
		let actual = hex_string_to_bytes("ff");
		assert_eq!(actual, Ok(vec![255]));
	}

	#[test]
	fn multibyte() {
		let actual = hex_string_to_bytes("abcd");
		assert_eq!(actual, Ok(vec![171, 205]));
	}

	#[test]
	fn multibyte_min() {
		let actual = hex_string_to_bytes("000000");
		assert_eq!(actual, Ok(vec![0, 0, 0]));
	}

	#[test]
	fn multibyte_max() {
		let actual = hex_string_to_bytes("ffffff");
		assert_eq!(actual, Ok(vec![255, 255, 255]));
	}

	#[test]
	fn empty_string() {
		let actual = hex_string_to_bytes("");
		assert_eq!(actual, Ok(vec![]));
	}

	#[test]
	fn odd_number_of_chars() {
		let actual = hex_string_to_bytes("fffff");
		assert!(actual.is_err());
	}

	#[test]
	fn illegal_char() {
		let actual = hex_string_to_bytes("fffgff");
		assert!(actual.is_err());
	}
}
