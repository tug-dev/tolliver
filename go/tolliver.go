package tolliver

import (
	"crypto/tls"
	"crypto/x509"
	"errors"

	"github.com/tug-dev/tolliver/go/internal/connections"
)

var InvalidInstanceOptions = errors.New("Invalid instance options")

type InstanceOptions struct {
	// Addresses of remotes to connect to on creation (calls Instance.NewConnection for each)
	RemoteAddrs []string

	// Path to the desired database file, defaults to "./tolliver.sqlite"
	DatabasePath string

	// Reference to the desired CAs to use to authenticate remotes. This is required
	CA *x509.CertPool

	// Certificate to present to remotes during TLS handshake. This is required
	InstanceCert *tls.Certificate

	// Port to listen on. Defaults to 7011 (T O L L)
	Port uint16
}

func NewInstance(opts *InstanceOptions) (*Instance, error) {
	err := populateDefaults(opts)
	if err != nil {
		return &Instance{}, err
	}

	i := Instance{
		certs:     []tls.Certificate{*opts.InstanceCert},
		authority: opts.CA,
	}
	i.conns = make([]*connections.Wrapper, 0, 10)
	i.listenOn(opts.Port)

	return &i, nil
}

func populateDefaults(options *InstanceOptions) error {
	if options.CA == nil || options.InstanceCert == nil {
		return InvalidInstanceOptions
	}
	if options.Port == 0 {
		options.Port = 7011
	}
	if options.DatabasePath == "" {
		options.DatabasePath = "./tolliver.sqlite"
	}

	return nil
}
