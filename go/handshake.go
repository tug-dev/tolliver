package tolliver

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"net"

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

	remoteSubscriptions := readSubscriptions(res, 25)
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

func awaitHandshake(conn net.Conn, instanceId []byte, subscriptions *[]SubcriptionInfo) (connectionWrapper, error) {
	req := make([]byte, 32)
	_, err := conn.Read(req)
	if err != nil {
		return connectionWrapper{}, err
	}

	fmt.Printf("Handshake req received from %s: %08b", conn.RemoteAddr(), req)

	conn.Write(buildHandshakeRes(instanceId, subscriptions, 0))

	return connectionWrapper{}, nil
}

func buildHandshakeRes(instanceId []byte, subscriptions *[]SubcriptionInfo, code uint8) []byte {
	res := make([]byte, 1)
	res[0] = byte(uint8(HandshakeResMessageCode))
	res = binary.BigEndian.AppendUint64(res, TolliverVersion)
	res = append(res, instanceId...)
	res = append(res, byte(code))
	res = binary.BigEndian.AppendUint32(res, uint32(len(*subscriptions)))

	for _, v := range *subscriptions {
		res = binary.BigEndian.AppendUint32(res, uint32(len(v.Channel)))
		res = binary.BigEndian.AppendUint32(res, uint32(len(v.Key)))
		res = append(res, []byte(v.Channel)...)
		res = append(res, []byte(v.Key)...)
	}

	return res
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
