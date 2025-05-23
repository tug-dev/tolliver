use std::io::Result;

fn main() -> Result<()> {
	protobuf_codegen::Codegen::new()
		.includes(&["../proto_files/"])
		.input("../proto_files/items.proto")
		.cargo_out_dir("protos")
		.run_from_script();
	Ok(())
}
