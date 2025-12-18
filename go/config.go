package tolliver

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
)

const InitialConnectionCapacity = 5

var InvalidInstanceOptions = errors.New("Invalid instance options")

type InstanceOptions struct {
	// Can be provided to avoid having to call NewConnection manually
	RemoteAddrs []ConnectionAddr
	// Defaults to "./tolliver.sqlite"
	DatabasePath string
	// This is the time between tolliver attempting to resend any undelivered messages in
	// miliseconds, defaults to 10_000
	RetryInterval uint
	// A reference to the certificate authority to expect to have signed certificates from
	// remotes, must be supplied
	CA *x509.Certificate
	// A reference to the certificate to provide to remotes for TLS, must be supplied
	InstanceCert *tls.Certificate
	// The port to listen for incoming connections from remotes on. Set to -1 if this
	// instance is not intended to listen for connections. Defaults to 7011
	Port int
}

func populateDefaults(options *InstanceOptions) error {
	if options.Port < 0 || options.Port > 65535 || options.CA == nil || options.InstanceCert == nil {
		return InvalidInstanceOptions
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

	return nil
}
