use std::io::Result;
fn main() -> Result<()> {
	prost_build::compile_protos(&["../../proto_files/items.proto"], &["../../proto_files/"])?;
	Ok(())
}
