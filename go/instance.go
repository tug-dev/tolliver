package tolliver

import (
	"crypto/tls"
	"database/sql"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	_ "modernc.org/sqlite"
)

// PUBLIC METHODS ------------------------------------------------------------------

func (inst *Instance) NewConnection(addr ConnectionAddr) error {
	// Ignore connections which have already been made.
	for _, v := range inst.ConnectionPool {
		if v.Hostname == addr.Host && v.Port == addr.Port {
			return nil
		}
	}

	// Create TLS config object for instantiating connections.
	tlsConfig := &tls.Config{
		Certificates:       inst.InstanceCertificates,
		RootCAs:            &inst.CertifcateAuthority,
		ServerName:         addr.Host,
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", addr.Host+":"+strconv.Itoa(addr.Port), tlsConfig)
	fmt.Println("Established connection to " + addr.Host + ":" + strconv.Itoa(addr.Port))
	if err != nil {
		panic(err)
	}

	inst.ConnectionPool = append(inst.ConnectionPool, ConnectionWrapper{
		Connection: conn,
		Hostname:   addr.Host,
		Port:       addr.Port,
	})

	handshakeErr := sendHandshake(conn)
	if handshakeErr != nil {
		return handshakeErr
	}

	fmt.Println("Succesful handshake")

	go handleConn(inst, conn)
	return nil
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
		Certificates:       inst.InstanceCertificates,
		RootCAs:            &inst.CertifcateAuthority,
		InsecureSkipVerify: true,
	}

	lst, err := tls.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port), tlsConfig)
	if err != nil {
		return err
	}

	go handleListener(inst, &lst, tlsConfig)

	inst.closeListener = func() {
		err = lst.Close()
		for err != nil {
			err = lst.Close()
		}
	}

	return nil
}

func handleListener(inst *Instance, lst *net.Listener, cfg *tls.Config) {
	fmt.Println("handling listener")
	for {
		conn, err := (*lst).Accept()
		if err != nil {
			continue
		}

		tlsConn := tls.Server(conn, cfg)

		go handleConn(inst, tlsConn)
	}
}

func handleConn(inst *Instance, conn *tls.Conn) {
	fmt.Println("Connection")
	conn.SetReadDeadline(time.Time{})

	for {
		buf := make([]byte, 1024)
		n, err := (*conn).Read(buf)
		if err != nil {
			continue
		}

		inst.handleMessage(buf[:n], conn)
	}
}

func (inst *Instance) handleMessage(raw []byte, conn *tls.Conn) {
	println("Message type is " + strconv.Itoa(int(raw[0])))
	println(string(raw[1:]))

	switch int(raw[0]) {
	case 0:
		println("Handshake message")
	default:
		println("Unrecognised message type")
	}
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

func sendHandshake(conn *tls.Conn) error {
	mes := make([]byte, 1)
	mes[0] = byte(HandshakeReqMessageCode)
	mes = binary.BigEndian.AppendUint64(mes, TolliverVersion)
	sendBytesOverTls(mes, conn)

	return nil
}

func sendBytesOverTls(mes []byte, conn *tls.Conn) {
	n, err := conn.Write(mes)
	var tot = n

	for err != nil || tot != len(mes) {
		n, err = conn.Write(mes[tot:])
		tot += n
	}
}

func matches(a, b SubcriptionInfo) bool {
	return (a.Channel == b.Channel || b.Channel == "") &&
		(a.Key == b.Key || b.Key == "")
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
