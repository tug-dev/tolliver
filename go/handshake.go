package tolliver

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"net"

	"github.com/google/uuid"
)

func sendHandshake(conn *tls.Conn, id uuid.UUID, subscriptions []SubcriptionInfo, hostname string, port int) (connectionWrapper, error) {
	r := Reader{bufio.NewReader(conn)}
	req := buildHandshakeReq(id, subscriptions)
	sendBytes(req, conn)

	remoteId, remoteSubscriptions, err := parseHandshakeResponse(r, conn)
	if err != nil {
		return connectionWrapper{}, err
	}

	return connectionWrapper{
		Connection:    conn,
		Hostname:      hostname,
		Port:          port,
		Subscriptions: remoteSubscriptions,
		RemoteId:      remoteId,
		R:             r,
	}, nil
}

func parseHandshakeResponse(r Reader, conn net.Conn) (uuid.UUID, []SubcriptionInfo, error) {
	var code byte
	var version uint64
	var id uuid.UUID
	var errorCode byte

	err := r.ReadAll(nil, &code, &version, &id, &errorCode)
	if err != nil {
		return uuid.UUID{}, nil, err
	}

	if code != HandshakeResMessageCode {
		return uuid.UUID{}, nil, errors.New("Malformed handshake response")
	}

	switch errorCode {
	case 0:
		fallthrough
	case 1:
		subs, err := readSubscriptions(r)
		return id, subs, err

	case 2:
		return uuid.UUID{}, nil, errors.New("Incompatible tolliver versions")

	case 3:
		if version != TolliverVersion {
			final := make([]byte, 2)
			final[0] = HandshakeFinMessageCode
			final[1] = HandshakeIncompatible
			sendBytes(final, conn)
			return uuid.UUID{}, nil, errors.New("Incompatible tolliver versions")
		} else {
			final := make([]byte, 2)
			final[0] = HandshakeFinMessageCode
			final[1] = HandshakeRequestCompatible
			sendBytes(final, conn)
			subs, err := readSubscriptions(r)
			return id, subs, err
		}

	default:
		return uuid.UUID{}, nil, errors.New("Invalid status code")
	}
}

func buildHandshakeReq(id uuid.UUID, subscriptions []SubcriptionInfo) []byte {
	req := make([]byte, 1)
	req[0] = byte(uint8(HandshakeReqMessageCode))
	req = binary.BigEndian.AppendUint64(req, TolliverVersion)
	req = append(req, id[:]...)
	req = binary.BigEndian.AppendUint32(req, uint32(len(subscriptions)))

	for _, v := range subscriptions {
		req = binary.BigEndian.AppendUint32(req, uint32(len(v.Channel)))
		req = binary.BigEndian.AppendUint32(req, uint32(len(v.Key)))
		req = append(req, []byte(v.Channel)...)
		req = append(req, []byte(v.Key)...)
	}

	return req
}

func awaitHandshake(conn net.Conn, instanceId uuid.UUID, subscriptions []SubcriptionInfo) (connectionWrapper, error) {
	r := Reader{bufio.NewReader(conn)}
	remoteID, remoteSubs, remoteVersion, err := parseHandshakeRequest(r)
	if err != nil {
		return connectionWrapper{}, err
	}

	var code byte
	if remoteVersion == TolliverVersion {
		code = 0
	}

	if remoteVersion < TolliverVersion {
		code = 2
	}

	if remoteVersion > TolliverVersion {
		code = 3
	}

	sendBytes(buildHandshakeRes(instanceId, subscriptions, code), conn)

	if code == 3 {
		err := parseHandshakeFinal(r)
		if err != nil {
			return connectionWrapper{}, err
		}
	}

	return connectionWrapper{Connection: conn, Subscriptions: remoteSubs, RemoteId: remoteID}, nil
}

func parseHandshakeFinal(r Reader) error {
	code, err := r.ReadByte()
	if err != nil {
		return err
	}
	if code != HandshakeFinMessageCode {
		return errors.New("Invalid message code")
	}

	status, err := r.ReadByte()
	if err != nil {
		return err
	}

	if status == 1 {
		return nil
	} else if status == 2 {
		return errors.New("Incompatible tolliver version")
	} else {
		return errors.New("Invalid status code")
	}
}

func parseHandshakeRequest(r Reader) (uuid.UUID, []SubcriptionInfo, uint64, error) {
	code, err := r.ReadByte()

	if err != nil {
		return uuid.UUID{}, nil, 0, err
	}
	if code != HandshakeReqMessageCode {
		return uuid.UUID{}, nil, 0, errors.New("Invalid handshake request")
	}

	version, err := r.ReadUint64()
	if err != nil {
		return uuid.UUID{}, nil, 0, err
	}

	id, err := r.ReadUUID()
	if err != nil {
		return uuid.UUID{}, nil, 0, err
	}

	subs, err := readSubscriptions(r)
	if err != nil {
		return uuid.UUID{}, nil, 0, err
	}

	return id, subs, version, nil
}

func buildHandshakeRes(instanceId uuid.UUID, subscriptions []SubcriptionInfo, code uint8) []byte {
	res := make([]byte, 1)
	res[0] = byte(uint8(HandshakeResMessageCode))
	res = binary.BigEndian.AppendUint64(res, TolliverVersion)
	res = append(res, instanceId[:]...)
	res = append(res, byte(code))
	res = binary.BigEndian.AppendUint32(res, uint32(len(subscriptions)))

	for _, v := range subscriptions {
		res = binary.BigEndian.AppendUint32(res, uint32(len(v.Channel)))
		res = binary.BigEndian.AppendUint32(res, uint32(len(v.Key)))
		res = append(res, []byte(v.Channel)...)
		res = append(res, []byte(v.Key)...)
	}

	return res
}

func readSubscriptions(r Reader) ([]SubcriptionInfo, error) {
	n, err := r.ReadUint32()
	if err != nil {
		return nil, err
	}

	subs := make([]SubcriptionInfo, n)

	for i := 0; uint32(i) < n; i++ {
		chanLen, err := r.ReadUint32()
		if err != nil {
			return nil, err
		}
		keyLen, err := r.ReadUint32()
		if err != nil {
			return nil, err
		}
		channel, err := r.ReadString(chanLen)
		if err != nil {
			return nil, err
		}
		key, err := r.ReadString(keyLen)
		if err != nil {
			return nil, err
		}

		subs[i] = SubcriptionInfo{
			Channel: channel,
			Key:     key,
		}
	}

	return subs, nil
}
