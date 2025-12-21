package connections

import (
	"net"

	"github.com/google/uuid"
	"github.com/tug-dev/tolliver/go/internal/common"
)

type Wrapper struct {
	Subscriptions []common.SubcriptionInfo
	Id            uuid.UUID
	Conn          net.Conn
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
		conn, _ := lst.Accept()
		handle(conn)
	}
}
