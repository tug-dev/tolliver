use std::{
	fmt::Debug,
	io::{self, Read, Write},
	net::TcpStream,
};

use prost::Message;

use crate::error::TolliverError;

type BodyLengthType = u16;

/// The number of bytes the body length is encoded in
const BODY_LENGTH_LENGTH: usize = 2;

pub struct TolliverConnection {
	pub stream: TcpStream,
}

impl TolliverConnection {
	/// Receive one message from the connection
	pub fn read<T>(&mut self) -> io::Result<T>
	where
		T: Message,
		T: Default + Debug + Send + Sync,
	{
		let mut body_length_buf = [0; BODY_LENGTH_LENGTH];
		self.stream.read_exact(&mut body_length_buf)?;
		let body_length = BodyLengthType::from_be_bytes(body_length_buf);
		let mut body_buf = vec![0; body_length.into()];
		self.stream.read_exact(&mut body_buf)?;
		let message = Message::decode(&mut &body_buf[..]).unwrap();
		Ok(message)
	}

	/// Sends a fast message with no deliverability guarantees. This attempts to
	/// return as early as possible when it fails so something else can be tried.
	pub fn fast_send(&mut self, object: &impl Message) -> Result<(), TolliverError> {
		// See protocol documentation for details
		let body_length: u16 = match object.encoded_len().try_into() {
			Ok(r) => r,
			Err(_) => {
				return Err(TolliverError::TolliverError(
					"Could not encode length into u16, most likely object is too large",
				))
			}
		};

		let total_length = BODY_LENGTH_LENGTH + body_length as usize;
		let mut buf = Vec::with_capacity(total_length);

		let body_length_bytes = body_length.to_be_bytes();
		debug_assert_eq!(body_length_bytes.len(), BODY_LENGTH_LENGTH);
		buf.extend(body_length_bytes);

		let mut body_buf = Vec::with_capacity(body_length.into());
		// Unwrap is safe, since we have reserved sufficient capacity in the vector.
		object.encode(&mut body_buf).unwrap();
		buf.extend(body_buf);
		self.stream.write_all(&buf)?;
		Ok(())
	}
}
