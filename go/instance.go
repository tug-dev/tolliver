package tolliver

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"slices"
	"strconv"

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
	conns     []*connections.Wrapper
	callbacks map[*common.SubcriptionInfo][]func([]byte)
	dbPath    string
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

// PUBLIC METHODS ------------------------------------------------------------------
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

	for _, v := range inst.conns {
		if slices.Equal(remId[:], v.Id[:]) {
			return nil
		}
	}

	inst.conns = append(inst.conns, &connections.Wrapper{Subscriptions: remSubs, Id: remId, Conn: conn})
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

	// TODO: How do we want to handle this. Could overwrite existing conn / have slice of conns and send to all.
	for _, v := range inst.conns {
		if slices.Equal(remId[:], v.Id[:]) {
			return
		}
	}

	inst.conns = append(inst.conns, &connections.Wrapper{Subscriptions: remSubs, Id: remId, Conn: conn})
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
			// TODO: Re send handshake
		case 1:
			// Handshake
		case 2:
			// Handshake
		case 3:
			// Regular message
			inst.processRegularMessage(r, conn, id)
		case 4:
			inst.proccessAck(r, id)
			// ack
		}
	}
}

func (inst *Instance) proccessAck(r *binary.Reader, id uuid.UUID) {
	mesId, err := r.ReadUint32()
	if err != nil {
		return
	}

	db.Ack(mesId, id, inst.dbPath)
}

func (inst *Instance) Debug() {
	fmt.Printf("Tolliver instance with ID % x\n", inst.id[:])
	fmt.Printf("Connections (%d):\n", len(inst.conns))
	for i, v := range inst.conns {
		fmt.Printf("(%d) Remote ID % x, Subs: ", i, v.Id[:])
		for _, s := range v.Subscriptions {
			fmt.Printf("(%s, %s), ", s.Channel, s.Key)
		}
		fmt.Printf("\n")
	}
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

		for k, v := range inst.callbacks {
			if k.Channel == channel && k.Key == key {
				for _, cb := range v {
					cb(body)
				}
			}
		}
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

	for _, v := range inst.conns {
		if slices.Equal(v.Id[:], id[:]) {
			if code == 0 {
				v.Subscriptions = append(v.Subscriptions, common.SubcriptionInfo{Channel: channel, Key: key})
			}
			if code == 1 {
				idx := -1
				for i, s := range v.Subscriptions {
					if s.Channel == channel && s.Key == key {
						idx = i
						break
					}
				}

				if idx != -1 {
					v.Subscriptions[idx] = v.Subscriptions[len(v.Subscriptions)-1]
					v.Subscriptions = v.Subscriptions[:len(v.Subscriptions)-1]
				}
			}
			return
		}
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
	ids := make([]uuid.UUID, 0, len(inst.conns))

	for _, v := range inst.conns {
		if channel == "tolliver" || slices.Contains(v.Subscriptions, common.SubcriptionInfo{Channel: channel, Key: key}) {
			conns = append(conns, v.Conn)
			ids = append(ids, v.Id)
		}
	}

	return conns, ids
}

func (inst *Instance) send(body []byte, channel, key string, reliable bool) {
	recipientConns, recipientIds := inst.findRecipients(channel, key)

	id := uint32(4294967295)
	if reliable {
		id = db.SaveMessage(body, recipientIds, channel, key, inst.dbPath)
	}
	mes := buildMes(body, id, channel, key)

	for _, v := range recipientConns {
		connections.SendBytes(mes, v)
	}
}
