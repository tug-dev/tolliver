package tolliver

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/tug-dev/tolliver/go/internal/db"
	_ "modernc.org/sqlite"
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

	// Interval to try resend messages after
	RetryInterval time.Duration
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

	database, err := sql.Open("sqlite", opts.DatabasePath)
	if err != nil {
		return &Instance{}, err
	}

	i.id = db.Init(database)
	i.db = database

	i.conns = make(map[uuid.UUID]net.Conn)
	i.listenOn(opts.Port)
	go i.retry(opts.RetryInterval)

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
	if options.RetryInterval == 0 {
		options.RetryInterval = time.Second
	}

	return nil
}
