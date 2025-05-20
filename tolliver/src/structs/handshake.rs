use std::{
	error::Error,
	fmt::{self, Display},
	io,
};

use crate::{error::TolliverError, HandshakeCodeType, VersionType, VERSION};

#[derive(Debug, PartialEq)]
pub enum HandshakeCode {
	Success,
	GeneralError,
	IncompatibleVersion(VersionType),
	Unauthorised,
	UnknowErrorCode(HandshakeCodeType),
}

impl Display for HandshakeCode {
	fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
		match self {
			Self::Success => write!(f, "Success"),
			Self::GeneralError => write!(f, "General error"),
			Self::IncompatibleVersion(server_version) => {
				write!(
					f,
					"Incompatible version. Client is {VERSION} while server is {server_version}"
				)
			}
			Self::Unauthorised => write!(f, "Unauthorised"),
			Self::UnknowErrorCode(code) => {
				write!(f, "Unknown error code {code}")
			}
		}
	}
}

impl HandshakeCode {
	/// Constructs a HandshakeResultCode from the result code returned by the server.
	/// # Examples
	///
	/// ```
	/// use tolliver::structs::handshake::HandshakeCode;
	/// let handshake_code = HandshakeCode::from_status_code(3, 0);
	/// assert_eq!(handshake_code, HandshakeCode::Unauthorised);
	/// ```
	///
	/// This can also be done with a specific version type:
	///
	/// ```
	/// use tolliver::structs::handshake::HandshakeCode;
	/// let handshake_code = HandshakeCode::from_status_code(2, 5);
	/// assert_eq!(handshake_code, HandshakeCode::IncompatibleVersion(5));
	/// ```
	pub fn from_status_code(code: HandshakeCodeType, server_version: VersionType) -> Self {
		match code {
			1 => Self::GeneralError,
			2 => Self::IncompatibleVersion(server_version),
			3 => Self::Unauthorised,
			code => Self::UnknowErrorCode(code),
		}
	}

	pub fn status_code(&self) -> HandshakeCodeType {
		match *self {
			Self::Success => 0,
			Self::GeneralError => 1,
			Self::IncompatibleVersion(_) => 2,
			Self::Unauthorised => 3,
			Self::UnknowErrorCode(code) => code,
		}
	}
}

#[derive(Debug)]
pub enum HandshakeError {
	TolliverError(TolliverError),
	Result(HandshakeCode),
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
