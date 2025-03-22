# Tolliver

> [!WARNING]
> This library is work in progress and anything can change at any time. Including the endianness of the version number encoding.

A message passing Rust library for sending both fast messages and those that require deliverability guarantees.


## Protocol

2 bytes - big endian u16 of version (max version is therefore 65536)
2 bytes - big endian u16 of the number of bytes in the body (max body size is therefore ~4.254 x 10^22 petabytes)
rest of message - body encoded with protocol buffers
