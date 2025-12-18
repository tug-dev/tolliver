package tolliver

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	_ "embed"
	"fmt"
	"net"
	"strconv"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type Instance struct {
	ConnectionPool       []connectionWrapper
	RetryInterval        uint
	InstanceCertificates []tls.Certificate
	CertifcateAuthority  x509.CertPool
	ListeningPort        int
	DatabasePath         string
	Database             *sql.DB
	CloseListener        func()
	Subscriptions        []SubcriptionInfo
	InstanceId           uuid.UUID
}

type ConnectionAddr struct {
	Host string
	Port int
}

type Message []byte

type connectionWrapper struct {
	Connection    net.Conn
	Hostname      string
	Port          int
	Subscriptions []SubcriptionInfo
	RemoteId      uuid.UUID
	R             Reader
}

//go:embed schema.sql
var Schema string

const (
	HandshakeReqMessageCode uint8 = iota
	HandshakeResMessageCode
	HandshakeFinMessageCode
	RegularMessageCode
	AckMessageCode
)

var (
	ConnectionAlreadyExists = errors.New("Did not create a connection to the specified remote as a connection ")
)

type TLSError struct {
	Location string
	Err      error
}

func (e *TLSError) Unwrap() error {
	return e.Err
}

func (e *TLSError) Error() string {
	return e.Location + ": " + e.Err.Error()
}

// PUBLIC METHODS ------------------------------------------------------------------

// Opens a TLS connection to the provided remote address. Attempts to complete a
// TLS handshake, then a Tolliver handshake. If both are successful this function
// creates a goroutine to listen for messages on the connection. Possible errors:
// ConnectionAlreadyExists
// TLSError
func (inst *Instance) NewConnection(addr common.ConnectionAddr) error {
	// Ignore connections which have already been made.
	for _, v := range inst.ConnectionPool {
		if addressesEqual(v, addr) {
			return common.ConnectionAlreadyExists
		}
	}

	// Create TLS config object for instantiating connections.
	tlsConfig := &tls.Config{
		Certificates: inst.InstanceCertificates,
		RootCAs:      &inst.CertifcateAuthority,
		// ServerName:   addr.Host,
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", addr.Host+":"+strconv.Itoa(addr.Port), tlsConfig)
	if err != nil {
		return &common.TLSError{
			Location: "Creating connection to remote",
			Err:      err,
		}
	}

	connInfo, handshakeErr := utils.SendHandshake(conn, inst.InstanceId, inst.Subscriptions, addr.Host, addr.Port)
	if handshakeErr != nil {
		panic(handshakeErr)
	}

	inst.connectionPool = append(inst.connectionPool, connInfo)

	go handleConn(inst, &connInfo)
	return nil
}

func addressesEqual(c connectionWrapper, a ConnectionAddr) bool {
	return c.Hostname == a.Host && c.Port == a.Port
}

// Write the message to the database and all of the intended recipients, then attempt to send the message.
func (inst *Instance) Send(msg Message, channel, key string) {

}

func (inst *Instance) UnreliableSend(msg Message, channel, key string) {

}

func (inst *Instance) Subscribe(channel, key string) {

}

func (inst *Instance) Unsubscribe(channel, key string) {

}

func (inst *Instance) RegisterCallback(cb func(Message), channel, key string) {

}

// INTERNAL METHODS -----------------------------------------------------------------

func (inst *Instance) processDatabase() {
}

func (inst *Instance) listenOn(port int) error {
	tlsConfig := &tls.Config{
		Certificates:       inst.instanceCertificates,
		RootCAs:            &inst.certifcateAuthority,
		InsecureSkipVerify: true,
	}

	lst, err := tls.Listen("tcp", ":"+strconv.Itoa(port), tlsConfig)
	if err != nil {
		return err
	}

	go handleListener(inst, lst)

	inst.closeListener = func() {
		err = lst.Close()
		for err != nil {
			err = lst.Close()
		}
	}

	return nil
}

func handleListener(inst *Instance, lst net.Listener) {
	for {
		conn, err := lst.Accept()
		if err != nil {
			continue
		}
		println("Accepted connection")

		connWrap, handshakeErr := awaitHandshake(conn, inst.instanceId, inst.subscriptions)
		if handshakeErr != nil {
			println(err)
			continue
		}

		println("Finished handshake")

		go handleConn(inst, &connWrap)
	}
}

func handleConn(inst *Instance, conn *connectionWrapper) {
	for {
		buf := make([]byte, 1024)
		n, err := conn.Connection.Read(buf)
		if err != nil {
			break
		}

		inst.handleMessage(buf[:n], conn)
	}
}

func (inst *Instance) handleMessage(raw []byte, conn *connectionWrapper) {
	fmt.Printf("Messaage from %08b\n", conn.RemoteId)
	fmt.Printf("%08b\n", raw)

}

// Opens the database and ensures it is initialised.
func (inst *Instance) initDatabase() error {
	db, err := sql.Open("sqlite", inst.databasePath)

	if err != nil {
		return err
	}

	db.Exec(string(Schema))

	rows, qErr := db.Query("select uuid from instance")

	if qErr != nil {
		return qErr
	}

	if !rows.Next() {
		instanceId, _ := uuid.NewV7()
		idBlob, _ := instanceId.MarshalBinary()
		inst.instanceId = instanceId

		_, insErr := db.Exec("insert into instance (uuid) values (?)", idBlob)
		if insErr != nil {
			panic(insErr.Error())
		}
	} else {
		rows.Scan(&inst.instanceId)
	}

	return nil
}

func (inst *Instance) loadSubscriptions() {

}
