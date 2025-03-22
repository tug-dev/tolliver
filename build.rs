use std::io::Result;
fn main() -> Result<()> {
	prost_build::compile_protos(&["src/proto_files/items.proto"], &["src/"])?;
	Ok(())
}
