package tolliver

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tug-dev/tolliver/go/internal/binary"
	"github.com/tug-dev/tolliver/go/internal/common"
	"github.com/tug-dev/tolliver/go/internal/connections"
	"github.com/tug-dev/tolliver/go/internal/db"
	"github.com/tug-dev/tolliver/go/internal/handshake"
)

type Instance struct {
	certs     []tls.Certificate
	authority *x509.CertPool
	subs      []common.SubcriptionInfo
	id        uuid.UUID
	conns     map[uuid.UUID]net.Conn
	callbacks map[*common.SubcriptionInfo][]func([]byte)
	db        *sql.DB
	l         sync.RWMutex
}

type DialError struct {
	location *net.TCPAddr
	err      error
}

func (t *DialError) Error() string {
	return t.location.String() + " >>> " + t.err.Error()
}

func (t *DialError) Unwrap() error {
	return t.err
}

const (
	HandshakeRequestMessageCode byte = iota
	HandshakeResponseMessageCode
	HandshakeFinalMessageCode
	RegularMessageCode
	AckMessageCode
)

const (
	AckSuccess byte = iota
	AckError
)

const ReservedTolliverChannel = "tolliver"

var ErrConnAlreadyExists = errors.New("This instance already has a connection to the requested remote address")

// Attempts to create a tolliver connection to the provided address by opening a TCP socket, performing a TLS handshake
// and then a tolliver handshake
func (inst *Instance) NewConnection(addr *net.TCPAddr, tlsServerName string) error {
	opts := &tls.Config{Certificates: inst.certs, RootCAs: inst.authority, ServerName: tlsServerName, InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", addr.String(), opts)
	if err != nil {
		return &DialError{
			location: addr,
			err:      err,
		}
	}

	remId, remSubs, err := handshake.SendTolliverHandshake(conn, inst.id, inst.subs)
	if err != nil {
		return err
	}

	inst.l.RLock()
	if inst.conns[remId] != nil {
		// INFO: Still doesn't solve receiving an incoming handshake claiming an existing uuid.
		panic(ErrConnAlreadyExists)
	}
	inst.l.RUnlock()

	for _, s := range remSubs {
		db.Subscribe(s.Channel, s.Key, remId, inst.db)
	}

	inst.l.Lock()
	inst.conns[remId] = conn
	inst.l.Unlock()
	go inst.handleConn(binary.NewReader(conn), conn, remId)
	return nil
}

// Notifies all instances this instance is currently conencted to that this instance wants to receive messages
// on the provided channel and key. This is stored in memory so it can be sent during handshakes, however it is
// not persisted to the database.
//
// Passing a blank string for either channel or key acts like a * wildcard, i.e this instance will receive messages
// regardless of the destination channel, key or both
func (inst *Instance) Subscribe(channel, key string) {
	if channel == ReservedTolliverChannel {
		panic("The tolliver channel is reserved for protocol messages")
	}

	inst.subs = append(inst.subs, common.SubcriptionInfo{Channel: channel, Key: key})
	inst.send(buildSub(channel, key), ReservedTolliverChannel, "", true)
}

// Publishses to all conencted nodes that this node no longer wishes to receive messages on a given key channel pair.
// If the provided key channel pair was in the nodes subscriptions list locally it will be removed and no longer sent to
// new connections during the handshake.
//
// TODO: do we want to change the behaviour such that passing blank strings here unsubscribes from all relevant channels.
func (inst *Instance) Unsubscribe(channel, key string) {
	if channel == ReservedTolliverChannel {
		panic("Cannot unsubscribe from the reserved tolliver channel")
	}

	idx := -1
	for i, v := range inst.subs {
		if v.Channel == channel && v.Key == key {
			idx = i
			break
		}
	}

	if idx != -1 {
		inst.subs[idx] = inst.subs[len(inst.subs)-1]
		inst.subs = inst.subs[:len(inst.subs)-1]
	}

	inst.send(buildUnSub(channel, key), ReservedTolliverChannel, "", true)
}

// Registers a callback on the given key channel pair. This function will be called by tolliver any time a message is
// received on that pair. As is the case with the Subscribe method, passing blank strings for key or channel to this
// behaves like a wildcard.
func (inst *Instance) Register(channel, key string, cb func([]byte)) {
	if inst.callbacks == nil {
		inst.callbacks = make(map[*common.SubcriptionInfo][]func([]byte))
	}

	w := &common.SubcriptionInfo{Channel: channel, Key: key}

	if inst.callbacks[w] == nil {
		inst.callbacks[w] = make([]func([]byte), 0, 5)
	}

	inst.callbacks[w] = append(inst.callbacks[w], cb)
}

// Sends a message to all instances which are currently connected and subscribed on the channel key pair. Saves the message and
// required metadata to ensure eventual delivery.
func (inst *Instance) Send(channel, key string, mes []byte) {
	inst.send(mes, channel, key, true)
}

// Attempts once to send a message to all connected instances subscribed to the key channel pair.
func (inst *Instance) UnreliableSend(channel, key string, mes []byte) {
	inst.send(mes, channel, key, false)
}

// INFO: Personally I think with a sensible retry interval (10seconds +) we shouldn't need to worry about immediate resends being, and multiple deliveries is assumed by users.
func (inst *Instance) retry(interval time.Duration) {
	for {
		time.Sleep(interval)
		notAcked := db.GetWork(inst.db)

		inst.l.RLock()
		for k, v := range notAcked {
			if c := inst.conns[k]; c != nil {
				connections.SendBytes(v, c)
			}
		}
		inst.l.RUnlock()
	}
}

func (inst *Instance) listenOn(port uint16) error {
	opts := &tls.Config{Certificates: inst.certs, RootCAs: inst.authority, InsecureSkipVerify: true}
	lst, err := tls.Listen("tcp", "127.0.0.1:"+strconv.Itoa(int(port)), opts)
	if err != nil {
		return err
	}

	go connections.HandleListener(lst, inst.awaitHandshake)

	return err
}

func (inst *Instance) awaitHandshake(conn net.Conn) {
	remId, remSubs, err := handshake.AwaitHandshake(conn, inst.id, inst.subs)
	if err != nil {
		return
	}

	// TODO: How do we want to handle this. Could overwrite existing conn / have slice of conns and send to all. (Same issue as when creating connection)
	inst.l.Lock()
	if inst.conns[remId] != nil {
		return
	}

	for _, s := range remSubs {
		db.Subscribe(s.Channel, s.Key, remId, inst.db)
	}
	inst.conns[remId] = conn

	inst.l.Unlock()
	go inst.handleConn(binary.NewReader(conn), conn, remId)
}

func (inst *Instance) handleConn(r *binary.Reader, conn net.Conn, id uuid.UUID) {
	for {
		mesType, err := r.ReadByte()
		if err != nil {
			continue
		}

		switch mesType {
		case HandshakeRequestMessageCode:
			// TODO: Re send handshake maybe
		case HandshakeResponseMessageCode:
			// Handshake
		case HandshakeFinalMessageCode:
			// Handshake
		case RegularMessageCode:
			// Regular message
			inst.processRegularMessage(r, conn, id)
		case AckMessageCode:
			// Ack
			inst.proccessAck(r, id)
		}
	}
}

func (inst *Instance) proccessAck(r *binary.Reader, id uuid.UUID) {
	var status byte
	var mesId uint64
	err := r.ReadAll(nil, &status, &mesId)
	if err != nil {
		return
	}
	if status != AckSuccess || mesId == 0 {
		return
	}

	db.Ack(mesId, id, inst.db)
}

func (inst *Instance) processRegularMessage(r *binary.Reader, conn net.Conn, id uuid.UUID) {
	var mesId, chanLen uint64
	err := r.ReadAll(nil, &mesId, &chanLen)
	if err != nil {
		return
	}
	channel, err := r.ReadString(chanLen)
	if err != nil {
		return
	}
	keyLen, err := r.ReadUint64()
	if err != nil {
		return
	}
	key, err := r.ReadString(keyLen)
	if err != nil {
		return
	}
	bodyLen, err := r.ReadUint64()
	if err != nil {
		return
	}
	maxInt := uint64(int(^uint(0) >> 1))
	if bodyLen > maxInt {
		return
	}

	if channel == ReservedTolliverChannel {
		inst.systemMessage(r, id, bodyLen)
	} else {
		body := make([]byte, int(bodyLen))
		r.FillBuf(body)

		inst.l.RLock()
		for k, v := range inst.callbacks {
			if (k.Channel == channel || k.Channel == "") && (k.Key == key || k.Key == "") {
				for _, cb := range v {
					cb(body)
				}
			}
		}
		inst.l.RUnlock()
	}

	// 0 is the message ID for unreliable messages
	if mesId != 0 {
		ack := buildAck(AckSuccess, mesId)
		connections.SendBytes(ack, conn)
	}
}

func buildAck(status byte, id uint64) []byte {
	w := binary.NewWriter()
	w.WriteAll(byte(4), status, id)
	return w.Join()
}

func (inst *Instance) systemMessage(r *binary.Reader, id uuid.UUID, expectedLength uint64) {
	code, err := r.ReadByte()
	if err != nil || !(code == 0 || code == 1) {
		return
	}
	var entries []common.SubcriptionInfo
	if err := r.ReadSubs(&entries); err != nil {
		return
	}

	bytesRead := uint64(1 + 8)
	for _, entry := range entries {
		bytesRead += 8 + uint64(len(entry.Channel)) + 8 + uint64(len(entry.Key))
	}
	if bytesRead != expectedLength {
		return
	}

	for _, entry := range entries {
		if code == 0 {
			db.Subscribe(entry.Channel, entry.Key, id, inst.db)
		}
		if code == 1 {
			db.Unsubscribe(entry.Channel, entry.Key, id, inst.db)
		}
	}
}

func buildSub(channel, key string) []byte {
	w := binary.NewWriter()
	w.WriteAll(byte(0), []common.SubcriptionInfo{{Channel: channel, Key: key}})
	return w.Join()
}

func buildUnSub(channel, key string) []byte {
	w := binary.NewWriter()
	w.WriteAll(byte(1), []common.SubcriptionInfo{{Channel: channel, Key: key}})
	return w.Join()
}

func buildMes(body []byte, id uint64, channel, key string) []byte {
	w := binary.NewWriter()
	w.WriteAll(byte(3), id, uint64(len(channel)), channel, uint64(len(key)), key, uint64(len(body)), body)
	return w.Join()
}

// TODO: Not exactly sure how an iterator would fit in here
func (inst *Instance) findRecipients(channel, key string) ([]net.Conn, []uuid.UUID) {
	conns := make([]net.Conn, 0, len(inst.conns))
	ids := db.GetSubscriberUUIDs(channel, key, inst.db)

	for _, v := range ids {
		conns = append(conns, inst.conns[v])
	}

	return conns, ids
}

func (inst *Instance) send(body []byte, channel, key string, reliable bool) {
	recipientConns, recipientIds := inst.findRecipients(channel, key)

	// This represents an unreliable message
	id := uint64(0)
	if reliable {
		id = db.SaveMessage(body, recipientIds, channel, key, inst.db)
	}
	mes := buildMes(body, id, channel, key)

	for _, v := range recipientConns {
		connections.SendBytes(mes, v)
	}
}
