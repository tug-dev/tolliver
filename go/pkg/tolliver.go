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
func NewClient(options ClientOptions)

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
