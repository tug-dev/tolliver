My thinking for the handshake is that we want to allow the option for backwards compatibility. So the initial request contains the version number of the first party, the response gives there version number and a status code which is one of the following:

0 - same versions, so handshake success.
1 - newer protocol version but support backwards compatibility so success.
2 - newer protocol version no backwards compatibility so failure.
3 - older protocol version so await final message as to whether the sender supports backwards compatibility. This is the only case where a handshake final message will occur.

On a failure, the party which found the failure sends the message with the error code and closes the connection. Once the handshake is a success both parties begin waiting for other messages. The server must process the subscriptions of the incoming connection synchronously before responding, and naturally no messages should be sent on a connection until the handshake is complete, and new message sending should be blocked while a new connection is being established such that all new messages are sent to even recent remotes.

Handshake req
code 0, format: 8 byte big endian u64 tolliver version - 16 bytes instance UUID(v7) - 4 bytes big endian u32 number of subscriptions - the subscriptions using the same format as for the subscription and unsubscription messages.

Handshake res
code 1, format: 8 byte big endian u64 tolliver version 16 bytes instance UUID(v7) 1 byte status code - 4 bytes big endian u32 number of subscriptions - the subscriptions using the same format as for the subscription and unsubscription messages.


Handshake final
code 2, format: status code 1 for backwards compatibility, 2 for failure (to match response codes)

Although there is no reason for the handshake to be sent again on an existing connection, if a handshake req message is received on an existing connection responses should be sent as normal. If unexpected handshake response or final messages are received they should simply be ignored by the receiving party.

Subscription and unsubscription messages are to be sent as regular information messages with no key on the reserved "tolliver" channel (as such the API for tolliver should forbid this channel from being used by application level messages). The body of the message will have the format of 1 byte code - 0 for sub 1 for unsub, 4 bytes big endian u32 channel name length - 4 bytes big endian u32 key name length - UTF-8 encoded channel name - UTF-8 encoded key name.

Regular message
code 3, format: 4 bytes big endian u32 channel name length - 4 bytes big endian u32 key name length - 4 bytes big endian u32 body length (therefore max message size is 2gb alex i have genuinely no idea how you figured that u16 for message length gives max message size of 4 petabytes) - 4 bytes big endian u32 local message id (taken from the database, should be set as 2^31 -1 for an unreliable message) - UTF-8 encoded channel name - UTF-8 encoded key name - message body

General acknowledgment
code 4, format: 4 bytes big endian u32 id of acknowledged message

The receiver sends an ack once per received message and pass the message to the application level code each time. The sender will resend their message at any interval they see fit until they have received the ack.
