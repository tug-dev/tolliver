package tolliver

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"net"
)

//go:embed schema.sql
var Schema string

const InitialConnectionCapacity = 5
const TolliverVersion = 1
const (
	HandshakeReqMessageCode = 0
	HandshakeResMessageCode = 1
	HandshakeFinMessageCode = 2
	RegularMessageCode      = 3
	AckMessageCode          = 4
)

const (
	HandshakeSuccess             = 0
	HandshakeBackwardsCompatible = 1
	HandshakeIncompatible        = 2
	HandshakeRequestCompatible   = 3
)

// Creates and returns a new tolliver instance built with the supplied InstanceOptions
// Returns an error if any of the instance options are invalid
// Note that the port supplied in the options may differ from that supplied in the options
// so ensure to use the field on the Instance if you need to reference the port.
func NewInstance(options InstanceOptions) (Instance, error) {
	// Process options
	if (options.Port >= 1 && options.Port < 1024) || options.Port > 65535 || options.CA == nil || options.InstanceCert == nil {
		return Instance{}, errors.New("Invalid instance options")
	}

	if options.Port == 0 {
		options.Port = 7011
	}

	if options.RetryInterval == 0 {
		options.RetryInterval = 10_000
	}

	if options.DatabasePath == "" {
		options.DatabasePath = "./tolliver.sqlite"
	}

	certs := make([]tls.Certificate, 1)
	certs[0] = *options.InstanceCert

	rootPool := x509.NewCertPool()
	rootPool.AddCert(options.CA)

	c := Instance{
		connectionPool:       make([]connectionWrapper, InitialConnectionCapacity),
		retryInterval:        options.RetryInterval,
		instanceCertificates: certs,
		certifcateAuthority:  *rootPool,
		ListeningPort:        options.Port,
		databasePath:         options.DatabasePath,
	}

	initErr := c.initDatabase()

	if initErr != nil {
		panic(initErr.Error())
	}

	c.loadSubscriptions()

	for _, v := range options.RemoteAddrs {
		c.NewConnection(v)
	}

	if options.Port != -1 {
		err := c.listenOn(options.Port)
		if err != nil {
			fmt.Println(err.Error())
		}
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
	// Defaults to "./tolliver.sqlite"
	DatabasePath string
	// This is the time between tolliver attempting to resend any undelivered messages in miliseconds, defaults to 10_000
	RetryInterval uint
	// A reference to the certificate authority to expect to have signed certificates from remotes, must be supplied
	CA *x509.Certificate
	// A reference to the certificate to provide to remotes for TLS, must be supplied
	InstanceCert *tls.Certificate
	// The port to listen for incoming connections from remotes on. Set to -1 if this instance is not intended to listen for connections. Defaults to 7011
	Port int
}

type Message []byte

type connectionWrapper struct {
	Connection    net.Conn
	Hostname      string
	Port          int
	Subscriptions []SubcriptionInfo
	RemoteId      []byte
}

type SubcriptionInfo struct {
	Channel string
	Key     string
}

type Instance struct {
	connectionPool       []connectionWrapper
	retryInterval        uint
	instanceCertificates []tls.Certificate
	certifcateAuthority  x509.CertPool
	ListeningPort        int
	databasePath         string
	database             *sql.DB
	closeListener        func()
	subscriptions        []SubcriptionInfo
	instanceId           []byte
}
