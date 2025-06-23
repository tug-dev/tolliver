use std::fmt::Debug;

use prost::{DecodeError, Message};

use super::tolliver_connection::ProtoIdType;

#[derive(Debug)]
pub struct ReadMessage {
	pub proto_id: ProtoIdType,
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
