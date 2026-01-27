# Tolliver Protocol Version 1

## Overview

Tolliver provides two key features:
- Messages which are guaranteed to deliver eventually
- Organising the sending and receiving of messages by channels (a broad type of message e.g. vm-shutdown) and keys (specific identifiers - a sparkler node id). Messages are published to a channel or key or both and similarly handlers on incoming messages are registered by these categories.

It is likely that each message channel will only contain data which is of a specific proto buf type, however this is not to be enforced by Tolliver.

## Messages

### Initial handshake

The server and the client first establish a TCP socket with TLS between them, after which the client sends a hello message that has information about it's version and what channels it would like to subscribe to from the start.

#### Handshake request

The client sends a message in the following format:

```
1 byte - message type, for initial handshake this is 0
8 bytes - big endian u64 of client version (max version is therefore 2^64)
16 bytes - client UUID v7
Rest of the message - In the format of a subscription message
```

#### Handshake response

Then the server replies in the following format:

```
1 byte - message type, for handshake response this is 1
8 bytes - big endian u64 of server version
16 bytes - server UUID v7
1 byte - handshake response code
Rest of the message - In the format of a subscription message
```

#### Handshake final

And then the client replies with:

```
1 byte - message type, for handshake final this is 2
1 byte - handshake final code
```

On a failure, the party which found the failure sends the message with the error code and closes the connection. Once the handshake is a success both parties begin waiting for other messages. The server must process the subscriptions of the incoming connection synchronously before responding, and naturally no messages should be sent on a connection until the handshake is complete, and new message sending should be blocked while a new connection is being established such that all new messages are sent to even recent remotes.

Although there is no reason for the handshake to be sent again on an existing connection, if a handshake req message is received on an existing connection responses should be sent as normal. If unexpected handshake response or final messages are received they should simply be ignored by the receiving party.

### Regular message

```
1 byte - message type, for a regular message this is 3
8 bytes - big endian u64 of local message id (taken from the database, should be set as 0 for an unreliable message)
8 bytes - big endian u64 of the number of bytes the channel string is
Number of bytes specified - UTF-8 encoded channel name
8 bytes - big endian u64 of the number of bytes the key string is
Number of bytes specified - UTF-8 encoded key name
8 bytes - big endian u64 of the number of bytes the body is
Number of bytes specified - Message body
```

### Regular message acknowledgment

```
1 byte - message type, for a regular message acknowledgment this is 4
1 byte - regular message acknowledgment status code
8 bytes - big endian u64 of the local message id
```

The receiver sends an ack once per received message and pass the message to the application level code each time. The sender will resend their message at any interval they see fit until they have received the ack.

### Subscription message

Subscription and unsubscription messages are to be sent as regular messages with no key on the reserved "tolliver" channel (as such the API for tolliver should forbid this channel from being used by application level messages). The body of the message will have the format of:

```
1 byte - 0 for subscription 1 for unsubscription
8 bytes - number of channels to (un)subscribe to
Repeated for each channel that needs to be (un)subscribed to:
  8 bytes - big endian u64 of the number of bytes the channel string is
  Number of bytes specified - UTF-8 encoded channel name
  8 bytes - big endian u64 of the number of bytes the key string is
  Number of bytes specified - UTF-8 encoded key name
```

## Status codes

### Handshake status codes

```
0 - Success
1 - General error
2 - Newer protocol version but support backwards compatibility so success.
3 - Newer protocol version no backwards compatibility so failure.
4 - older protocol version so await final message as to whether the sender supports backwards compatibility. This is the only case where a handshake final message will occur.
```

### Handshake response status codes

```
0 - Success
1 - General error
2 - Newer protocol version but support backwards compatibility so success.
3 - Newer protocol version no backwards compatibility so failure.
4 - Older protocol version so await final message as to whether the sender supports backwards compatibility. This is the only case where a handshake final message will occur.
```

### Handshake final status codes

```
0 - Success
1 - General error
2 - Incompatible version
```

### Regular message acknowledgment code

```
0 - Success
1 - General error
```

## Versioning

The Tolliver client and server versions are always kept in sync. While the version code is 0, the protocol may change at any time without warning.
