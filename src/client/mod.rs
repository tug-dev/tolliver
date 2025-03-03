use std::{
    io::{self, Write},
    net::TcpStream,
};

pub fn connect() -> io::Result<()> {
    let mut stream = TcpStream::connect("127.0.0.1:8080")?;

    let data = b"shutdown";

    stream.write_all(data)?;

    Ok(())
}
