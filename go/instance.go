package tolliver

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"net"
	"os"
	"strconv"
	"time"

	_ "modernc.org/sqlite"
)

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
	Timeout              time.Time
	InstanceCertificates []tls.Certificate
	CertifcateAuthority  x509.CertPool
	ListeningPort        int
	DatabasePath         string
	closeListener        func()
}

func (inst *Instance) processDatabase() {
}

func (inst *Instance) listenOn(port int) error {
	tlsConfig := &tls.Config{
		Certificates: inst.InstanceCertificates,
		RootCAs:      &inst.CertifcateAuthority,
	}

	lst, err := tls.Listen("tcp", strconv.Itoa(port), tlsConfig)
	if err != nil {
		return err
	}

	go handleListener(inst, &lst)

	inst.closeListener = func() {
		err = lst.Close()
		for err != nil {
			err = lst.Close()
		}
	}

	return nil
}

func handleListener(inst *Instance, lst *net.Listener) {
	for {
		conn, err := (*lst).Accept()
		if err != nil {
			continue
		}

		go handleConn(inst, &conn)
	}
}

func handleConn(inst *Instance, conn *net.Conn) {
	(*conn).SetReadDeadline(time.Time{})

	for {
		buf := make([]byte, 1024)
		n, err := (*conn).Read(buf)
		if err != nil {
			continue
		}

		inst.handleMessage(buf[:n])
	}
}

func (inst *Instance) handleMessage(raw []byte) {
	println(string(raw))
}

// Opens the database and ensures it is initialised.
func (inst *Instance) initDatabase() error {
	db, err := sql.Open("sqlite", inst.DatabasePath)
	schemaQ, schemaErr := os.ReadFile("./schema.sql")

	if schemaErr != nil {
		return schemaErr
	}

	if err != nil {
		return err
	}

	db.Exec(string(schemaQ))
	return nil
}

// Connect to a new remote. If there is an issue with the arguments supplied an error will be returned with one of the following types:
//
// Unable to resolve hostname
// Remote not accepting connections on the given port
// Remote not authorized
// Tolliver handshake failed
func (inst *Instance) NewConnection(addr ConnectionAddr) error {
	for _, v := range inst.ConnectionPool {
		if v.Hostname == addr.Host && v.Port == addr.Port {
			return nil
		}
	}

	tlsConfig := &tls.Config{
		Certificates: inst.InstanceCertificates,
		RootCAs:      &inst.CertifcateAuthority,
		ServerName:   addr.Host,
	}

	conn, err := tls.Dial("tcp", addr.Host+":"+strconv.Itoa(addr.Port), tlsConfig)
	if err != nil {
		println(err.Error())
		var certErr *tls.CertificateVerificationError
		if errors.As(err, &certErr) {
			println("Certificate invalid")
			return errors.New("Remote not authorized")
		}
	}

	inst.ConnectionPool = append(inst.ConnectionPool, ConnectionWrapper{
		Connection: conn,
		Hostname:   addr.Host,
		Port:       addr.Port,
	})

	connErr := make(chan error)
	go manageConnection(inst, conn, connErr)

	return <-connErr
}

// Write the message to the database and all of the intended recipients, then attempt to send the message.
func (inst *Instance) Send(msg Message, channel, key string) {

}

func matches(a, b SubcriptionInfo) bool {
	return (a.Channel == b.Channel || b.Channel == "") &&
		(a.Key == b.Key || b.Key == "")
}

func (inst *Instance) UnreliableSend(msg Message, channel, key string) {

}

// Performs the tolliver handshake over the connection then waits for messages and passes these to the
// instance when they are received.
func manageConnection(inst *Instance, conn *tls.Conn, errChan chan error) {
	for {
		if conn.ConnectionState().HandshakeComplete {
			buf := make([]byte, 0)
			_, err := conn.Read(buf)
			// We can ignore any errors here as the sender will resend the message if no ack is received
			if err != nil {
				continue
			}
			// c.onReceived()
		}
	}
}

// Closes all connections of the instace. Note this method does not clear subscriptions
// nor unsent messages. It is simply to be used to free resources gracefully once the
// instance is no longer required.
func (inst *Instance) Close() {

}
