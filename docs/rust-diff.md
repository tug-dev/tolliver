# Rust implementation differences vs docs/protocol.md

- Handshake subscriptions: docs include subscription lists in handshake request/response messages; Rust sends only message type, version, UUID, and response status code. See `rust/tolliver/src/client/mod.rs` and `rust/tolliver/src/structs/incoming.rs`.
- Regular message format: docs define message type 3, message id, channel/key strings, and body length as u64; Rust uses message type 3, channel/key strings, and `body_len (u16)`, but omits the message id. See `rust/tolliver/src/structs/tolliver_connection.rs`.
- Acknowledgments: docs define message type 4 acks and resend-until-ack semantics; Rust doesn’t implement ack messages or resend-on-ack at all. See `rust/tolliver/src/structs/tolliver_connection.rs`.
- Subscriptions: docs define subscription/unsubscription messages on a reserved `tolliver` channel; Rust has no subscription handling. See `rust/tolliver/src/structs/tolliver_connection.rs`.
- Transport: docs require TLS; Rust uses raw `TcpStream`/`TcpListener` only. See `rust/tolliver/src/client/mod.rs` and `rust/tolliver/src/server/mod.rs`.
- Repeat handshakes: docs say a handshake request received on an existing connection should be handled normally and unexpected handshake response/final messages should be ignored; Rust only accepts regular messages after connection setup and returns an error for any other message type. See `rust/tolliver/src/structs/tolliver_connection.rs`.
