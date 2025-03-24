use std::{error::Error, fmt, io};

use crate::client::VERSION;

use super::VersionType;

#[derive(Debug)]
pub enum HandshakeError {
	IOError(io::Error),
	GeneralError,
	IncompatibleVersion(VersionType),
	Unauthorized,
	UnknowErrorCode(u8),
}

impl fmt::Display for HandshakeError {
	fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
		match self {
			HandshakeError::IOError(e) => write!(f, "{}", e),
			HandshakeError::GeneralError => write!(f, "General error"),
			HandshakeError::IncompatibleVersion(server_version) => {
				write!(
					f,
					"Incompatible version. Client is {VERSION} while server is {server_version}"
				)
			}
			HandshakeError::Unauthorized => write!(f, "Unauthorised"),
			HandshakeError::UnknowErrorCode(code) => write!(f, "Unknown error code {code}"),
		}
	}
}

impl Error for HandshakeError {
	fn source(&self) -> Option<&(dyn Error + 'static)> {
		match self {
			HandshakeError::IOError(e) => Some(e),
			_ => None,
		}
	}
}

impl From<io::Error> for HandshakeError {
	fn from(value: io::Error) -> Self {
		HandshakeError::IOError(value)
	}
}
