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

type TLSError struct {
	location string
	err      error
}

func (t *TLSError) Error() string {
	return t.location + ": " + t.err.Error()
}

func (t *TLSError) Unwrap() error {
	return t.err
}

var ConnectionAlreadyExists = errors.New("This instance already has a connection to the requested remote address")

func (inst *Instance) NewConnection(addr net.TCPAddr, tlsServerName string) error {
	opts := &tls.Config{Certificates: inst.certs, RootCAs: inst.authority, ServerName: tlsServerName, InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", addr.String(), opts)
	if err != nil {
		return &TLSError{
			location: "Creating connection to " + addr.String(),
			err:      err,
		}
	}

	remId, remSubs, err := handshake.SendTolliverHandshake(conn, inst.id, inst.subs)
	if err != nil {
		return err
	}

	// TODO: What should happen here
	inst.l.RLock()
	if inst.conns[remId] != nil {
		return nil
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

func (inst *Instance) Subscribe(channel, key string) {
	if channel == "tolliver" {
		return
	}

	inst.send(buildSub(channel, key), "tolliver", "", true)
}

func (inst *Instance) Unsubscribe(channel, key string) {
	if channel == "tolliver" {
		return
	}

	inst.send(buildUnSub(channel, key), "tolliver", "", true)
}

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

func (inst *Instance) Send(channel, key string, mes []byte) {
	inst.send(mes, channel, key, true)
}

func (inst *Instance) UnreliableSend(channel, key string, mes []byte) {
	inst.send(mes, channel, key, false)
}

func (inst *Instance) retry(interval time.Duration) {
	for {
		time.Sleep(interval)
		inst.l.Lock()
		notAcked := db.GetWork(inst.db)
		inst.l.Unlock()

		inst.l.RLock()
		for k, v := range notAcked {
			if inst.conns[k] != nil {
				connections.SendBytes(v, inst.conns[k])
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
	inst.l.RLock()
	if inst.conns[remId] != nil {
		return
	}
	inst.l.RUnlock()

	for _, s := range remSubs {
		db.Subscribe(s.Channel, s.Key, remId, inst.db)
	}

	inst.l.Lock()
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
		case 0:
			// TODO: Re send handshake maybe
		case 1:
			// Handshake
		case 2:
			// Handshake
		case 3:
			// Regular message
			inst.processRegularMessage(r, conn, id)
		case 4:
			// Ack
			inst.proccessAck(r, id)
		}
	}
}

func (inst *Instance) proccessAck(r *binary.Reader, id uuid.UUID) {
	mesId, err := r.ReadUint32()
	if err != nil {
		return
	}

	db.Ack(mesId, id, inst.db)
}

func (inst *Instance) processRegularMessage(r *binary.Reader, conn net.Conn, id uuid.UUID) {
	var chanLen, keyLen, bodyLen, mesId uint32
	var channel, key string
	err := r.ReadAll(nil, &chanLen, &keyLen, &bodyLen, &mesId)
	if err != nil {
		return
	}

	err = r.ReadAll([]uint32{chanLen, keyLen}, &channel, &key)
	if err != nil {
	}

	if channel == "tolliver" {
		inst.systemMessage(r, id, bodyLen)
	} else {
		body := make([]byte, bodyLen)
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

	if mesId != uint32(4294967295) {
		// Send ack
		ack := buildAck(mesId)
		connections.SendBytes(ack, conn)
	}
}

func buildAck(id uint32) []byte {
	w := binary.NewWriter()
	w.WriteAll(byte(4), id)
	return w.Join()
}

func (inst *Instance) systemMessage(r *binary.Reader, id uuid.UUID, expectedLength uint32) {
	var code byte
	var chanLen, keyLen uint32
	err := r.ReadAll(nil, &code, &chanLen, &keyLen)
	if err != nil || chanLen+keyLen+9 != expectedLength || !(code == 0 || code == 1) {
		return
	}

	var channel, key string
	err = r.ReadAll([]uint32{chanLen, keyLen}, &channel, &key)
	if err != nil {
		return
	}

	if code == 0 {
		db.Subscribe(channel, key, id, inst.db)
	}
	if code == 1 {
		db.Unsubscribe(channel, key, id, inst.db)
	}
}

func buildSub(channel, key string) []byte {
	w := binary.NewWriter()
	w.WriteAll(byte(0), uint32(len(channel)), uint32(len(key)), channel, key)
	return w.Join()
}

func buildUnSub(channel, key string) []byte {
	w := binary.NewWriter()
	w.WriteAll(byte(1), uint32(len(channel)), uint32(len(key)), channel, key)
	return w.Join()
}

func buildMes(body []byte, id uint32, channel, key string) []byte {
	w := binary.NewWriter()
	w.WriteAll(byte(3), uint32(len(channel)), uint32(len(key)), uint32(len(body)), id, channel, key, body)
	return w.Join()
}

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

	id := uint32(4294967295)
	if reliable {
		id = db.SaveMessage(body, recipientIds, channel, key, inst.db)
	}
	mes := buildMes(body, id, channel, key)

	for _, v := range recipientConns {
		connections.SendBytes(mes, v)
	}
}
