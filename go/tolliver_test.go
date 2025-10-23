package tolliver

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
	"testing"
)

func TestAll(t *testing.T) {
	fmt.Println("Enter certificate name:")
	reader := bufio.NewReader(os.Stdin)
	certificatePath, _ := reader.ReadString('\n')
	certificatePath = certificatePath[:len(certificatePath)-1]

	instanceCert, err := tls.LoadX509KeyPair(certificatePath+".crt", certificatePath+".key")
	if err != nil {
		panic(err)
	}

	rootPair, _ := tls.LoadX509KeyPair("./root.crt", "./root.key")
	rootCert := rootPair.Leaf

	fmt.Println("Enter tolliver port:")
	portStr, _ := reader.ReadString('\n')
	portStr = portStr[:len(portStr)-1]
	port, pErr := strconv.ParseInt(portStr, 10, 32)

	if pErr != nil {
		panic(pErr)
	}

	// fmt.Println("Enter subscriptions ([channel,key:]*):")
	// subscriptionsString, _ := reader.ReadString('\n')
	// subscriptionsSlice := strings.Split(subscriptionsString, ":")

	cfg := InstanceOptions{
		CA:           rootCert,
		InstanceCert: &instanceCert,
		Port:         int(port),
	}

	inst, tErr := NewInstance(cfg)
	if tErr != nil {
		panic(tErr)
	}

	fmt.Println("Enter localhost port to connect to:")
	remPortStr, _ := reader.ReadString('\n')
	remPortStr = remPortStr[:len(remPortStr)-1]
	remPort, _ := strconv.ParseInt(remPortStr, 10, 32)

	errC := inst.NewConnection(ConnectionAddr{Host: "127.0.0.1", Port: int(remPort)})
	fmt.Printf("errC: %v\n", errC)
}
