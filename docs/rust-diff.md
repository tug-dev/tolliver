# Rust implementation differences vs docs/protocol.md

- Handshake final: docs define a third “handshake final” message; Rust has no such step. See `rust/tolliver/src/client/mod.rs` and `rust/tolliver/src/structs/incoming.rs`.
- Handshake status codes: docs define codes 0–4 with backward-compat semantics; Rust defines 0 success, 1 general, 2 incompatible, 3 unauthorised, no code 4. See `rust/tolliver/src/structs/handshake.rs`.
- Regular message format: docs define message type 3, message id, channel/key strings, and body length as u64; Rust uses `proto_id (u64) + body_len (u16) + body` with no message type, id, channel, or key. See `rust/tolliver/src/structs/tolliver_connection.rs`.
- Acknowledgments: docs define message type 4 acks and resend-until-ack semantics; Rust doesn’t implement ack messages or resend-on-ack at all. See `rust/tolliver/src/structs/tolliver_connection.rs`.
- Subscriptions: docs define subscription/unsubscription messages on a reserved `tolliver` channel; Rust has no subscription handling. See `rust/tolliver/src/structs/tolliver_connection.rs`.
- Transport: docs require TLS; Rust uses raw `TcpStream`/`TcpListener` only. See `rust/tolliver/src/client/mod.rs` and `rust/tolliver/src/server/mod.rs`.
