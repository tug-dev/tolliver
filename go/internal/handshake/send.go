package handshake

import (
	"crypto/tls"

	"github.com/google/uuid"
	"github.com/tug-dev/tolliver/go/internal/binary"
	"github.com/tug-dev/tolliver/go/internal/common"
	"github.com/tug-dev/tolliver/go/internal/connections"
)

type handshakeRes struct {
	Status  byte
	Version uint64
	Id      uuid.UUID
	Subs    []common.SubcriptionInfo
}

func SendTolliverHandshake(conn *tls.Conn, id uuid.UUID, subscriptions []common.SubcriptionInfo) (uuid.UUID, []common.SubcriptionInfo, error) {
	r := binary.NewReader(conn)
	req := buildHandshakeReq(id, subscriptions)
	connections.SendBytes(req, conn)

	res, err := parseHandshakeResponse(r)
	if err != nil {
		return uuid.UUID{}, nil, err
	}

	switch res.Status {
	case HandshakeSuccess:
		fallthrough
	case HandshakeBackwardsCompatible:
		return res.Id, res.Subs, nil

	case HandshakeRequestCompatible:
		fin := buildHandshakeFin(HandshakeIncompatible)
		connections.SendBytes(fin, conn)
		fallthrough
	case HandshakeIncompatible:
		return res.Id, res.Subs, IncompatibleVersions
	default:
		return res.Id, res.Subs, UnexpectedMessageCode
	}
}

func buildHandshakeReq(id uuid.UUID, subscriptions []common.SubcriptionInfo) []byte {
	w := binary.NewWriter()
	w.WriteAll(HandshakeReqMessageCode, common.TolliverVersion, id, uint32(len(subscriptions)), subscriptions)

	return w.Join()
}

func buildHandshakeFin(code byte) []byte {
	w := binary.NewWriter()
	w.WriteAll(HandshakeFinMessageCode, code)

	return w.Join()
}

func parseHandshakeResponse(r *binary.Reader) (handshakeRes, error) {
	var code byte
	var version uint64
	var id uuid.UUID
	var errorCode byte
	var subs []common.SubcriptionInfo

	err := r.ReadAll(nil, &code, &version, &id, &errorCode, subs)
	if err != nil {
		return handshakeRes{}, err
	}
	if code != HandshakeResMessageCode {
		return handshakeRes{}, UnexpectedMessageCode
	}
	return handshakeRes{Status: errorCode, Version: version, Id: id, Subs: subs}, nil
}
