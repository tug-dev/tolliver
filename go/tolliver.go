package tolliver

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"log"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/tug-dev/tolliver/go/internal/db"
	_ "modernc.org/sqlite"
)

var InvalidInstanceOptions = errors.New("invalid instance options")

type InstanceOptions struct {
	// Addresses of remotes to connect to on creation (calls Instance.NewConnection for each)
	Remotes []RemoteAddr

	// Path to the desired database file, defaults to "./tolliver.sqlite"
	DatabasePath string

	// Reference to the desired CAs to use to authenticate remotes. This is required
	CA *x509.CertPool

	// Certificate to present to remotes during TLS handshake. This is required
	InstanceCert *tls.Certificate

	// Interface to listen on (e.g. 127.0.0.1, 0.0.0.0)
	Interface string

	// Port to listen on. If 0, will be an OS assigned port.
	Port uint16

	// Interval to try resend messages after
	RetryInterval time.Duration

	Logger *slog.Logger
}

func NewInstance(opts *InstanceOptions) (*Instance, error) {
	err := populateDefaults(opts)
	if err != nil {
		return &Instance{}, err
	}

	i := Instance{
		certs:     []tls.Certificate{*opts.InstanceCert},
		authority: opts.CA,
		logger:    opts.Logger,
	}

	i.logger.Info("Instance created")

	database, err := sql.Open("sqlite", opts.DatabasePath)
	if err != nil {
		return &Instance{}, err
	}

	i.logger.Info("Database connection made")

	i.id = db.Init(database)
	i.db = database

	i.logger.Info("Database initialized")

	i.conns = make(map[uuid.UUID]net.Conn)
	go i.retry(opts.RetryInterval)

	err = i.listenOn(opts.Interface + ":" + strconv.Itoa(int(opts.Port)))
	if err != nil {
		log.Fatal("Could not start listener: " + err.Error())
	}

	for _, r := range opts.Remotes {
		i.NewConnection(r)
	}

	return &i, nil
}

func populateDefaults(options *InstanceOptions) error {
	if options.CA == nil || options.InstanceCert == nil {
		return InvalidInstanceOptions
	}
	if options.DatabasePath == "" {
		options.DatabasePath = "./tolliver.sqlite"
	}
	if options.RetryInterval == 0 {
		options.RetryInterval = time.Second
	}

	return nil
}

type RemoteAddr struct {
	net.Addr
	ServerName string
}
