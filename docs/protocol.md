# Initial handshake

The server and the client first establish a TCP socket between them, after which the client sends a hello message that authenticates it and has information about it's version. The client sends a message in the following format:

```
2 bytes - big endian u16 of client version (max version is therefore 65536)
32 bytes - 256-bit api key
```

Then the server replies in the following format:

```
1 byte - handshake (success/error) code, with a 0 corresponding with success while 1-255 being an error.
2 bytes - big endian u16 of server version
```

## Server handshake codes

```
0 - Success
1 - General error
2 - Incompatible version
3 - Unauthorized
```

# Information messages

```
2 bytes - big endian u16 of the number of bytes in the body (max body size is therefore ~4.254 x 10^22 petabytes)
Rest of message - body encoded with protocol buffers
```

# Versioning

The Tolliver client and server versions are always kept in sync. While the version code is 0, the protocol may change at any time without warning.
