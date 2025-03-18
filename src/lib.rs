pub mod client;
pub mod server;

#[cfg(test)]
mod tests {
	use super::*;

	#[test]
	fn start_server() {
		let server = server::TolliverServer::bind().unwrap().run().unwrap();
		server.send_stop().unwrap();
		client::TolliverClient::connect().unwrap();
		server.wait().unwrap().unwrap();
	}
}
