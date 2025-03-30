use std::{error::Error, fmt, io};

#[derive(Debug)]
pub enum TolliverError {
	TolliverError(String),
	IOError(io::Error),
	SqliteError(rusqlite::Error),
}

impl fmt::Display for TolliverError {
	fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
		match self {
			TolliverError::TolliverError(e) => write!(f, "{}", e),
			TolliverError::IOError(e) => write!(f, "IO error: {}", e),
			TolliverError::SqliteError(e) => write!(f, "SQLite eror: {}", e),
		}
	}
}

impl Error for TolliverError {
	fn source(&self) -> Option<&(dyn Error + 'static)> {
		match self {
			TolliverError::TolliverError(_) => None,
			TolliverError::IOError(e) => Some(e),
			TolliverError::SqliteError(e) => Some(e),
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
