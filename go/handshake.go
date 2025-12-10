package tolliver

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"go/version"
	"net"
)

func sendHandshake(conn *tls.Conn, id []byte, subscriptions []SubcriptionInfo, hostname string, port int) (connectionWrapper, error) {
	r := Reader{bufio.NewReader(conn)}
	req := buildHandshakeReq(id, subscriptions)
	sendBytes(req, conn)

	remoteId, remoteSubscriptions, err := parseHandshakeResponse(r)
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

func parseHandshakeResponse(r Reader, conn net.Conn) error {
	code, err := r.ReadByte()

	if err != nil {
		return err
	}
	if code != HandshakeResMessageCode {
		return errors.New("Malformed handshake response")
	}

	version, err := r.ReadUint64()
	if err != nil {
		return err
	}

	errorCode, err := r.ReadByte()

	switch errorCode {
	case 2:
		return errors.New("Incompatible tolliver versions")

	case 3:
		if version != TolliverVersion {
			final := make([]byte, 2)
			final[0] = byte(uint8(HandshakeFinMessageCode))
			final[1] = byte(uint8(HandshakeIncompatible))
			sendBytes(final, conn)
			return errors.New("Incompatible tolliver versions")
		} else {
			final := make([]byte, 2)
			final[0] = byte(uint8(HandshakeFinMessageCode))
			final[1] = byte(uint8(HandshakeRequestCompatible))
			sendBytes(final, conn)
		}
	}

	return nil
}

func buildHandshakeReq(id []byte, subscriptions []SubcriptionInfo) []byte {
	req := make([]byte, 1)
	req[0] = byte(uint8(HandshakeReqMessageCode))
	req = binary.BigEndian.AppendUint64(req, TolliverVersion)
	req = append(req, id...)
	req = binary.BigEndian.AppendUint32(req, uint32(len(subscriptions)))

	for _, v := range subscriptions {
		req = binary.BigEndian.AppendUint32(req, uint32(len(v.Channel)))
		req = binary.BigEndian.AppendUint32(req, uint32(len(v.Key)))
		req = append(req, []byte(v.Channel)...)
		req = append(req, []byte(v.Key)...)
	}

	return req
}

func parseHandshakeRequest(r Reader) ([]byte, []SubcriptionInfo, error) {
	remoteID := req[9:25]
	remoteSubs := readSubscriptions(req, 25)

	return remoteID, remoteSubs, 0
}

func awaitHandshake(conn net.Conn, instanceId []byte, subscriptions []SubcriptionInfo) (connectionWrapper, error) {
	req := make([]byte, 32)
	_, err := conn.Read(req)
	remoteID, remoteSubs := parseHandshakeRequest(req)
	if err != nil {
		return connectionWrapper{}, err
	}

	conn.Write(buildHandshakeRes(instanceId, subscriptions, 0))

	return connectionWrapper{Connection: conn, Subscriptions: remoteSubs, RemoteId: remoteID}, nil
}

func buildHandshakeRes(instanceId []byte, subscriptions []SubcriptionInfo, code uint8) []byte {
	res := make([]byte, 1)
	res[0] = byte(uint8(HandshakeResMessageCode))
	res = binary.BigEndian.AppendUint64(res, TolliverVersion)
	res = append(res, instanceId...)
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

func readSubscriptions(res []byte, start int) []SubcriptionInfo {
	var n uint32
	binary.BigEndian.PutUint32(res[start:start+4], n)
	subs := make([]SubcriptionInfo, n)

	pointer := start + 4

	for i := 0; uint32(i) < n; i++ {
		var chanLen, keyLen uint32
		binary.BigEndian.PutUint32(res[pointer:pointer+4], chanLen)
		binary.BigEndian.PutUint32(res[pointer+4:pointer+8], keyLen)

		subs[i] = SubcriptionInfo{
			Channel: string(res[pointer+8 : pointer+8+int(chanLen)]),
			Key:     string(res[pointer+8+int(chanLen) : pointer+8+int(chanLen)+int(keyLen)]),
		}
		pointer += 8 + int(chanLen) + int(keyLen)
	}

	return subs
}
