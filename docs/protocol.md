# Tolliver Protocol Version 1

## Overview

Can we for simplicity have every single message start with a 1 byte message type including handshake. I see no reason not to and it will make implementation a little cleaner.


Tolliver provides two key features:
- Messages which are guaranteed to deliver eventually
- Organising the sending and receiving of messages by channels (a broad type of message e.g. vm-shutdown) and keys (specific identifiers - a sparkler node id). Messages are published to a channel or key or both and similarly handlers on incoming messages are registered by these categories.

It is likely that each message channel will only contain data which is of a specific proto buf type, however this is not to be enforced by Tolliver.

## Messages

### Initial handshake

The server and the client first establish a TCP socket between them, after which the client sends a hello message that authenticates it and has information about it's version. For this to be sent in a single TCP segment it should be below 536 bytes (or 1448 bytes, haven't figured out which yet). The client sends a message in the following format:

<!-- Is there any problem with the handshake being over multiple TCP/TLS segments? -->
<!---->
<!-- Given we are using TLS forgo API key and send the client UUID in the handshake? -->
<!-- Also add the subscriptions that are wanted -->


```
8 bytes - big endian u64 of client version (max version is therefore 2^64)
32 bytes - 256-bit api key
```

Then the server replies in the following format:

<!-- Again include server UUID -->

```
1 byte - handshake status code, with a 0 corresponding with success while 1-255 being an error.
8 bytes - big endian u64 of server version
```

### Channel subscription message

```
1 byte - message type, for subscription this is 1
4 bytes - big endian u32 of the number of bytes the channel string is
Rest of message - UTF-8 encoded string of the channel name
```

<!-- Potentially include the channel name in the acknowledgement in the case that a client tried to subscribe to multiple channels. -->
### Subscription response

```
1 byte - message type, for subscription response this is 2
1 bytes - the channel subscription status code
```

### Channel unsubscription message

```
1 byte - message type, for unsubscription this is 3
4 bytes - big endian u32 of the number of bytes the channel string is
Rest of message - UTF-8 encoded string of the channel name
```

### Unsubscription response

```
1 byte - message type, for unsubscription response this is 4
1 bytes - the channel subscription status code
```

<!-- Add in messages for subscription to all messages with a given key and a message and key packet. -->

### Information message

<!-- Remove the proto format id and add a message identifier? -->

```
1 byte - message type, for info message this is 3
8 bytes - big endian u64 of the id of the proto format of the message (therefore max of ~1.8 x 10^19 unique proto formats can be used)
2 bytes - big endian u16 of the number of bytes in the body (max body size is therefore ~4.254 x 10^22 petabytes)
Rest of message - body encoded with protocol buffers
```

### Information message response

```
1 byte - message type, for info message response this is 4
1 byte - information message response status code
---Everything beyond this is only included if the status code is 1---
8 bytes - big endian u64 of the id of the proto format of the message (therefore max of ~1.8 x 10^19 unique proto formats can be used)
2 bytes - big endian u16 of the number of bytes in the body (max body size is therefore ~4.254 x 10^22 petabytes)
Rest of message - body encoded with protocol buffers
```

## Status codes

### Handshake status codes

```
0 - Success
1 - General error
2 - Incompatible version
3 - Unauthorized
```

### Channel subscription status codes

```
0 - Success
1 - General error
2 - Unauthorized
```

### Information message status codes
```
0 - Success
1 - Success with aditional information
2 - General error
```

## Versioning

The Tolliver client and server versions are always kept in sync. While the version code is 0, the protocol may change at any time without warning.
