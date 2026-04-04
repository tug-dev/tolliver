use std::{
	error::Error,
	fmt::{self},
	io,
};

use crate::{error::TolliverError, StatusCode};

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

#[derive(Debug)]
pub enum HandshakeError {
	TolliverError(TolliverError),
	Result(HandshakeResponseCode),
}

impl fmt::Display for HandshakeError {
	fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
		match self {
			HandshakeError::TolliverError(e) => write!(f, "{}", e),
			HandshakeError::Result(code) => write!(f, "{}", code),
		}
	}
}

impl Error for HandshakeError {
	fn source(&self) -> Option<&(dyn Error + 'static)> {
		match self {
			HandshakeError::TolliverError(e) => e.source(),
			_ => None,
		}
	}
}

impl From<io::Error> for HandshakeError {
	fn from(value: io::Error) -> Self {
		HandshakeError::TolliverError(value.into())
	}
}

impl From<TolliverError> for HandshakeError {
	fn from(value: TolliverError) -> Self {
		HandshakeError::TolliverError(value)
	}
}
