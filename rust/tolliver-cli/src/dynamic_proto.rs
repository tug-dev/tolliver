use std::path::Path;

use protobuf::reflect::{FileDescriptor, MessageDescriptor};
use protobuf_parse::Parser;

pub fn message_from_proto_file(
	proto_path: &str,
	proto_name: &str,
) -> Result<MessageDescriptor, String> {
	// Get directory of .proto file
	let proto_directory = match Path::new(proto_path).parent() {
		Some(res) => res,
		None => {
			return Err(
				"File given has no parent directory (Hint: use ./file instead of file)".to_string(),
			);
		}
	};
	// Parse proto file
	let parsed = match Parser::new()
		.pure()
		.includes(&[proto_directory.to_path_buf()])
		.input(proto_path)
		.parse_and_typecheck()
	{
		Ok(res) => res,
		Err(e) => {
			return Err(format!("Error processing .proto file: {:?}", e));
		}
	};
	let mut file_descriptor_protos = parsed.file_descriptors;

	// This is our .proto file converted to `FileDescriptorProto`
	let file_descriptor_proto = file_descriptor_protos.pop().unwrap();
	// Now this `FileDescriptorProto` initialised for reflective access
	let file_descriptor = FileDescriptor::new_dynamic(file_descriptor_proto, &[]).unwrap();
	// Find the message
	match file_descriptor.message_by_package_relative_name(proto_name) {
		Some(res) => Ok(res),
		None => {
			return Err(format!(
				"Could not find message {proto_name} in file {proto_path}"
			));
		}
	}
}
