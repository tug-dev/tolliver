package connections

import (
	"fmt"
	"net"
)

type Wrapper struct {
	// Subscriptions []common.SubcriptionInfo
	// Id            uuid.UUID
	Conn net.Conn
}

func SendBytes(mes []byte, conn net.Conn) {
	n, err := conn.Write(mes)
	var tot = n

	for err != nil || tot != len(mes) {
		n, err = conn.Write(mes[tot:])
		tot += n
	}
}

func HandleListener(lst net.Listener, handle func(conn net.Conn)) {
	for {
		conn, err := lst.Accept()
		if err != nil {
			fmt.Printf("%e\n", err)
			continue
		}
		handle(conn)
	}
}
