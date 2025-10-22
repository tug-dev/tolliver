My thinking for the handshake is that we want to allow the option for backwards compatibility. So the initial request contains the version number of the first party, the response gives there version number and a status code which is one of the following:

0 - same versions, so handshake success.
1 - newer protocol version but support backwards compatibility so success.
2 - newer protocol version no backwards compatibility so failure.
3 - older protocol version so await final message as to whether the sender supports backwards compatibility. This is the only case where a handshake final message will occur.

On a failure, the party which found the failure sends the message with the error code and closes the connection. Once the handshake is a success both parties begin waiting for other messages.

Handshake req
code 0, format: 8 byte big endian u64 tolliver version 16 bytes instance UUID(v7)

Handshake res
code 1, format: 8 byte big endian u64 tolliver version 16 bytes instance UUID(v7) 1 byte status code

Handshake final
code 2, format: status code 0 for backwards compatibility, 1 for failure.

My thinking is that subscription requests should be persistent (in the db) which this message format allows for.

Subscription message
code 3, format: 4 bytes big endian u32 of channel name length (0 is valid here for no channel) 4 bytes big endian u32 of key name length (again 0 is valid) UTF-8 encoded channel name followed directly by key name.

Subscription confirmed
code 4, format: echo of the subscription message (not including the message code)

Unsubscription message
code 5, format: exact same as subscription message

Unsubscription confirmed
code 6, format: exact same as unsubscription confirmation.

Information (regular) message
code 7, format: 4 bytes big endian u32 channel name length - 4 bytes big endian u32 key name length - 4 bytes big endian u32 body length (therefore max message size is 2gb alex i have genuinely no idea how you figured that u16 for message length gives max message size of 4 petabytes) - 4 bytes big endian u32 local message id (taken from the database, should be set as 2^31 -1 for an unreliable message) - UTF-8 encoded channel name - UTF-8 encoded key name - message body

Information (regular) message acknowledgment
code 8, format: 4 bytes big endian u32 id of acknowledged message

API for the tolliver library
(As much planning for me as anything)

static library method - createInstance(options) returns a new tolliver instance based on the supplied options (TLS certificates, port to listen for incoming connections, database file to use, interval between attempting to reprocess any remaining tasks in the database)

methods on the tolliver instance:
NewConnection() - connect to a given host and try to perform the tolliver handshake. blocks indefinitely trying (and retrying) to create a TLS connection. only returns an error if the tolliver handshake fails or the TLS handshake fails due to incorrect certificates. 

These methods write the relevant information to the database and make one attempt to complete the action.
Subscribe, Unsubscribe, Send, UnreliableSend

RegisterCallback - registers a callback function on a specific channel key pair.

