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
	"github.com/tug-dev/tolliver/go/internal/handshake"
)

type Instance struct {
	certs     []tls.Certificate
	authority *x509.CertPool
	subs      []common.SubcriptionInfo
	id        uuid.UUID
	conns     []connections.Wrapper
	callbacks map[*common.SubcriptionInfo]func([]byte)
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

	inst.conns = append(inst.conns, connections.Wrapper{Subscriptions: remSubs, Id: remId, Conn: conn})
	println("Outgoing handle")
	go inst.handleConn(binary.NewReader(conn), conn, remId)
	return nil
}

func (inst *Instance) Subscribe(channel, key string) {
	inst.sendAll(buildSub(channel, key), "tolliver", "")
}

func (inst *Instance) Unsubscribe(channel, key string) {

}

func (inst *Instance) Register(channel, key string, cb func([]byte)) {

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

	inst.conns = append(inst.conns, connections.Wrapper{Subscriptions: remSubs, Id: remId, Conn: conn})
	println("Incoming handle")
	go inst.handleConn(binary.NewReader(conn), conn, remId)
}

func (inst *Instance) handleConn(r *binary.Reader, conn net.Conn, id uuid.UUID) {
	println("Handle called")
	// fmt.Printf("Handling connection to node with id %08b \n", id)
	for {
		mesType, err := r.ReadByte()
		fmt.Printf("%08b \n", mesType)
		if err != nil {
			continue
		}

		switch mesType {
		case 0:
			// TODO: Re send handshake
		case 1:
			fallthrough
		case 2:
		case 3:
			// Regular message
			inst.processRegularMessage(r, conn, id)
		case 4:
			// Ack
		}
	}
}

func (inst *Instance) processRegularMessage(r *binary.Reader, conn net.Conn, id uuid.UUID) {
	var chanLen, keyLen, bodyLen, mesId uint32
	var channel, key string
	err := r.ReadAll(nil, &chanLen, &keyLen, &mesId)
	if err != nil {
		return
	}

	err = r.ReadAll([]uint32{chanLen, keyLen}, &channel, &key)
	if err != nil {
		return
	}

	if channel == "tolliver" {
		inst.systemMessage(r, id, bodyLen)
	} else {
		body := make([]byte, bodyLen)
		r.FillBuf(body)

		for k, v := range inst.callbacks {
			if k.Channel == channel && k.Key == key {
				v(body)
			}
		}
	}

	// Send ack
	ack := buildAck(mesId)
	connections.SendBytes(ack, conn)
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

	println(channel + " " + key)

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

func buildMes(body []byte, id uint32, channel, key string) []byte {
	w := binary.NewWriter()
	w.WriteAll(byte(3), uint32(len(channel)), uint32(len(key)), uint32(len(body)), id, channel, key, body)
	return w.Join()
}

func (inst *Instance) sendAll(body []byte, channel, key string) {
	// TODO: make message durable
	id := uint32(1)
	mes := buildMes(body, id, channel, key)

	for _, v := range inst.conns {
		if channel == "tolliver" || slices.Contains(v.Subscriptions, common.SubcriptionInfo{Channel: channel, Key: key}) {
			connections.SendBytes(mes, v.Conn)
		}
	}
}
