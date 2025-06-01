use std::path::Path;

use protobuf::reflect::FileDescriptor;
use protobuf_parse::Parser;
use tolliver::structs::tolliver_connection::TolliverConnection;

use super::structs::Function;

mod items {
	include!(concat!(env!("OUT_DIR"), "/protos/mod.rs"));
}

pub fn handle_receive(function: Function, connection: &mut TolliverConnection) {
	let bytes = match connection.read_bytes() {
		Ok(bytes) => bytes,
		Err(e) => {
			eprintln!("Could not read message: {e}");
			return;
		}
	};
	let (proto_path, proto_name) = match (function.args.get(0), function.args.get(1)) {
		(Some(path), Some(name)) => (path, name),
		(None, None) => {
			println!("Bytes: {:?}", bytes);
			return;
		}
		_ => {
			eprintln!("Usage: receive <proto path> <message name>");
			return;
		}
	};
	// Get directory of .proto file
	let proto_directory = match Path::new(proto_path).parent() {
		Some(res) => res,
		None => {
			eprintln!("File given has no parent directory (Hint: use ./file instead of file)");
			return;
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
			eprintln!("Error processing .proto file: {:?}", e);
			return;
		}
	};
	let mut file_descriptor_protos = parsed.file_descriptors;

	// This is our .proto file converted to `FileDescriptorProto`
	let file_descriptor_proto = file_descriptor_protos.pop().unwrap();
	// Now this `FileDescriptorProto` initialised for reflective access
	let file_descriptor = FileDescriptor::new_dynamic(file_descriptor_proto, &[]).unwrap();
	// Find the message
	let message_descriptor = match file_descriptor.message_by_package_relative_name(proto_name) {
		Some(res) => res,
		None => {
			eprintln!("Could not find message {proto_name} in file {proto_path}");
			return;
		}
	};
	let message = match message_descriptor.parse_from_bytes(&bytes) {
		Ok(res) => res,
		Err(e) => {
			eprintln!("Could not parse bytes into proto: {e}");
			eprintln!("Bytes: {:?}", bytes);
			return;
		}
	};
	let output = protobuf::text_format::print_to_string(&*message);
	println!("{output}");
}
