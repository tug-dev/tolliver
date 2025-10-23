package tolliver

import (
	"crypto/tls"
	"encoding/binary"
	"errors"

	"github.com/google/uuid"
)

func sendHandshake(conn *tls.Conn, id []byte, subscriptions *[]SubcriptionInfo, hostname string, port int) (connectionWrapper, error) {
	req := buildHandshakeReq(id, subscriptions)
	sendBytes(req, conn)
	res := make([]byte, 1)
	_, err := conn.Read(res)

	if err != nil {
		panic(err.Error())
	}

	err = processHandshakeRes(res, conn)
	if err != nil {
		return connectionWrapper{}, err
	}

	remoteSubscriptions := readSubscriptions(res)
	remoteId, idErr := uuid.ParseBytes(res[9:25])
	if idErr != nil {
		return connectionWrapper{}, idErr
	}

	remoteIdBytes, _ := remoteId.MarshalBinary()

	return connectionWrapper{
		Connection:    conn,
		Hostname:      hostname,
		Port:          port,
		Subscriptions: remoteSubscriptions,
		RemoteId:      remoteIdBytes,
	}, nil
}

func processHandshakeRes(res []byte, conn *tls.Conn) error {
	if res[0] != byte(uint8(HandshakeResMessageCode)) {
		return errors.New("Malformed handshake response")
	}

	switch uint8(res[9]) {
	case 2:
		return errors.New("Incompatible tolliver versions")

	case 3:
		var remoteVersion uint64
		binary.BigEndian.PutUint64(res[1:8], remoteVersion)
		if remoteVersion != TolliverVersion {
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

func buildHandshakeReq(id []byte, subscriptions *[]SubcriptionInfo) []byte {
	req := make([]byte, 1)
	req[0] = byte(uint8(HandshakeReqMessageCode))
	req = binary.BigEndian.AppendUint64(req, TolliverVersion)
	req = append(req, id...)
	req = binary.BigEndian.AppendUint32(req, uint32(len(*subscriptions)))

	for _, v := range *subscriptions {
		req = binary.BigEndian.AppendUint32(req, uint32(len(v.Channel)))
		req = binary.BigEndian.AppendUint32(req, uint32(len(v.Key)))
		req = append(req, []byte(v.Channel)...)
		req = append(req, []byte(v.Key)...)
	}

	return req
}

func readSubscriptions(res []byte) []SubcriptionInfo {
	var n uint32
	binary.BigEndian.PutUint32(res[25:29], n)
	subs := make([]SubcriptionInfo, n)

	pointer := 29

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
