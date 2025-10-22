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
code 2, format: status code 0 for backwards compatibility, 1 for failure.

Subscription and unsubscription messages are to be sent as regular information messages with no key on the reserved "tolliver" channel (as such the API for tolliver should forbid this channel from being used by application level messages). The body of the message will have the format of 4 bytes big endian u32 channel name length - 4 bytes big endian u32 key name length - UTF-8 encoded channel name - UTF-8 encoded key name.

Regular message
code 3, format: 4 bytes big endian u32 channel name length - 4 bytes big endian u32 key name length - 4 bytes big endian u32 body length (therefore max message size is 2gb alex i have genuinely no idea how you figured that u16 for message length gives max message size of 4 petabytes) - 4 bytes big endian u32 local message id (taken from the database, should be set as 2^31 -1 for an unreliable message) - UTF-8 encoded channel name - UTF-8 encoded key name - message body

General acknowledgment
code 4, format: 4 bytes big endian u32 id of acknowledged message

API for the tolliver library
(As much planning for me as anything)

static library method - createInstance(options) returns a new tolliver instance based on the supplied options (TLS certificates, port to listen for incoming connections, database file to use, interval between attempting to reprocess any remaining tasks in the database, list of initial remotes)

methods on the tolliver instance:
NewConnection() - connect to a given host and try to perform the tolliver handshake. blocks indefinitely trying (and retrying) to create a TLS connection. only returns an error if the tolliver handshake fails or the TLS handshake fails due to incorrect certificates. 
RemoveConnection() - remove a remote from the connection pool of the instance.

These methods write the relevant information to the database and make one attempt to complete the action.
Subscribe, Unsubscribe, Send, UnreliableSend

RegisterCallback - registers a callback function on a specific channel key pair.

