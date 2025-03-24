use std::{error::Error, fmt, io};

#[derive(Debug)]
pub enum TolliverError {
	IOError(io::Error),
	TolliverError(String),
}

impl fmt::Display for TolliverError {
	fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
		match self {
			TolliverError::IOError(e) => write!(f, "{}", e),
			TolliverError::TolliverError(e) => write!(f, "{}", e),
		}
	}
}

impl Error for TolliverError {
	fn source(&self) -> Option<&(dyn Error + 'static)> {
		match self {
			TolliverError::IOError(e) => Some(e),
			TolliverError::TolliverError(_) => None,
		}
	}
}

impl From<io::Error> for TolliverError {
	fn from(value: io::Error) -> Self {
		TolliverError::IOError(value)
	}
}
