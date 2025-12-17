package tolliver

import "net"

func sendBytes(mes []byte, conn net.Conn) {
	n, err := conn.Write(mes)
	var tot = n

	for err != nil || tot != len(mes) {
		n, err = conn.Write(mes[tot:])
		tot += n
	}
}
