use std::fmt::{self};

use crate::StatusCode;

#[derive(Debug, PartialEq, Copy, Clone)]
pub enum HandshakeResponseCode {
	Success = 0,
	GeneralError = 1,
	NewerVersionWithCompatibility = 2,
	NewerVersionWithoutCompatibility = 3,
	OlderVersion = 4, // Handshake final will determine if this succeeds
}

impl fmt::Display for HandshakeResponseCode {
	fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
		let num = *self as StatusCode;
		write!(f, "{num}")
	}
}

#[derive(Debug, PartialEq)]
pub enum HandshakeFinalCode {
	Success = 0,
	GeneralError = 1,
	IncompatibleVersion = 2,
}
