package tolliver

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"time"
)

const InitialConnectionCapacity = 5
const TolliverVersion = 1
const (
	HandshakeMessageCode       = 0
	ChannelSubcriptionCode     = 1
	SubscriptionResponseCode   = 2
	ChannelUnsubscribeCode     = 3
	UnsubscriptionResponseCode = 4
)

// Creates and returns a new tolliver instance built with the supplied InstanceOptions
// Returns an error if any of the instance options are invalid
// Note that the port supplied in the options may differ from that supplied in the options
// so ensure to use the field on the Instance if you need to reference the port.
func NewInstance(options InstanceOptions) (Instance, error) {
	certs := make([]tls.Certificate, 1)
	certs[0] = *options.InstanceCert

	rootPool := x509.NewCertPool()
	rootPool.AddCert(options.CA)

	c := Instance{
		make([]ConnectionWrapper, InitialConnectionCapacity),
		options.RetryInterval,
		certs,
		*rootPool,
		options.Port,
		options.DatabasePath,
		func() {},
	}

	c.initDatabase()

	for _, v := range options.RemoteAddrs {
		c.NewConnection(v)
	}

	if (options.Port != -1 && options.Port < 1024) || options.Port > 65535 {
		return c, errors.New("Invalid instance options")
	}

	if options.Port != -1 {
		c.listenOn(options.Port)
	}

	return c, nil
}

// Defines the details for connecting to a remote
type ConnectionAddr struct {
	Host string
	Port int
}

// The options for creating a new instance
type InstanceOptions struct {
	// Can be provided to avoid having to call NewConnection manually
	RemoteAddrs []ConnectionAddr
	// Leaving as the empty string defaults to "./tolliver.sqlite"
	DatabasePath string
	// This is the time between tolliver attempting to resend any undelivered messages in miliseconds
	RetryInterval int
	// A reference to the certificate authority to expect to have signed certificates from remotes
	CA *x509.Certificate
	// A reference to the certificate to provide to remotes for TLS
	InstanceCert *tls.Certificate
	// The port to listen for incoming connections from remotes on. Set to -1 if this instance is not intended to listen for connections.
	Port int
}

type Message []byte

type ConnectionWrapper struct {
	Connection    *tls.Conn
	Hostname      string
	Port          int
	Subscriptions []SubcriptionInfo
}

type SubcriptionInfo struct {
	Channel string
	Key     string
}

type address struct {
	Hostname string
	Port     int
}

type Instance struct {
	ConnectionPool       []ConnectionWrapper
	RetryInterval        int
	InstanceCertificates []tls.Certificate
	CertifcateAuthority  x509.CertPool
	ListeningPort        int
	DatabasePath         string
	closeListener        func()
}
