use std::{
	fmt::Debug,
	io::{self, Read, Write},
	net::TcpStream,
};

use prost::Message;

type VersionType = u16;
type BodyLengthType = u16;

/// The version of the protocol
const VERSION: VersionType = 0;
/// The number of bytes the version number is encoded in
const VERSION_LENGTH: usize = 2;
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
		let mut version_buf = [0; VERSION_LENGTH];
		self.stream.read_exact(&mut version_buf)?;
		let version = VersionType::from_be_bytes(version_buf);
		if version != VERSION {
			todo!("Version mismatch")
		}
		let mut body_length_buf = [0; BODY_LENGTH_LENGTH];
		self.stream.read_exact(&mut body_length_buf)?;
		let body_length = BodyLengthType::from_be_bytes(body_length_buf);
		let mut body_buf = vec![0; body_length.into()];
		self.stream.read_exact(&mut body_buf)?;
		let message = Message::decode(&mut &body_buf[..]).unwrap();
		Ok(message)
	}

	/// Sends a fast message with no deliverability guarantees
	pub fn fast_send(&mut self, object: impl Message) -> io::Result<()> {
		// See protocol documentation for details
		let body_length: u16 = match object.encoded_len().try_into() {
			Ok(r) => r,
			Err(_) => panic!("could not encode length into u16, most likely object is too large"),
		};

		let total_length = VERSION_LENGTH + BODY_LENGTH_LENGTH + body_length as usize;
		let mut buf = Vec::with_capacity(total_length);

		let version_bytes = VERSION.to_be_bytes();
		debug_assert_eq!(version_bytes.len(), VERSION_LENGTH);
		buf.extend(version_bytes);
		let body_length_bytes = body_length.to_be_bytes();
		debug_assert_eq!(body_length_bytes.len(), BODY_LENGTH_LENGTH);
		buf.extend(body_length_bytes);

		let mut body_buf = Vec::with_capacity(body_length.into());
		// Unwrap is safe, since we have reserved sufficient capacity in the vector.
		object.encode(&mut body_buf).unwrap();
		buf.extend(body_buf);
		self.stream.write_all(&buf)
	}
}
