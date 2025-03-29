use std::{error::Error, fmt, io};

#[derive(Debug)]
pub enum TolliverError<'a> {
	IOError(io::Error),
	TolliverError(&'a str),
}

impl<'a> fmt::Display for TolliverError<'a> {
	fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
		match self {
			TolliverError::IOError(e) => write!(f, "{}", e),
			TolliverError::TolliverError(e) => write!(f, "{}", e),
		}
	}
}

impl<'a> Error for TolliverError<'a> {
	fn source(&self) -> Option<&(dyn Error + 'static)> {
		match self {
			TolliverError::IOError(e) => Some(e),
			TolliverError::TolliverError(_) => None,
		}
	}
}

impl<'a> From<io::Error> for TolliverError<'a> {
	fn from(value: io::Error) -> Self {
		TolliverError::IOError(value)
	}
}
