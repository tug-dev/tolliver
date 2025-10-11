package tolliver

import (
	"crypto/tls"
	// "crypto/x509"
	"database/sql"
	// "errors"
	_ "modernc.org/sqlite"
	"os"
	// "strconv"
	"time"
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
	// CertifcateAuthority  x509.CertPool
	ListeningPort int
	DatabasePath  string
}

func (inst *Instance) processDatabase() {
}

func (inst *Instance) listenOn(port int) {
}

// Opens the database and ensures it is initialised.
func (inst *Instance) initDatabase() {
	db, err := sql.Open("sqlite", inst.DatabasePath)
	schemaQ, schemaErr := os.ReadFile("./schema.sql")

	if schemaErr != nil {
		panic(schemaErr.Error())
	}

	if err != nil {
		panic(err.Error())
	}

	db.Exec(string(schemaQ))
}

// Connect to a new remote. If there is an issue with the arguments supplied an error will be returned with one of the following types:
//
// Unable to resolve hostname
// Remote not accepting connections on the given port
// Remote not authorized
// Tolliver handshake failed
func (inst *Instance) NewConnection(addr ConnectionAddr) error {
	// for _, v := range inst.ConnectionPool {
	// 	if v.Hostname == addr.Host && v.Port == addr.Port {
	// 		return nil
	// 	}
	// }
	//
	// tlsConfig := &tls.Config{
	// 	Certificates: inst.InstanceCertificates,
	// 	RootCAs:      &inst.CertifcateAuthority,
	// 	ServerName:   addr.Host,
	// }
	//
	// conn, err := tls.Dial("tcp", addr.Host+":"+strconv.Itoa(addr.Port), tlsConfig)
	// if err != nil {
	// 	println(err.Error())
	// 	var certErr *tls.CertificateVerificationError
	// 	if errors.As(err, &certErr) {
	// 		println("Certificate invalid")
	// 		return errors.New("Remote not authorized")
	// 	}
	// }
	//
	// inst.ConnectionPool = append(inst.ConnectionPool, ConnectionWrapper{
	// 	Connection: conn,
	// 	Hostname:   addr.Host,
	// 	Port:       addr.Port,
	// })
	//
	// connErr := make(chan error)
	// go manageConnection(inst, conn, connErr)
	//
	// return <-connErr
	return nil
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
