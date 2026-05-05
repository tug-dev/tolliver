use std::{error::Error, fmt, io, string};

#[derive(Debug)]
pub enum TolliverError {
	Custom(String),
	IOError(io::Error),
	SqliteError(rusqlite::Error),
	InvalidUtf8(string::FromUtf8Error),
}

impl fmt::Display for TolliverError {
	fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
		match self {
			TolliverError::Custom(e) => write!(f, "{}", e),
			TolliverError::IOError(e) => write!(f, "IO error: {}", e),
			TolliverError::SqliteError(e) => write!(f, "SQLite eror: {}", e),
			TolliverError::InvalidUtf8(e) => write!(f, "Invalid UTF-8: {}", e),
		}
	}
}

impl Error for TolliverError {
	fn source(&self) -> Option<&(dyn Error + 'static)> {
		match self {
			TolliverError::Custom(_) => None,
			TolliverError::IOError(e) => Some(e),
			TolliverError::SqliteError(e) => Some(e),
			TolliverError::InvalidUtf8(e) => Some(e),
		}
	}
}

impl From<io::Error> for TolliverError {
	fn from(value: io::Error) -> Self {
		TolliverError::IOError(value)
	}
}

impl From<rusqlite::Error> for TolliverError {
	fn from(value: rusqlite::Error) -> Self {
		TolliverError::SqliteError(value)
	}
}

impl From<string::FromUtf8Error> for TolliverError {
	fn from(value: string::FromUtf8Error) -> Self {
		TolliverError::InvalidUtf8(value)
	}
}
