package tolliver

import (
	"crypto/x509"
	"time"
)

// Defines the details for connecting to a server
type ConnectionAddr struct {
	Host       string
	Port       int
	CA         *x509.Certificate
	ClientCert *x509.Certificate
}

// The options for creating a new client
type ClientOptions struct {
	RemoteAddrs  []ConnectionAddr
	DatabasePath string
	// This is the time after which a message will be resent if it's not acknowledged an that SendAndWait will timeout after.
	Timeout time.Time
}

// Creates a new client instance.
// For each remote address provided in the options, NewConnection is called on the newly created client.
// After opening the database, the client will generate a new ID for this node if one does not already exist.
func NewClient(options ClientOptions) Client

// Defines the ID of a channel to publish messages on
type Channel struct {
	// The type of channel, e.g. "server:start"
	Type string
	// A specific key, e.g. the server ID to subscribe to. Leave empty for no key, in which case the node will receive all messages in the channel type.
	Key string
}

// This may be a generated type from protobuf definitions or something
type Message any

type EventType uint8

const (
	EventTypeMsg EventType = iota
	EventTypeAck
)

type Event struct {
	Type    EventType
	Channel Channel
	// Message will be set if the event type is EventTypeMsg, otherwise it will be nil
	Message Message
}

// Represents a message that has been published.
type PublishedMessage interface {
	// Removes the message from the database. This does NOT guarantee that the message will not have been received by a server already.
	// However, this will mean the message will not be received at an arbitary point in the future.
	Cancel()
}

type MessageHandler interface {
	Unregister()
}

// A Client represents a tolliver instance that can send and receive messages on channels that it subscribes to.
//
// A client should not expect to receive a message only once, but it should expect to receive a message AT LEAST once.
// Unless "UnreliableSend" is used, when a send method returns, it is guaranteed that the message will eventually be delivered to all current subscribers
// as long as a connection can be formed at some point.
type Client interface {
	// Creates a new connection to a server.
	//
	// Upon opening a connection the node ID and subscribed channels of this node are sent in the handshake.
	// The server will also inform the client what channels it is subscribed to and it's ID.
	NewConnection(addr ConnectionAddr)

	// Removes a server from the connection pool.
	EndConnection(addr ConnectionAddr)

	// Subscribes to a channel. When subscribing to a channel
	//  - It will be written to the database that the client is subscribed to the specific channel.
	//  - It will be broadcast to all connections that the client has now subscribed to this channel.
	//  - On every new connection created or reconnect, the subscription will be included in the data sent in the handshake.
	//
	// Due to the nature of TCP and subscriptions being sent in the handshake, it is guaranteed that subscriptions will be received and processed
	// by the remote server BEFORE message acknowledgements that happen after the subscription. This is essential for the pattern of:
	//  - Remote publishes message to start something
	//  - Client subscribes to channel for messages relating to the service that has been started
	//  - Client acknowledges processing of message
	Subscribe(channel Channel)

	// Unsubscribes from a channel. Messages will no longer be received on this channel.
	// It will be written to the database as an internal message that all nodes need to receive and only removed once it has been globally received.
	Unsubscribe(channel Channel)

	// Publish a message and waits for all servers to acknowlege the message. Returns an error if it times out (see ClientOptions.Timeout).
	// Please note that on a timeout, the message will still eventually be sent as it will have been written to the database at this point.
	// You can cancel the message but this does not guarantee it will not have already been received.
	SendAndWait(channel Channel, msg Message) (sent PublishedMessage, err error)

	// Publish a message, returns as soon as the message is written to the database as at the point that message will eventually be received and acknowledged.
	//
	// Internally this will:
	//  - Write the message to the database and get the assigned ID of the message
	//  - Start a goroutine to write the message to all servers with open connections
	//  - If there is a timeout before the message is acknowledged, try again on repeat.
	Send(channel Channel, msg Message) (sent PublishedMessage)

	// Sends a message without writing it to the database.
	UnreliableSend(channel Channel, msg Message) error

	// Registes a handler for messages in a specific channel. The "handler" function will be called everytime a message is received in that channel.
	// If "ok" is returned as true, the message will be acknowledged with the remote server, otherwise the handler will be called again to retry at increasing intervals.
	RegisterHandler(channel Channel, handler func(msg Message) (ok bool)) MessageHandler

	// Creates a channel for messages received on all channels.
	// This enables the pattern of
	//  for evt := range client.Receive() { /* ... */ }
	// After receiving and processing a message, you should call "Acknowledge(evt)".
	Receive() chan Event

	// Acknowledge the successful processing of an event so that remote clients do not have to try resending it.
	Acknowledge(evt Event)
}
