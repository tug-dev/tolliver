package handshake

import (
	"net"

	"github.com/google/uuid"
	"github.com/tug-dev/tolliver/go/internal/binary"
	"github.com/tug-dev/tolliver/go/internal/common"
	"github.com/tug-dev/tolliver/go/internal/connections"
)

type handshakeReq struct {
	Version uint64
	Id      uuid.UUID
	Subs    []common.SubcriptionInfo
}

func AwaitHandshake(conn net.Conn, instanceId uuid.UUID, subscriptions []common.SubcriptionInfo) (uuid.UUID, []common.SubcriptionInfo, error) {
	r := binary.NewReader(conn)
	req, err := parseHandshakeRequest(r)
	if err != nil {
		return uuid.UUID{}, nil, err
	}

	code := HandshakeSuccess
	// Currently backwards compatibility is not supported
	if req.Version < common.TolliverVersion {
		code = HandshakeIncompatible
	}

	// Request to see if new version is compatible
	if req.Version > common.TolliverVersion {
		code = HandshakeRequestCompatible
	}

	connections.SendBytes(buildHandshakeRes(instanceId, subscriptions, code), conn)

	if code == HandshakeRequestCompatible {
		err := parseHandshakeFinal(r)
		if err != nil {
			return uuid.UUID{}, nil, err
		}
	}
	if code == HandshakeIncompatible {
		return req.Id, req.Subs, IncompatibleVersions
	}

	return req.Id, req.Subs, nil
}

func parseHandshakeRequest(r *binary.Reader) (handshakeReq, error) {
	var code byte
	var version uint64
	var id uuid.UUID
	var subs []common.SubcriptionInfo

	err := r.ReadAll(nil, &code, &version, &id)
	if err != nil {
		return handshakeReq{}, err
	}

	if code != HandshakeReqMessageCode {
		return handshakeReq{}, UnexpectedMessageCode
	}
	if err := r.ReadSubs(&subs); err != nil {
		return handshakeReq{}, err
	}

	return handshakeReq{Version: version, Id: id, Subs: subs}, nil
}

func buildHandshakeRes(id uuid.UUID, subscriptions []common.SubcriptionInfo, code byte) []byte {
	w := binary.NewWriter()
	w.WriteAll(HandshakeResMessageCode, common.TolliverVersion, id, code, subscriptions)

	return w.Join()
}

func parseHandshakeFinal(r *binary.Reader) error {
	var code, status byte
	err := r.ReadAll(nil, &code, &status)
	if err != nil {
		return err
	}
	if code != HandshakeFinMessageCode {
		return UnexpectedMessageCode
	}
	if status == HandshakeFinalSuccess {
		return nil
	}
	if status == HandshakeFinalIncompatible {
		return IncompatibleVersions
	}
	return HandshakeFailed
}
