pub mod client;
pub mod server;

#[cfg(test)]
mod tests {
	use super::*;

	#[test]
	fn start_server() {
		let server = server::TolliverServer::bind().unwrap().run().unwrap();
		client::connect().unwrap();
		server.wait().unwrap().unwrap();
	}
}
