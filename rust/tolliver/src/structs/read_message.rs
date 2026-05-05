use std::fmt::Debug;

use prost::{DecodeError, Message};

#[derive(Debug)]
pub struct ReadMessage {
	pub channel: Box<str>,
	pub key: Box<str>,
	pub body: Vec<u8>,
}

impl ReadMessage {
	pub fn read<T>(&mut self) -> Result<T, DecodeError>
	where
		T: Message,
		T: Default + Debug + Send + Sync,
	{
		Message::decode(&mut &self.body[..])
	}
}
