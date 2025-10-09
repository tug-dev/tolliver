package tolliver

import (
	"crypto/tls"
	"crypto/x509"
	"strconv"
)

type ConnectionWrapper struct {
	Connection *tls.Conn
	Hostname   string
	Port       int
}

type Client struct {
	ConnectionPool []ConnectionWrapper
}

func (c *Client) NewConnection(addr ConnectionAddr) {
	caPool := x509.NewCertPool()
	caPool.AddCert(addr.CA)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*addr.ClientCert},
		RootCAs:      caPool,
		ServerName:   addr.Host,
	}

	conn, err := tls.Dial("tcp", addr.Host+":"+strconv.Itoa(addr.Port), tlsConfig)
	if err != nil {
		panic(err.Error())
	}

	c.ConnectionPool = append(c.ConnectionPool, ConnectionWrapper{
		Connection: conn,
		Hostname:   addr.Host,
		Port:       addr.Port,
	})
}

func (c *Client) EndConnection(addr ConnectionAddr) {
	for i, v := range c.ConnectionPool {
		if v.Hostname == addr.Host && v.Port == addr.Port {
			// Potentially need to do some tear down here
			v.Connection.Close()

			c.ConnectionPool[i] = c.ConnectionPool[len(c.ConnectionPool)-1]
			c.ConnectionPool = c.ConnectionPool[:len(c.ConnectionPool)-1]
			return
		}
	}
}

func (c *Client) Subscribe(channel Channel) {

}

func (c *Client) Unsubscribe(channel Channel) {

}
func (c *Client) SendAndWait(channel Channel, msg Message) (sent PublishedMessage, err error) {
	return nil, nil
}

func (c *Client) Send(channel Channel, msg Message) (sent PublishedMessage) {
	return nil
}

func (c *Client) UnreliableSend(channel Channel, msg Message) error {
	return nil
}

func (c *Client) RegisterHandler(channel Channel, handler func(msg Message) (ok bool)) MessageHandler {
	return nil
}

func (c *Client) Receive() chan Event {
	return nil
}

func (c *Client) Acknowledge(evt Event) {

}

// A Client represents a tolliver instance that can send and receive messages on channels that it subscribes to.
//
// A client should not expect to receive a message only once, but it should expect to receive a message AT LEAST once.
// Unless "UnreliableSend" is used, when a send method returns, it is guaranteed that the message will eventually be delivered to all current subscribers
// as long as a connection can be formed at some point.
type Clinterface interface {
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
