package tolliver

import (
	"crypto/tls"
	"database/sql"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// PUBLIC METHODS ------------------------------------------------------------------

func (inst *Instance) NewConnection(addr ConnectionAddr) error {
	// Ignore connections which have already been made.
	for _, v := range inst.connectionPool {
		if v.Hostname == addr.Host && v.Port == addr.Port {
			return nil
		}
	}

	// Create TLS config object for instantiating connections.
	tlsConfig := &tls.Config{
		Certificates:       inst.instanceCertificates,
		RootCAs:            &inst.certifcateAuthority,
		ServerName:         addr.Host,
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", addr.Host+":"+strconv.Itoa(addr.Port), tlsConfig)
	if err != nil {
		panic(err)
	}

	connInfo, handshakeErr := sendHandshake(conn, inst.instanceId, &inst.subscriptions, addr.Host, addr.Port)
	if handshakeErr != nil {
		return handshakeErr
	}

	inst.connectionPool = append(inst.connectionPool, connInfo)

	go handleConn(inst, &connInfo)
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
		Certificates:       inst.instanceCertificates,
		RootCAs:            &inst.certifcateAuthority,
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
	for {
		conn, err := (*lst).Accept()
		if err != nil {
			continue
		}

		tlsConn := tls.Server(conn, cfg)

		go handleConn(inst, &connectionWrapper{Connection: tlsConn})
	}
}

func handleConn(inst *Instance, conn *connectionWrapper) {
	conn.Connection.SetReadDeadline(time.Time{})

	for {
		buf := make([]byte, 1024)
		n, err := conn.Connection.Read(buf)
		if err != nil {
			continue
		}

		inst.handleMessage(buf[:n], conn)
	}
}

func (inst *Instance) handleMessage(raw []byte, conn *connectionWrapper) {
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
	db, err := sql.Open("sqlite", inst.databasePath)
	schemaQ, schemaErr := os.ReadFile("./schema.sql")

	if schemaErr != nil {
		return schemaErr
	}

	if err != nil {
		return err
	}

	db.Exec(string(schemaQ))

	rows, qErr := db.Query("select uuid from instance")

	if qErr != nil {
		return qErr
	}

	if !rows.Next() {
		instanceId, _ := uuid.NewV7()
		idBlob, _ := instanceId.MarshalBinary()
		inst.instanceId = idBlob

		_, insErr := db.Exec("insert into instance (uuid) values (?)", idBlob)
		if insErr != nil {
			panic(insErr.Error())
		}
	} else {
		rows.Scan(inst.instanceId)
	}

	return nil
}

func (inst *Instance) loadSubscriptions() {

}

func sendBytesOverTls(mes []byte, conn *tls.Conn) {
	n, err := conn.Write(mes)
	var tot = n

	for err != nil || tot != len(mes) {
		n, err = conn.Write(mes[tot:])
		tot += n
	}
}

func matches(msgSubscription, remoteSubscription SubcriptionInfo) bool {
	if msgSubscription.Channel == "tolliver" {
		return true
	}

	return (msgSubscription.Channel == remoteSubscription.Channel || remoteSubscription.Channel == "") &&
		(msgSubscription.Key == remoteSubscription.Key || remoteSubscription.Key == "")
}
